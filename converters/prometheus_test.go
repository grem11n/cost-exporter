package converters

import (
	"sync"
	"testing"

	"github.com/VictoriaMetrics/metrics"
	intmetrics "github.com/grem11n/cost-exporter/internal/metrics"
	"github.com/stretchr/testify/assert"
)

var (
	testProm   = Prometheus{Namespace: "prometheus"}
	testMetric = intmetrics.Metric{
		Name:   "test",
		Prefix: "aws_ce",
		Tags:   map[string]string{"foo": "bar"},
		Value:  0.27,
	}
	testCache = sync.Map{}
)

func TestCreateVMetric(t *testing.T) {
	vm := metrics.NewSet()
	testProm.createVMetric(vm, testMetric)
	assert.Equal(t, 1, len(vm.ListMetricNames()))
}

func TestConvert(t *testing.T) {
	testCache.Clear()
	ns := "test"
	testCache.Store(ns, testMetric)

	ok := testProm.convert(&testCache, ns)
	assert.True(t, ok)

	got, ok := testCache.Load("prometheus")
	assert.True(t, ok)

	gotB, ok := got.([]byte)
	assert.True(t, ok)
	assert.Equal(t, "aws_ce_test{foo=\"bar\"} 0.27\n", string(gotB))
}
