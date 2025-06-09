package converters

import (
	"sync"
	"time"

	"github.com/grem11n/cost-exporter/logger"
	"github.com/mitchellh/mapstructure"
)

type CloudWatch struct {
	Namespace string
}

func init() {
	logger.Info("Initializing PrometheusAWS converter")
	Register(namespace, func(conf ConverterConfig) Converter {
		var cw CloudWatch
		if err := mapstructure.Decode(conf, &cw); err != nil {
			logger.Fatalf("unable to decode CloudWatch config: %w", err)
		}
		logger.Debug("CloudWatch config: ", cw)
		return &cw
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
	return true
}
