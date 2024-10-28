package aws

import (
	"context"
	"crypto/md5"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/enriquebris/goconcurrentqueue"
	"github.com/grem11n/aws-cost-meter/config"
	"github.com/grem11n/aws-cost-meter/logger"
)

const (
	defaultPollInterval   = 1         // hour, because this is the minimum time granularity
	awsMetricsCachePrefix = "raw_aws" // this is important that prefix starts with raw_
	maxRetryCount         = 3         // maximum number of allowed retries to get AWS metrics before giving up
)

type Client struct {
	config *config.AWSConfig
	ce     *costexplorer.Client
	inputs *goconcurrentqueue.FixedFIFO
	mu     sync.Mutex
	cache  *sync.Map
}

type input struct {
	ceInput    *costexplorer.GetCostAndUsageInput
	readyTs    int64
	retryCount int
}

func New(config *config.AWSConfig, cache *sync.Map) (*Client, error) {
	cfg, err := awsConfig.LoadDefaultConfig(context.TODO(),
		awsConfig.WithRegion("us-east-1"), // Const Explorer is global, hence us-east-1
	)
	if err != nil {
		logger.Errorf("unable to load AWS config: %w", err)
		return nil, err
	}

	// Assume a specific role if provided
	if config.AssumeRole != "" {
		stsClient := sts.NewFromConfig(cfg)
		provider := stscreds.NewAssumeRoleProvider(stsClient, config.AssumeRole)
		cfg.Credentials = aws.NewCredentialsCache(provider)
		creds, err := cfg.Credentials.Retrieve(context.Background())
		if err != nil {
			logger.Errorf("unable to retrieve AWS credentials with AssumeRole: %w", err)
		}
		cfg.Credentials = credentials.StaticCredentialsProvider{
			Value: aws.Credentials{
				AccessKeyID:     creds.AccessKeyID,
				SecretAccessKey: creds.SecretAccessKey,
				SessionToken:    creds.SessionToken,
				Source:          "assumerole",
			},
		}
	}
	inputs := generateInitialInputs(config.Metrics)
	return &Client{
		config: config,
		ce:     costexplorer.NewFromConfig(cfg),
		inputs: inputs,
		mu:     sync.Mutex{},
		cache:  cache,
	}, nil
}

func (c *Client) GetCostAndUsageMatricsLoop() {
	for {
		c.GetCostAndUsageMatrics()
	}
}

func (c *Client) GetCostAndUsageMatrics() {
	var results []costexplorer.GetCostAndUsageOutput
	obj, err := c.inputs.DequeueOrWaitForNextElement()
	in := obj.(input) // this type cast should be safe, since we control inputs
	if err != nil {
		logger.Error(err)
		return
	}
	if in.retryCount > maxRetryCount {
		// TODO: Think about how to propagate this error further
		logger.Fatalf("Cannot get CostAndUsage metrics", err)
	}
	// Check if the metrics are up for refresh
	if in.readyTs > time.Now().Unix() {
		// Put the input back in the queue
		c.enqueuWithTs(in, in.readyTs, in.retryCount)
		return
	}

	logger.Info("making a call to AWS")
	out, err := c.ce.GetCostAndUsage(context.TODO(), in.ceInput)
	if err != nil {
		logger.Errorf("Cannot get CostAndUsage metrics", err, in.retryCount)
		// Insert 10 sec delay before retry
		readyTs := time.Now().Add(10 * time.Second).Unix()
		retry := in.retryCount + 1
		c.enqueuWithTs(in, readyTs, retry)
		return
	}

	if out == nil {
		// TODO: Should we exit here instead?
		logger.Error("CostAndUsage metrics are empty")
		readyTs := time.Now().Add(10 * time.Second).Unix()
		retry := in.retryCount + 1
		c.enqueuWithTs(in, readyTs, retry)
		return
	}

	c.mu.Lock()
	results = append(results, *out)
	c.mu.Unlock()

	if out.NextPageToken != nil {
		in.ceInput.NextPageToken = out.NextPageToken
		out, err = c.ce.GetCostAndUsage(context.TODO(), in.ceInput)
		if err != nil {
			logger.Errorf("Cannot get CostAndUsage metrics", err)
			// Insert 10 sec delay before retry
			readyTs := time.Now().Add(10 * time.Second).Unix()
			retry := in.retryCount + 1
			c.enqueuWithTs(in, readyTs, retry)
			return
		}
		c.mu.Lock()
		results = append(results, *out)
		c.mu.Unlock()
	}
	// There is no need to delay for the whole month
	readyTs := time.Now().Add(24 * time.Hour).Unix()
	// If we need hourly metrics, we need to fetch them every hour
	if strings.EqualFold(string(in.ceInput.Granularity), "hourly") {
		readyTs = time.Now().Add(1 * time.Hour).Unix()
	}
	c.enqueuWithTs(in, readyTs, 0)
	key := getMetricCacheKey(in.ceInput)
	c.cache.Swap(key, results)
}

func (c *Client) enqueuWithTs(in input, ts int64, retries int) {
	in.readyTs = ts
	in.retryCount = retries
	if err := c.inputs.Enqueue(in); err != nil {
		logger.Error(err)
	}
}

func getMetricCacheKey(ceInput *costexplorer.GetCostAndUsageInput) string {
	// Calculate a hash of the input
	suffix := md5.Sum([]byte(fmt.Sprintf("%v", ceInput)))
	return fmt.Sprintf("%s_%s", awsMetricsCachePrefix, suffix)
}

func generateInitialInputs(metrics []*config.MetricsConfig) *goconcurrentqueue.FixedFIFO {
	inputs := goconcurrentqueue.NewFixedFIFO(len(metrics))
	for _, metric := range metrics {
		el, err := buildCostAndUsageInput(metric, nil)
		if err != nil {
			logger.Errorf("Cannot build AWS CostAndUsageInput", err)
		}
		inp := input{
			ceInput:    el,
			readyTs:    time.Now().Unix(),
			retryCount: 0,
		}
		if err := inputs.Enqueue(inp); err != nil {
			logger.Error(err)
		}
	}
	return inputs
}

// Build the input separately, since filters cannot be empty when making a query
// But they can be empty in the config
func buildCostAndUsageInput(metric *config.MetricsConfig, pageToken *string) (*costexplorer.GetCostAndUsageInput, error) {
	nowUtc := time.Now().UTC()
	var endDate string
	var startDate string

	// AWS requires different time formats depending on granularity
	switch strings.ToLower(metric.Granularity) {
	case "monthly":
		interval, _ := time.ParseDuration("730h") // 1 month
		endDate = nowUtc.Format("2006-01-02")
		startDate = nowUtc.Add(-interval).Format("2006-01-02")
	case "daily":
		interval, _ := time.ParseDuration("24h") // 1 day
		endDate = nowUtc.Format("2006-01-02")
		startDate = nowUtc.Add(-interval).Format("2006-01-02")
	case "hourly":
		interval, _ := time.ParseDuration("1h") // 1 hour
		endDate = nowUtc.Format(time.RFC3339)
		startDate = nowUtc.Add(-interval).Format(time.RFC3339)
	default:
		return nil, fmt.Errorf("Unsupported granularity: %s. Supported: monthly, daily, hourly", metric.Granularity)
	}

	if reflect.ValueOf(metric.Filter).IsZero() {
		return &costexplorer.GetCostAndUsageInput{
			TimePeriod: &types.DateInterval{
				Start: aws.String(startDate),
				End:   aws.String(endDate),
			},
			Granularity:   types.Granularity(strings.ToUpper(metric.Granularity)),
			Metrics:       metric.Metrics,
			GroupBy:       metric.GroupBy,
			NextPageToken: pageToken,
		}, nil
	}
	return &costexplorer.GetCostAndUsageInput{
		TimePeriod: &types.DateInterval{
			Start: aws.String(startDate),
			End:   aws.String(endDate),
		},
		Granularity:   types.Granularity(strings.ToUpper(metric.Granularity)),
		Metrics:       metric.Metrics,
		GroupBy:       metric.GroupBy,
		Filter:        &metric.Filter,
		NextPageToken: pageToken,
	}, nil
}
