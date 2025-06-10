package converters

import (
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	intmetrics "github.com/grem11n/cost-exporter/internal/metrics"
	"github.com/grem11n/cost-exporter/logger"
)

// Namespace is used as a prefix in cache
type CloudWatch struct {
	Namespace string
}

func init() {
	logger.Info("Initializing CloudWatch converter")
	name := "cloudwatch"
	Register(name, func(conf ConverterConfig) Converter {
		return &CloudWatch{
			Namespace: name,
		}
	})
}

func (c *CloudWatch) Convert(cache *sync.Map, fetchPrefix string) {
	logger.Info("Converting cost metrics to the CloudWatch format")
	for {
		if ok := c.convert(cache, fetchPrefix); ok {
			time.Sleep(cooldown * time.Minute)
		}
	}
}

func (c *CloudWatch) convert(cache *sync.Map, fetchPrefix string) bool {
	var metrics []types.MetricDatum
	cache.Range(func(key, value any) bool {
		ks, ok := key.(string)
		if !ok {
			logger.Error("cache key is not a string, got %T", key)
			// Return on this iteration, but continue the loop
			return true
		}
		if strings.HasPrefix(ks, fetchPrefix) {
			metric, ok := value.(intmetrics.Metric)
			if !ok {
				logger.Warn("wrong metric type found in %s: %T", fetchPrefix, value)
			}
			// weird AWS name for tags
			var dimensions []types.Dimension
			for k, v := range metric.Tags {
				dimensions = append(dimensions, types.Dimension{
					Name:  aws.String(k),
					Value: aws.String(v),
				})
			}
			metrics = append(metrics, types.MetricDatum{
				MetricName: aws.String(metric.Name),
				Unit:       types.StandardUnitCount,
				Value:      aws.Float64(metric.Value),
				Dimensions: dimensions,
			})
		}
		return true
	})

	// Handle the case when the metrics are not yet present
	if len(metrics) == 0 {
		return false
	}

	logger.Debug("Writing CloudWatch metrics to cache with key: ", c.Namespace)
	cache.Swap(c.Namespace, metrics)
	logger.Debug("CloudWatch metrics: ", metrics)
	return true
}
