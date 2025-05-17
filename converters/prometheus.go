package converters

import (
	"bytes"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/VictoriaMetrics/metrics"
	"github.com/ettle/strcase"
	intmetrics "github.com/grem11n/cost-exporter/internal/metrics"
	"github.com/grem11n/cost-exporter/logger"
)

type Prometheus struct{}

const (
	namespace = "prometheus"
	// We do not need to convert metrics too frequently,
	// since they are propagated hourly
	cooldown               = 30 // minutes
	costMetricsCounterName = "cost_exporter_cost_metrics_total{job=\"cost-exporter\",converter=\"prometheus\"}"
	conversionDurationName = "cost_exporter_prometheus_aws_conversion_duration{job=\"cost-exporter\"}"
)

var (
	costMetricsCounter *metrics.Counter
	conversionDuration *metrics.Histogram
)

func init() {
	logger.Info("Initializing PrometheusAWS converter")
	Register(namespace, func() Conveter { return &Prometheus{} })
	// Maybe initiate all the metrics in a loop if there are too many
	logger.Info("Initializing PrometheusAWS converter metrics")
	costMetricsCounter = intmetrics.InternalMetricsSet.GetOrCreateCounter(costMetricsCounterName)
	conversionDuration = intmetrics.InternalMetricsSet.GetOrCreateHistogram(conversionDurationName)
}

func (p *Prometheus) Convert(cache *sync.Map, fetchPrefix string) {
	logger.Info("Converting AWS metrics to the Prometheus format")
	for {
		if ok := p.convert(cache, fetchPrefix); ok {
			time.Sleep(cooldown * time.Minute)
		}
	}
}

func (p *Prometheus) convert(cache *sync.Map, fetchPrefix string) bool {
	startTs := time.Now()
	vm := metrics.NewSet()
	cache.Range(func(key, value any) bool {
		ks, ok := key.(string)
		if !ok {
			logger.Error("cache key is not a string, got %T", key)
			// Return on this iteration, but continue the loop
			return true
		}
		if strings.HasPrefix(ks, fetchPrefix) {
			metric, ok := value.(intmetrics.Metric)
			if !ok {
				logger.Warn("wrong metric type found in %s: %T", fetchPrefix, value)
			}
			p.createVMetric(vm, metric)
		}
		return true
	})

	// Handle the case when the metrics are not yet present
	if len(vm.ListMetricNames()) == 0 {
		return false
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

func (p *Prometheus) createVMetric(vm *metrics.Set, metric intmetrics.Metric) {
	logger.Debug("Got metric: ", metric)
	var tags []string
	for k, v := range metric.Tags {
		tags = append(tags, fmt.Sprintf("%s=\"%s\"", k, v))
	}
	tagStr := strings.Join(tags, ",")
	metricName := fmt.Sprintf(
		"%s_%s{%s}",
		strcase.ToSnake(metric.Prefix),
		strcase.ToSnake(metric.Name),
		tagStr,
	)
	vm.GetOrCreateGauge(metricName, func() float64 {
		return metric.Value
	})
}
