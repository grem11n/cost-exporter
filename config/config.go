package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

// By default use ./config.yaml file
const (
	defaultConfigPath = "./config.yaml"
	// prefix for validation failed error messages
	vfp = "config validation failed:"
)

var (
	validAWSMetrics = map[string]bool{
		"UnblendedCost":    true,
		"BlendedCost":      true,
		"NetUnblendedCost": true,
		"NetAmortizedCost": true,
		"UsageQuantity":    true,
	}
	validMetricGranularity = map[string]bool{
		"daily":   true,
		"monthly": true,
		"hourly":  true,
	}
)

func New(configPath string) (*Config, error) {
	if configPath == "" {
		configPath = os.Getenv("CONFIG")
		if configPath == "" {
			configPath = defaultConfigPath
		}
	}

	viper.SetConfigFile(configPath)
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("unable to read the config file %s: %w", configPath, err)
	}
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("Unable to read the config file %s: %w", configPath, err)
	}

	if err := validate(&config); err != nil {
		return nil, fmt.Errorf("Unable to read the config file %s: %w", configPath, err)
	}

	return &config, nil
}

// Move these validations as well as types to the AWS package
func validate(config *Config) error {
	if err := validateEmptyMetrics(config); err != nil {
		return err
	}
	for _, metric := range config.AWS.Metrics {
		// Validate metric type
		if err := validateMetricType(&metric); err != nil {
			return err
		}
		// Validate granularity
		if err := validateGranularity(&metric); err != nil {
			return err
		}
	}
	return nil
}

func validateEmptyMetrics(config *Config) error {
	if len(config.AWS.Metrics) == 0 {
		return fmt.Errorf("No metrics found in config")
	}
	return nil
}

func validateMetricType(metric *MetricsConfig) error {
	for _, mtype := range metric.Metrics {
		if _, ok := validAWSMetrics[mtype]; !ok {
			supportedMetrics := []string{}
			for k := range validAWSMetrics {
				supportedMetrics = append(supportedMetrics, k)
			}
			supportedMetricsStr := strings.Join(supportedMetrics, ", ")
			return fmt.Errorf("%s unsupported metric type. Supported metrics are: %s. Got: %s",
				vfp,
				supportedMetricsStr,
				mtype,
			)
		}
	}
	return nil
}

func validateGranularity(metric *MetricsConfig) error {
	gran := strings.ToLower(metric.Granularity)
	if !validMetricGranularity[gran] {
		return fmt.Errorf("%s unsdupported granularity. Supported types: MONTHLY, DAILY, HOURLY. Got: %s", vfp, gran)
	}
	return nil
}
