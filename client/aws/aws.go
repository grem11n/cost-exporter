package aws

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer/types"
	"github.com/enriquebris/goconcurrentqueue"
	"github.com/grem11n/aws-cost-meter/config"
	"github.com/grem11n/aws-cost-meter/logger"
)

const (
	defaultPollInterval = 1 // hour, because this is the minimum time granularity
)

type Client struct {
	config *config.AWSConfig
	ce     *costexplorer.Client
	inputs *goconcurrentqueue.FixedFIFO
	mu     sync.Mutex
}

type input struct {
	ceInput *costexplorer.GetCostAndUsageInput
	delayTs int64
}

func New(config *config.AWSConfig) (*Client, error) {
	cfg, err := awsConfig.LoadDefaultConfig(context.TODO(),
		awsConfig.WithRegion("us-east-1"), // Const Explorer is global, hence us-east-1
	)
	if err != nil {
		logger.Errorf("unable to load AWS config: %w", err)
		return nil, err
	}
	inputs := generateInitialInputs(config.Metrics)
	return &Client{
		config: config,
		ce:     costexplorer.NewFromConfig(cfg),
		inputs: inputs,
		mu:     sync.Mutex{},
	}, nil
}

func generateInitialInputs(metrics []config.MetricsConfig) *goconcurrentqueue.FixedFIFO {
	inputs := goconcurrentqueue.NewFixedFIFO(len(metrics))
	for _, metric := range metrics {
		el, err := buildCostAndUsageInput(metric, nil)
		if err != nil {
			logger.Errorf("Cannot build AWS CostAndUsageInput", err)
		}
		inp := input{ceInput: el, delayTs: time.Now().Unix()}
		inputs.Enqueue(inp)
	}
	return inputs
}

// GetCostAndUsageMatrics gets CostAndUsage information from AWS in the background
func (c *Client) GetCostAndUsageMatrics(cache *sync.Map) {
	for {
		c.getCostAndUsageMatrics(cache)
	}
}

func (c *Client) getCostAndUsageMatrics(cache *sync.Map) {
	//TODO: remove after tests
	// Do not make requests to AWS, because  they are costly.
	// Return a stub instead

	var results []costexplorer.GetCostAndUsageOutput
	obj, err := c.inputs.DequeueOrWaitForNextElement()
	in := obj.(input) // this type cast should be safe, since we control inputs
	if err != nil {
		logger.Error(err)
	}
	// Check if the metrics are up for refresh
	if in.delayTs > time.Now().Unix() {
		return
	}
	// TODO: Debug to not contact AWS
	out := CeStub[1]

	//var err error
	//out, err := c.ce.GetCostAndUsage(context.TODO(), in.ceInput)
	//if err != nil {
	//	logger.Errorf("Cannot get CostAndUsage metrics", err)
	//	// Insert 5 sec delay before retry
	//	c.insertDelay(i, time.Now().Add(5*time.Second).Unix())
	//}
	//if out == nil {
	//	logger.Error("CostAndUsage metrics are empty")
	//	continue
	//}
	c.mu.Lock()
	results = append(results, *out)
	c.mu.Unlock()
	//if out.NextPageToken != nil {
	//	in.ceInput.NextPageToken = out.NextPageToken
	//	out, err = c.ce.GetCostAndUsage(context.TODO(), in.ceInput)
	//	if err != nil {
	//		logger.Errorf("Cannot get CostAndUsage metrics", err)
	//		// Insert 5 sec delay before retry
	//		c.insertDelay(i, time.Now().Add(5*time.Second).Unix())
	//	}
	//	c.mu.Lock()
	//	results = append(results, out)
	//	c.mu.Unlock()
	//}
	// There is no need to delay for the whole month
	delayTs := time.Now().Add(24 * time.Hour).Unix()
	// If we need hourly metrics, we need to fetch them every hour
	if strings.EqualFold(string(in.ceInput.Granularity), "hourly") {
		delayTs = time.Now().Add(1 * time.Hour).Unix()
	}
	in.delayTs = delayTs
	err = c.inputs.Enqueue(in)
	if err != nil {
		logger.Error(err)
	}
	cache.Store("aws", results)
}

func (c *Client) getCostAndUsageMetric(metric config.MetricsConfig, pageToken *string) (*costexplorer.GetCostAndUsageOutput, error) {
	input, err := buildCostAndUsageInput(metric, pageToken)
	if err != nil {
		return nil, err
	}

	out, err := c.ce.GetCostAndUsage(context.TODO(), input)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Build the input separately, since filters cannot be empty when making a query
// But they can be empty in the config
func buildCostAndUsageInput(metric config.MetricsConfig, pageToken *string) (*costexplorer.GetCostAndUsageInput, error) {
	nowUtc := time.Now().UTC()
	var endDate string
	var startDate string

	// AWS requires different time formats depending on granularity
	switch strings.ToLower(metric.Granularity) {
	case "monthly":
		interval, _ := time.ParseDuration("731h") // 1 month
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
