package converters

var (
	testPrometheus     = Prometheus{}
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
