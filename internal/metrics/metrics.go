/*
This package contains a VictoriaMetrics Set
to write metrics produced by the cost-exporter itself
*/
package intmetrics

import (
	"bytes"
	"sync"
	"time"

	"github.com/VictoriaMetrics/metrics"
)

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
