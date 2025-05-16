/*
This package contains a VictoriaMetrics Set
to write metrics produced by the cost-exporter itself
*/
package metrics

import (
	"bytes"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/VictoriaMetrics/metrics"
)

var (
	defaultTags = map[string]string{
		"job": "cost-exporter",
	}
)

type Metric struct {
	Value float64
	Name  string
	Tags  map[string]string
}

func AddMetrics(cache *sync.Map, namespace string, metrics []Metric) {
	for _, m := range metrics {
		AddMetric(cache, namespace, m)
	}
}

func AddMetric(cache *sync.Map, namespace string, metric Metric) {
	metric.addDefaultTags()
	tags := []string{}
	for k := range metric.Tags {
		tags = append(tags, k)
	}
	tagsStr := strings.Join(tags, "_")
	key := fmt.Sprintf("%s_%s_%s", namespace, metric.Name, tagsStr)
	cache.Swap(key, metric)
}

func (m *Metric) addDefaultTags() {
	for k, v := range defaultTags {
		m.Tags[k] = v
	}
}

const cooldown = 10 // seconds

// Global Set for self metrics
var InternalMetricsSet *metrics.Set

func init() {
	InternalMetricsSet = metrics.NewSet()
}

// Publish metrics to the cache in the Prometheus format
func Publish(key string, cache *sync.Map) {
	for {
		publish(key, cache)
		time.Sleep(cooldown * time.Second)
	}
}
func publish(key string, cache *sync.Map) {
	var res bytes.Buffer
	InternalMetricsSet.WritePrometheus(&res)
	cache.Swap(key, res.Bytes())
}
