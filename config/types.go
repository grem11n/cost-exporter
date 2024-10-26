package config

import (
	"github.com/aws/aws-sdk-go-v2/service/costexplorer/types"
)

// Config stores configuration required to run.
// This is a direct representation of  the config file.
type Config struct {
	AWS     *AWSConfig       `mapstructure:"aws"`
	Outputs []*OutputsConfig `mapstructure:"outputs"`
}

// AWSConfig stores information related to AWS:
//   - Auth information
//   - A list of Cost Explorer metrics (see `example.config.yaml` for an example)
type AWSConfig struct {
	Metrics []*MetricsConfig `mapstructure:"metrics"`
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
