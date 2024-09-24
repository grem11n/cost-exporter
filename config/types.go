package config

import (
	"github.com/aws/aws-sdk-go-v2/service/costexplorer/types"
)

type Config struct {
	AWS        AWSConfig          `mapstructure:"aws"`
	Metrics    []MetricsConfig    `mapstructure:"metrics"`
	Converters []ConvertersConfig `mapstructure:"converters"`
	Outputs    []OutputsConfig    `mapstructure:"outputs"`
}

type MetricsConfig struct {
	Name         string                  `mapstructure:"name"`
	PollInterval string                  `mapstructure:"poll_interval"`
	Granularity  string                  `mapstructure:"granularity"`
	Metrics      []string                `mapstructure:"metrics"`
	GroupBy      []types.GroupDefinition `mapstructure:"group_by"`
	Filter       types.Expression        `mapstructure:"filter"`
}

type AWSConfig struct{}
type ConvertersConfig struct{}
type OutputsConfig struct{

}
