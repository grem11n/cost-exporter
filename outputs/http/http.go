package httpOut

import (
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/grem11n/aws-cost-meter/config"
	"github.com/grem11n/aws-cost-meter/logger"
)

type httpOut struct {
	config config.OutputsConfig
	server *http.Server
	mux    *http.ServeMux
	cache  *sync.Map
}

func New(config config.OutputsConfig, cache *sync.Map) *httpOut {
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
	res, ok := h.cache.Load("aws_processed") // hardcoded for test
	if !ok {
		logger.Warn("cannot find metrics in the metric cache")
	}
	out := res.(string) // because we write string
	logger.Debug(out)
	if _, err := io.WriteString(w, out); err != nil {
		logger.Error(err)
	}
}
