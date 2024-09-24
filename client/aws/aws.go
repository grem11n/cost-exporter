package aws

import (
	"context"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer/types"
	"github.com/grem11n/aws-cost-meter/config"
	"github.com/grem11n/aws-cost-meter/logger"
	"golang.org/x/sync/errgroup"
)

const (
	defaultPollInterval = 1 // hour, because this is the minimum time granularity
)

type Client struct {
	config *config.Config
	ce     *costexplorer.Client
}

func New(config *config.Config) (*Client, error) {
	cfg, err := awsConfig.LoadDefaultConfig(context.TODO(),
		awsConfig.WithRegion("us-east-1"), // Const Explorer is global, hence us-east-1
	)
	if err != nil {
		logger.Errorf("unable to load AWS config: %w", err)
		return nil, err
	}
	return &Client{
		ce: costexplorer.NewFromConfig(cfg),
	}, nil
}

func (c *Client) GetCostAndUsageMatrics() ([]*costexplorer.GetCostAndUsageOutput, error) {
	var results []*costexplorer.GetCostAndUsageOutput
	var mu sync.Mutex
	var grp errgroup.Group
	for _, metric := range c.config.Metrics {
		grMetric := metric
		grp.Go(func() error {
			var err error
			// We assume we are on the page 0 the first time, hence pageToken == 0
			out, err := c.getCostAndUsageMetric(grMetric, nil)
			mu.Lock()
			results = append(results, out)
			mu.Unlock()
			if out.NextPageToken != nil {
				out, err = c.getCostAndUsageMetric(grMetric, out.NextPageToken)
				mu.Lock()
				results = append(results, out)
				mu.Unlock()
			}
			return err
		})
	}
	if err := grp.Wait(); err != nil {
		return nil, err
	}
	return results, nil
}

func (c *Client) getCostAndUsageMetric(metric config.MetricsConfig, pageToken *string) (*costexplorer.GetCostAndUsageOutput, error) {
	input := c.buildCostAndUsageInput(metric, pageToken)

	out, err := c.ce.GetCostAndUsage(context.TODO(), input)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Build the input separately, since filters cannot be empty when making a query
// But they can be empty in the config
func (c *Client) buildCostAndUsageInput(metric config.MetricsConfig, pageToken *string) *costexplorer.GetCostAndUsageInput {
	interval, err := time.ParseDuration(metric.PollInterval)
	if err != nil || interval == 0 {
		interval = time.Duration(defaultPollInterval) * time.Hour
	}
	endDate := time.Now().UTC().Format(time.RFC3339)
	startDate := time.Now().Add(-interval).UTC().Format(time.RFC3339)

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
		}
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
	}
}
