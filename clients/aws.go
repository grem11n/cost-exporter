package clients

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/VictoriaMetrics/metrics"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/enriquebris/goconcurrentqueue"
	intmetrics "github.com/grem11n/cost-exporter/internal/metrics"
	"github.com/grem11n/cost-exporter/logger"
	"github.com/mitchellh/mapstructure"
)

const (
	maxRetryCount          = 3 // maximum number of allowed retries to get AWS metrics before giving up
	keyPrefix              = "aws"
	awsCallsSuccessName    = "cost_exporter_aws_calls_total{job=\"cost-exporter\",result=\"success\"}"
	awsCallsFailureName    = "cost_exporter_aws_calls_total{job=\"cost-exporter\",result=\"failure\"}"
	getMetricsDurationName = "cost_exporter_aws_get_metrics_duration{job=\"cost-exporter\"}"
)

var (
	awsCallsSuccess    *metrics.Counter
	awsCallsFailure    *metrics.Counter
	getMetricsDuration *metrics.Histogram
)

type AWS struct {
	AssumeRole string           `mapstructure:"assume_role"`
	Metrics    []*MetricsConfig `mapstructure:"metrics"`
	ce         *costexplorer.Client
	inputs     *goconcurrentqueue.FixedFIFO
	mu         sync.Mutex
}

type AWSConfig struct {
	AssumeRole string           `mapstructure:"role,omitempty"`
	Metrics    []*MetricsConfig `mapstructure:"metrics"`
}

// MetricsConfig maps to the `costexplorer.GetCostAndUsageInput` type.
// For more information about each field, see:
// https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/costexplorer#GetCostAndUsageInput
type MetricsConfig struct {
	Granularity string                  `mapstructure:"granularity"`
	Metrics     []string                `mapstructure:"metrics"`
	GroupBy     []types.GroupDefinition `mapstructure:"group_by"`
	Filter      types.Expression        `mapstructure:"filter"`
}

type input struct {
	index      int
	ceInput    *costexplorer.GetCostAndUsageInput
	readyTs    int64
	retryCount int
}

func init() {
	logger.Info("Initializing AWS client")
	Register("aws", func(conf ClientConfig) Client {
		var cfg AWSConfig
		if err := mapstructure.Decode(conf, &cfg); err != nil {
			logger.Fatalf("unable to decode AWS config: %w", err)
		}
		logger.Debug("AWS config: ", cfg)
		ceCfg, err := config.LoadDefaultConfig(context.TODO(),
			config.WithRegion("us-east-1"), // Const Explorer is global, hence us-east-1
		)
		if err != nil {
			logger.Fatalf("unable to load AWS config: %w", err)
		}
		// Assume a specific role if provided
		if cfg.AssumeRole != "" {
			stsClient := sts.NewFromConfig(ceCfg)
			provider := stscreds.NewAssumeRoleProvider(stsClient, cfg.AssumeRole)
			ceCfg.Credentials = aws.NewCredentialsCache(provider)
			creds, err := ceCfg.Credentials.Retrieve(context.Background())
			if err != nil {
				logger.Errorf("unable to retrieve AWS credentials with AssumeRole: %w", err)
			}
			ceCfg.Credentials = credentials.StaticCredentialsProvider{
				Value: aws.Credentials{
					AccessKeyID:     creds.AccessKeyID,
					SecretAccessKey: creds.SecretAccessKey,
					SessionToken:    creds.SessionToken,
					Source:          "assumerole",
				},
			}
		}
		inputs := generateInitialInputs(cfg.Metrics)
		return &AWS{
			ce:     costexplorer.NewFromConfig(ceCfg),
			inputs: inputs,
			mu:     sync.Mutex{},
		}
	})
	// Maybe initiate all the metrics in a loop if there are too many
	logger.Info("Initializing AWS Client metrics")
	awsCallsSuccess = intmetrics.InternalMetricsSet.GetOrCreateCounter(awsCallsSuccessName)
	awsCallsFailure = intmetrics.InternalMetricsSet.GetOrCreateCounter(awsCallsFailureName)
	getMetricsDuration = intmetrics.InternalMetricsSet.GetOrCreateHistogram(getMetricsDurationName)
}

func (a *AWS) GetMetrics(cache *sync.Map) {
	for {
		a.getCostAndUsageMetrics(cache)
	}
}

func (a *AWS) getCostAndUsageMetrics(cache *sync.Map) {
	startTs := time.Now()
	var results []costexplorer.GetCostAndUsageOutput
	obj, err := a.inputs.DequeueOrWaitForNextElement()
	in := obj.(input) // this type cast should be safe, since we control inputs
	if err != nil {
		logger.Error(err)
		return
	}
	if in.retryCount > maxRetryCount {
		logger.Fatalf("Cannot get CostAndUsage metrics", err)
	}
	// Check if the metrics are up for refresh
	if in.readyTs > time.Now().Unix() {
		// Put the input back in the queue
		a.enqueuWithTs(in, in.readyTs, in.retryCount)
		return
	}

	logger.Info("Making a call to AWS")
	out := a.costAndUsageCall(in)
	if out == nil {
		// We have already enqueued a retry
		return
	}

	a.mu.Lock()
	results = append(results, *out)
	a.mu.Unlock()

	for out.NextPageToken != nil {
		in.ceInput.NextPageToken = out.NextPageToken
		out = a.costAndUsageCall(in)
		a.mu.Lock()
		results = append(results, *out)
		a.mu.Unlock()
	}
	// There is no need to delay for the whole month
	readyTs := time.Now().Add(24 * time.Hour).Unix()
	// If we need hourly metrics, we need to fetch them every hour
	if strings.EqualFold(string(in.ceInput.Granularity), "hourly") {
		readyTs = time.Now().Add(1 * time.Hour).Unix()
	}
	a.enqueuWithTs(in, readyTs, 0)
	key := fmt.Sprintf("%s_%d", keyPrefix, in.index)
	logger.Debugf("Adding AWS metrics to the cache. Key: %s", key)
	cache.Swap(key, results)
	logger.Debug("Metrics: ", results)
	getMetricsDuration.UpdateDuration(startTs)
}

func (a *AWS) costAndUsageCall(in input) *costexplorer.GetCostAndUsageOutput {
	out, err := a.ce.GetCostAndUsage(context.TODO(), in.ceInput)
	if err != nil {
		logger.Error("Cannot get CostAndUsage metrics", err, in.retryCount)
		awsCallsFailure.Inc()
		// Insert 10 sec delay before retry
		readyTs := time.Now().Add(10 * time.Second).Unix()
		retry := in.retryCount + 1
		a.enqueuWithTs(in, readyTs, retry)
		return nil
	}

	if out == nil {
		// TODO: Should we exit here instead?
		logger.Error("CostAndUsage metrics are empty")
		awsCallsFailure.Inc()
		readyTs := time.Now().Add(10 * time.Second).Unix()
		retry := in.retryCount + 1
		a.enqueuWithTs(in, readyTs, retry)
		return nil
	}
	awsCallsSuccess.Inc()
	return out
}

func generateInitialInputs(metrics []*MetricsConfig) *goconcurrentqueue.FixedFIFO {
	inputs := goconcurrentqueue.NewFixedFIFO(len(metrics))
	for i, metric := range metrics {
		el, err := buildCostAndUsageInput(metric, nil)
		if err != nil {
			logger.Errorf("Cannot build AWS CostAndUsageInput", err)
		}
		inp := input{
			index:      i,
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
func buildCostAndUsageInput(metric *MetricsConfig, pageToken *string) (*costexplorer.GetCostAndUsageInput, error) {
	nowUtc := time.Now().UTC()
	var endDate string
	var startDate string

	// AWS requires different time formats depending on granularity
	switch strings.ToLower(metric.Granularity) {
	case "monthly":
		interval, _ := time.ParseDuration("730h") //nolint:errcheck
		endDate = nowUtc.Format("2006-01-02")
		startDate = nowUtc.Add(-interval).Format("2006-01-02")
	case "daily":
		interval, _ := time.ParseDuration("24h") //nolint:errcheck
		endDate = nowUtc.Format("2006-01-02")
		startDate = nowUtc.Add(-interval).Format("2006-01-02")
	case "hourly":
		interval, _ := time.ParseDuration("1h") //nolint:errcheck
		endDate = nowUtc.Format(time.RFC3339)
		startDate = nowUtc.Add(-interval).Format(time.RFC3339)
	default:
		return nil, fmt.Errorf("unsupported granularity: %s. Supported: monthly, daily, hourly", metric.Granularity)
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

func (a *AWS) enqueuWithTs(in input, ts int64, retries int) {
	in.readyTs = ts
	in.retryCount = retries
	if err := a.inputs.Enqueue(in); err != nil {
		logger.Error(err)
	}
}
