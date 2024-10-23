package prometheus

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/VictoriaMetrics/metrics"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer"
	"github.com/grem11n/aws-cost-meter/logger"
	"github.com/iancoleman/strcase"
)

// TODO: Move to config
const (
	awsCachePrefix          = "aws"
	awsCacheProcessedPrefix = "aws_prometheus"
	defaultMetricPrefix     = "ce_exporter"
	defaultCacheTTL         = 1 * time.Hour // the minimal granulatiry of AWS cost metrics
)

func ConvertAWSMetrics(cache *sync.Map) bytes.Buffer {
	var awsMetrics []costexplorer.GetCostAndUsageOutput
	cache.Range(func(key, value any) bool {
		if strings.HasPrefix(key.(string), "raw_aws") { // hardcode
			res, ok := value.([]costexplorer.GetCostAndUsageOutput)
			if !ok {
				logger.Warn("cache doesn't have %s metrics", awsCachePrefix)
			}
			awsMetrics = append(awsMetrics, res...)
		}
		return true
	})

	metricNameMap, err := discoverAWSMetrics(awsMetrics)
	if err != nil {
		logger.Error(err)
	}

	vm := metrics.NewSet()
	for name, groups := range metricNameMap {
		for group, amount := range groups {
			metricName := fmt.Sprintf("%s_%s{job=\"%s\",dimension=\"%s\"}", defaultMetricPrefix, strcase.ToSnake(name), "ce-exporter", group)
			vm.GetOrCreateGauge(metricName, func() float64 {
				return amount
			})
		}
	}
	var res bytes.Buffer
	vm.WritePrometheus(&res)
	return res
}

// Analyze the raw metrics structure to discover, which metrics are present
func discoverAWSMetrics(metrics []costexplorer.GetCostAndUsageOutput) (map[string]map[string]float64, error) {
	// Using maps to convert the raw format and deduplicate metrics "in flight"
	var metricNameMap = make(map[string]map[string]float64)

	for _, metrics := range metrics {
		// There is only a single element in the .ResultsByTime, because of how we craft the time period in the initial request
		if len(metrics.ResultsByTime) == 0 {
			return nil, errors.New("No metrics were found!")
		}
		for _, group := range metrics.ResultsByTime[0].Groups {
			for costType := range group.Metrics {
				groupName := group.Keys[0] // there is just a single key
				amount, err := strconv.ParseFloat(*group.Metrics[costType].Amount, 64)
				if err != nil {
					logger.Errorf("cannot parse metric amount: ", err)
					return nil, err
				}
				prev := metricNameMap[costType]
				if prev == nil {
					prev = make(map[string]float64)
				}
				prev[groupName] = amount
				metricNameMap[costType] = prev
			}
		}
	}
	return metricNameMap, nil
}
