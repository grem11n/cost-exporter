package serve

import (
	"github.com/grem11n/aws-cost-meter/cache"
	"github.com/grem11n/aws-cost-meter/client/aws"
	"github.com/grem11n/aws-cost-meter/config"
	logger "github.com/grem11n/aws-cost-meter/logger"
	"github.com/grem11n/aws-cost-meter/outputs/prometheus"
)

func Run(config *config.Config) error {
	awsClient, err := aws.New(config.AWS)
	if err != nil {
		logger.Errorf("Unable to create AWS client: %w", err)
		return err
	}

	if err = awsClient.GetCostAndUsageMatrics(); err != nil {
		logger.Errorf("Unable to get cost and usage metrics: %w", err)
		return err
	}

	// Get the raw metrics cache
	rawMetrics := cache.GetRawCache()
	if err := prometheus.ConvertRawMetrics(rawMetrics); err != nil {
		return err
	}

	// Execute a given output.
	//fmt.Printf("Config: %+v\n", config)
	return nil
}
