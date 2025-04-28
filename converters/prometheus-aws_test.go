package converters

import (
	"sync"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/costexplorer"
	"github.com/grem11n/cost-exporter/clients"
	"github.com/stretchr/testify/assert"
)

var (
	testPrometheus     = PrometheusAWS{}
	expectedMetricsMap = map[string]map[string]float64{
		"NetAmortizedCost": {
			"AWS Cost Explorer":             0.19,
			"Amazon DynamoDB":               0,
			"Amazon Simple Storage Service": 5,
			"AmazonCloudWatch":              0.31,
		},
		"NetUnblendedCost": {
			"AWS Cost Explorer":             0.19,
			"Amazon DynamoDB":               0,
			"Amazon Simple Storage Service": 5,
			"AmazonCloudWatch":              0.35,
		},
	}
	expectedPrometheusMetrics = `aws_ce_net_amortized_cost{job="cost-exporter",dimension="AWS Cost Explorer"} 0.19
aws_ce_net_amortized_cost{job="cost-exporter",dimension="Amazon DynamoDB"} 0
aws_ce_net_amortized_cost{job="cost-exporter",dimension="Amazon Simple Storage Service"} 5
aws_ce_net_amortized_cost{job="cost-exporter",dimension="AmazonCloudWatch"} 0.31
aws_ce_net_unblended_cost{job="cost-exporter",dimension="AWS Cost Explorer"} 0.19
aws_ce_net_unblended_cost{job="cost-exporter",dimension="Amazon DynamoDB"} 0
aws_ce_net_unblended_cost{job="cost-exporter",dimension="Amazon Simple Storage Service"} 5
aws_ce_net_unblended_cost{job="cost-exporter",dimension="AmazonCloudWatch"} 0.35
`
)

func TestDiscoverMetrics(t *testing.T) {
	var m []costexplorer.GetCostAndUsageOutput
	for _, c := range clients.CeStub {
		m = append(m, *c)
	}
	got, err := testPrometheus.discoverAWSMetrics(m)
	assert.NoError(t, err)
	assert.Equal(t, expectedMetricsMap, got)
}

func TestConvertAWSMetricsPositive(t *testing.T) {
	var m []costexplorer.GetCostAndUsageOutput
	for _, c := range clients.CeStub {
		m = append(m, *c)
	}
	var mp sync.Map
	mp.Store("aws_0", m)
	testPrometheus.convertAWSMetrics(&mp)
	got, ok := mp.Load(namespace)
	assert.True(t, ok, "Cannot load test results from cache")
	gotb, ok := got.([]byte)
	assert.True(t, ok, "Cannot load test results from cache")
	assert.Equal(t, expectedPrometheusMetrics, string(gotb))
}
