package prometheus

import (
	"fmt"

	"github.com/VictoriaMetrics/metrics"
	"github.com/grem11n/aws-cost-meter/cache"
)

func ConvertRawMetrics(raw *cache.RawCache) (*metrics.Set, error) {
	for _, metrics := range raw.CostAndUsageMetrics {
		for _, cost := range metrics.ResultsByTime {
			fmt.Println(cost)
		}
	}
	return nil, nil
}
