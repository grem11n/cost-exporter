package config

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer/types"

	"github.com/stretchr/testify/assert"
)

var goodConfig = &Config{
	AWS: &AWSConfig{
		Metrics: []*MetricsConfig{
			{
				Granularity: "monthly",
				Metrics:     []string{"NetAmortizedCost", "NetUnblendedCost"},
				GroupBy: []types.GroupDefinition{
					{
						Key:  aws.String("SERVICE"),
						Type: "DIMENSION",
					},
				},
				Filter: types.Expression{},
			},
		},
	},
	Outputs: []*OutputsConfig{
		{
			Type:      "http",
			Converter: "prometheus",
		},
	},
}

// Not sure how well it would keep up with example changes
func TestNew(t *testing.T) {
	// Assuming the default config in the repo
	conf, err := New("../exmaple.config.yaml") // hardcoded
	assert.NoError(t, err)
	assert.Equal(t, goodConfig, conf)
}

func TestEmptyConfig(t *testing.T) {
	emptyCfg := &Config{}
	err := validate(emptyCfg)
	assert.ErrorContains(t, err, "config is empty")
}

func TestValidateEmptyMetricsNegative(t *testing.T) {
	emptyCfg := &Config{AWS: &AWSConfig{}}
	err := validateEmptyMetrics(emptyCfg)
	assert.ErrorContains(t, err, "No metrics found")
}

func TestValidateMetricType(t *testing.T) {
	badCfg := &MetricsConfig{
		Metrics: []string{"BigMacPrice"},
	}
	err := validateMetricType(badCfg)
	assert.ErrorContains(t, err, "unsupported metric type")
}

func TestGranularity(t *testing.T) {
	badCfg := &MetricsConfig{
		Granularity: "täglish",
	}
	err := validateGranularity(badCfg)
	assert.ErrorContains(t, err, "unsupported granularity")
}
