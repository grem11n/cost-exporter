package serve

import (
	"fmt"

	"github.com/grem11n/aws-cost-meter/client/aws"
	"github.com/grem11n/aws-cost-meter/config"
	logger "github.com/grem11n/aws-cost-meter/logger"
)

func Run(config *config.Config) error {
	awsClient, err := aws.New(config)
	if err != nil {
		logger.Errorf("Unable to create AWS client: %w", err)
		return err
	}
	metrics, err := awsClient.GetCostAndUsageMatrics()
	if err != nil {
		logger.Errorf("Unable to get cost and usage metrics: %w", err)
		return err
	}
	for _, metric := range metrics.CostAndUsageMetrics {
		fmt.Printf("Metrics: %+v\n", metric)
		for k, v := range metric.ResultsByTime[0].Total {
			fmt.Printf("%s: %v %v\n", k, *v.Amount, *v.Unit)
		}
	}
	return nil
}
