package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/grem11n/cost-exporter/clients"
	"github.com/grem11n/cost-exporter/outputs"
	"github.com/grem11n/cost-exporter/probes"
	"github.com/spf13/viper"
)

const (
	defaultConfigPath = "./config.yaml"
)

type Config struct {
	Clients       map[string]clients.ClientConfig `mapstructure:"clients"`
	MetricsFormat string                          `mapstructure:"metrics_format,omitempty"`
	Outputs       map[string]outputs.OutputConfig `mapstructure:"outputs"`
	Probes        probes.ProbeConfig              `mapstructure:"kubernetes_probes,omitempty"`
}

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

	if err := config.populateDefaults(); err != nil {
		return nil, err
	}

	return &config, nil
}

func (c *Config) populateDefaults() error {
	if c.Clients == nil {
		return errors.New("client configuration is required. Only AWS is supported")
	}

	if c.Outputs == nil {
		c.Outputs = make(map[string]outputs.OutputConfig)
		c.Outputs["http"] = outputs.Http{}
	}
	return nil
}
