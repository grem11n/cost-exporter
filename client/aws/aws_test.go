package aws

import (
	"time"

	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer/types"
	"github.com/grem11n/aws-cost-meter/config"
	"github.com/stretchr/testify/assert"
)

var (
	testCostAndUsageInput = costexplorer.GetCostAndUsageInput{
		TimePeriod:  &types.DateInterval{},
		Granularity: types.Granularity("DAILY"),
		Metrics:     []string{"UnblendedCost"},
		GroupBy: []types.GroupDefinition{
			{
				Key:  aws.String("SERVICE"),
				Type: "DIMENSION",
			},
		},
		NextPageToken: nil,
	}
	testMetric = config.MetricsConfig{
		Granularity: "daily",
		Metrics:     []string{"UnblendedCost"},
		GroupBy: []types.GroupDefinition{
			{
				Key:  aws.String("SERVICE"),
				Type: "DIMENSION",
			},
		},
	}
)

func TestBuildCostAndUsageInputNoFilter(t *testing.T) {
	interval, _ := time.ParseDuration("24h") // 1 day
	endDate := time.Now().UTC().Format("2006-01-02")
	startDate := time.Now().UTC().Add(-interval).Format("2006-01-02")
	expected := testCostAndUsageInput
	expected.TimePeriod = &types.DateInterval{
		End:   aws.String(endDate),
		Start: aws.String(startDate),
	}

	got, err := buildCostAndUsageInput(&testMetric, nil)
	assert.NoError(t, err)
	assert.Equal(t, &expected, got)
}

func TestBuildCostAndUsageInputFilter(t *testing.T) {
	tm := testMetric
	tm.Filter = types.Expression{
		Dimensions: &types.DimensionValues{
			Key:    "SERVICE",
			Values: []string{"AWS Cost Explorer"},
		},
	}
	interval, _ := time.ParseDuration("24h") // 1 day
	endDate := time.Now().UTC().Format("2006-01-02")
	startDate := time.Now().UTC().Add(-interval).Format("2006-01-02")
	expected := testCostAndUsageInput
	expected.TimePeriod = &types.DateInterval{
		End:   aws.String(endDate),
		Start: aws.String(startDate),
	}
	expected.Filter = &types.Expression{
		Dimensions: &types.DimensionValues{
			Key:    "SERVICE",
			Values: []string{"AWS Cost Explorer"},
		},
	}

	got, err := buildCostAndUsageInput(&tm, nil)
	assert.NoError(t, err)
	assert.Equal(t, &expected, got)
}

func TestGenerateInitialInputs(t *testing.T) {
	metrics := []*config.MetricsConfig{&testMetric}
	inputs := generateInitialInputs(metrics)
	assert.Equal(t, 1, inputs.GetCap())
	assert.Equal(t, 1, inputs.GetLen())
}
