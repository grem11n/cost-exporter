package httpOut

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/grem11n/aws-cost-meter/config"
	"github.com/grem11n/aws-cost-meter/converters/prometheus"
	"github.com/grem11n/aws-cost-meter/logger"
)

const (
	cachePrefix = "converted_prometheus"
	// Since metrics are really updated at least each hour, we can cache end results for at least an hour
	defaultTTL = 1 * time.Hour
)

type httpOut struct {
	config *config.OutputsConfig
	server *http.Server
	mux    *http.ServeMux
	cache  *sync.Map
}

type result struct {
	data   bytes.Buffer
	expire int64
}

func New(config *config.OutputsConfig, cache *sync.Map) *httpOut {
	port := 3333
	mux := http.NewServeMux()
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}
	return &httpOut{
		config: config,
		server: server,
		mux:    mux,
		cache:  cache,
	}
}

func (h *httpOut) Output() error {
	h.mux.HandleFunc("/", h.handleRoot)
	h.mux.HandleFunc("/metrics", h.handleMetrics)

	logger.Info(fmt.Sprintf("Started listening on: %s", h.server.Addr))
	return h.server.ListenAndServe()
}

func (h *httpOut) handleRoot(w http.ResponseWriter, r *http.Request) {
	if _, err := io.WriteString(w, "Welcome to Costs Exporter! Check out the /metrics endpoint"); err != nil {
		logger.Error(err)
	}
}

func (h *httpOut) handleMetrics(w http.ResponseWriter, _ *http.Request) {
	logger.Info("Got request for /metrics")
	// Checking the results in cache first
	r, ok := h.cache.Load(cachePrefix)
	res, ok := r.(result)
	if !ok || res.expire < time.Now().Unix() {
		logger.Debug("output cache miss")
		conv := prometheus.ConvertAWSMetrics(h.cache)
		res = result{data: conv, expire: time.Now().Add(defaultTTL).Unix()}
		h.cache.Swap(cachePrefix, res)
	}
	logger.Debug(res.data.String())
	if _, err := io.WriteString(w, res.data.String()); err != nil {
		logger.Error(err)
	}
}
