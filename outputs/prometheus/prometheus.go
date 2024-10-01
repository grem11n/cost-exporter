package prometheus

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/VictoriaMetrics/metrics"
	"github.com/grem11n/aws-cost-meter/cache"
	"github.com/grem11n/aws-cost-meter/logger"
	"github.com/iancoleman/strcase"
)

// TODO: Move to config
const defaultMetricPrefix = "ce_exporter"

type PrometheusConfig struct{}

func (p *PrometheusConfig) Output() error {
	return nil
}

func ConvertRawMetrics(raw *cache.RawCache) error {
	metricNameMap, err := discoverMetrics(raw)
	if err != nil {
		return err
	}
	fmt.Printf("Metrics: %+v\n", metricNameMap)

	vm := metrics.NewSet()
	for name, groups := range metricNameMap {
		for group, amount := range groups {
			metricName := fmt.Sprintf("%s_%s{job=\"%s\",dimension=\"%s\"}", defaultMetricPrefix, strcase.ToSnake(name), "ce-exporter", group)
			fmt.Printf("name: %s. amount: %f", metricName, amount)
			vm.GetOrCreateGauge(metricName, func() float64 {
				return amount
			})
		}
	}
	vm.WritePrometheus(os.Stdout)
	return nil
}

// Analyze the raw metrics structure to discover, which metrics are present
func discoverMetrics(raw *cache.RawCache) (map[string]map[string]float64, error) {
	// Using maps to convert the raw format and deduplicate metrics "in flight"
	var metricNameMap = make(map[string]map[string]float64)

	for _, metrics := range raw.CostAndUsageMetrics {
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
