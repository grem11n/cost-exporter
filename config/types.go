package config

import (
	"github.com/aws/aws-sdk-go-v2/service/costexplorer/types"
)

type Config struct {
	AWS     AWSConfig       `mapstructure:"aws"`
	Metrics []MetricsConfig `mapstructure:"metrics"`
	Outputs []OutputsConfig `mapstructure:"outputs"`
}

// TODO: Calculate metrics names based on metrics themselves
type MetricsConfig struct {
	Name        string                  `mapstructure:"name"`
	Granularity string                  `mapstructure:"granularity"`
	Metrics     []string                `mapstructure:"metrics"`
	GroupBy     []types.GroupDefinition `mapstructure:"group_by"`
	Filter      types.Expression        `mapstructure:"filter"`
}

type AWSConfig struct{}

// OutputsConfig holds configuration specific for each output implementation.
type OutputsConfig struct {
	Prometheus PrometheusConfig `mapstructure:"prometheus"`
}

// PrometheusConfig holds configuration for HTTP output in Prometheus format.
type PrometheusConfig struct {
	Host string
	Port int64
}
