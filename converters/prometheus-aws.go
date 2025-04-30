package converters

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
	"github.com/ettle/strcase"
	intmetrics "github.com/grem11n/cost-exporter/internal/metrics"
	"github.com/grem11n/cost-exporter/logger"
)

type PrometheusAWS struct{}

const (
	awsCachePrefix      = "aws_"
	namespace           = "prometheus-aws"
	defaultMetricPrefix = "aws_ce"
	// We do not need to convert metrics too frequently,
	// since they are propagated hourly
	cooldown               = 30 // minutes
	costMetricsCounterName = "cost_exporter_cost_metrics_total{job=\"cost-exporter\",converter=\"prometheus-aws\"}"
	conversionDurationName = "cost_exporter_prometheus_aws_conversion_duration{job=\"cost-exporter\"}"
)

var (
	costMetricsCounter *metrics.Counter
	conversionDuration *metrics.Histogram
)

func init() {
	logger.Info("Initializing PrometheusAWS converter")
	Register(namespace, func() Conveter { return &PrometheusAWS{} })
	// Maybe initiate all the metrics in a loop if there are too many
	logger.Info("Initializing PrometheusAWS converter metrics")
	costMetricsCounter = intmetrics.InternalMetricsSet.GetOrCreateCounter(costMetricsCounterName)
	conversionDuration = intmetrics.InternalMetricsSet.GetOrCreateHistogram(conversionDurationName)
}

func (p *PrometheusAWS) Convert(cache *sync.Map) {
	logger.Info("Converting AWS metrics to the Prometheus format")
	for {
		if ok := p.convertAWSMetrics(cache); ok {
			time.Sleep(cooldown * time.Minute)
		}
	}
}

func (p *PrometheusAWS) convertAWSMetrics(cache *sync.Map) bool {
	startTs := time.Now()
	var awsMetrics []costexplorer.GetCostAndUsageOutput
	cache.Range(func(key, value any) bool {
		if strings.HasPrefix(key.(string), awsCachePrefix) {
			res, ok := value.([]costexplorer.GetCostAndUsageOutput)
			if !ok {
				logger.Warn("cache doesn't have %s metrics", awsCachePrefix)
			}
			awsMetrics = append(awsMetrics, res...)
		}
		return true
	})

	// Handle the case if there are no metrics yet
	if len(awsMetrics) == 0 {
		return false
	}

	metricNameMap, err := p.discoverAWSMetrics(awsMetrics)
	if err != nil {
		logger.Error(err)
		return false
	}

	vm := metrics.NewSet()
	for name, groups := range metricNameMap {
		for group, amount := range groups {
			metricName := fmt.Sprintf("%s_%s{job=\"%s\",dimension=\"%s\"}", defaultMetricPrefix, strcase.ToSnake(name), "cost-exporter", group)
			vm.GetOrCreateGauge(metricName, func() float64 {
				return amount
			})
		}
	}
	var res bytes.Buffer
	vm.WritePrometheus(&res)
	logger.Debug("Writing Prometheus metrics to cache with key: ", namespace)
	cache.Swap(namespace, res.Bytes())
	logger.Debug("Prometheus metrics: ", res.String())

	costMetricsCounter.Set(uint64(len(vm.ListMetricNames())))
	conversionDuration.UpdateDuration(startTs)
	return true
}

// Analyze the raw metrics structure to discover, which metrics are present
func (p *PrometheusAWS) discoverAWSMetrics(metrics []costexplorer.GetCostAndUsageOutput) (map[string]map[string]float64, error) {
	// Using maps to convert the raw format and deduplicate metrics "in flight"
	var metricNameMap = make(map[string]map[string]float64)

	for _, metrics := range metrics {
		// There is only a single element in the .ResultsByTime, because of how we craft the time period in the initial request
		if len(metrics.ResultsByTime) == 0 {
			return nil, errors.New("no metrics were found")
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
