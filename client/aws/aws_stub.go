package aws

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer/types"
	"github.com/aws/smithy-go/middleware"
)

var CeStub []*costexplorer.GetCostAndUsageOutput = []*costexplorer.GetCostAndUsageOutput{
	{
		DimensionValueAttributes: []types.DimensionValuesWithAttributes{},
		GroupDefinitions: []types.GroupDefinition{
			{
				Key:  aws.String("SERVICE"),
				Type: types.GroupDefinitionTypeDimension,
			},
		},
		NextPageToken: nil,
		ResultsByTime: []types.ResultByTime{
			{
				Estimated: true,
				Groups: []types.Group{
					{
						Keys: []string{"AWS Cost Explorer"},
						Metrics: map[string]types.MetricValue{
							"NetAmortizedCost": {
								Amount: aws.String("0.19"),
								Unit:   aws.String("USD"),
							},
							"NetUnblendedCost": {
								Amount: aws.String("0.19"),
								Unit:   aws.String("USD"),
							},
						},
					},
				},
				TimePeriod: &types.DateInterval{
					Start: aws.String("2024-10-01"),
					End:   aws.String("2024-10-02"),
				},
				Total: map[string]types.MetricValue{},
			},
		},
		ResultMetadata: middleware.Metadata{},
	},
	{
		DimensionValueAttributes: []types.DimensionValuesWithAttributes{},
		GroupDefinitions: []types.GroupDefinition{
			{
				Key:  aws.String("SERVICE"),
				Type: types.GroupDefinitionTypeDimension,
			},
		},
		NextPageToken: nil,
		ResultsByTime: []types.ResultByTime{
			{
				Estimated: true,
				Groups: []types.Group{
					{
						Keys: []string{"AWS Cost Explorer"},
						Metrics: map[string]types.MetricValue{
							"NetAmortizedCost": {
								Amount: aws.String("0.19"),
								Unit:   aws.String("USD"),
							},
							"NetUnblendedCost": {
								Amount: aws.String("0.19"),
								Unit:   aws.String("USD"),
							},
						},
					},
					{
						Keys: []string{"Amazon DynamoDB"},
						Metrics: map[string]types.MetricValue{
							"NetAmortizedCost": {
								Amount: aws.String("0"),
								Unit:   aws.String("USD"),
							},
							"NetUnblendedCost": {
								Amount: aws.String("0"),
								Unit:   aws.String("USD"),
							},
						},
					},
					{
						Keys: []string{"Amazon Simple Storage Service"},
						Metrics: map[string]types.MetricValue{
							"NetAmortizedCost": {
								Amount: aws.String("5"),
								Unit:   aws.String("USD"),
							},
							"NetUnblendedCost": {
								Amount: aws.String("5"),
								Unit:   aws.String("USD"),
							},
						},
					},
					{
						Keys: []string{"AmazonCloudWatch"},
						Metrics: map[string]types.MetricValue{
							"NetAmortizedCost": {
								Amount: aws.String("0.31"),
								Unit:   aws.String("USD"),
							},
							"NetUnblendedCost": {
								Amount: aws.String("0.35"),
								Unit:   aws.String("USD"),
							},
						},
					},
				},
				TimePeriod: &types.DateInterval{
					Start: aws.String("2024-10-01"),
					End:   aws.String("2024-10-02"),
				},
				Total: map[string]types.MetricValue{},
			},
		},
		ResultMetadata: middleware.Metadata{},
	},
}
