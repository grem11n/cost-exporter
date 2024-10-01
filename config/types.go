package config

import (
	"github.com/aws/aws-sdk-go-v2/service/costexplorer/types"
)

type Config struct {
	AWS     *AWSConfig      `mapstructure:"aws"`
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

type AWSConfig struct {
	Metrics []MetricsConfig `mapstructure:"metrics"`
}

// Enum values for output types
type OutputType string

const (
	OutputTypeHTTP   OutputType = "http"
	OutputTypeStdOut OutputType = "stdout"
)

// Enum type for available converters
type OutputConverter string

const (
	OutputConverterPrometheus OutputConverter = "prometheus"
)

// OutputsConfig holds configuration specific for each output implementation.
type OutputsConfig struct {
	Type      OutputType
	Converter OutputConverter
}

// PrometheusConfig holds configuration for HTTP output in Prometheus format.
type PrometheusConfig struct {
	Host string `mapstructure:"host"`
	Port int64  `mapstructure:"port"`
}
