package cache

import (
	"sync"

	"github.com/aws/aws-sdk-go-v2/service/costexplorer"
	"github.com/grem11n/aws-cost-meter/logger"
)

var lock = &sync.Mutex{}

// RawCache is a singleton that holds AWS outputs for any type of metrics
type RawCache struct {
	// Basic CostAndUsageMetrics
	CostAndUsageMetrics []*costexplorer.GetCostAndUsageOutput
}

var rawCacheInstance *RawCache

func GetRawCache() *RawCache {
	if rawCacheInstance == nil {
		lock.Lock()
		defer lock.Unlock()
		if rawCacheInstance == nil {
			logger.Info("Creating a new raw cache since it doesn't exist")
			rawCacheInstance = &RawCache{}
		}
	}
	return rawCacheInstance
}
