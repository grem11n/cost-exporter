package outputs

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/grem11n/cost-exporter/logger"
)

type HTTP struct {
	Path string
	Port int
}

const (
	defaultPath = "/metrics"
	defaultPort = 8080
)

func init() {
	logger.Info("Initializing HTTP output")
	Register("http", func(OutputConfig) Output { return &HTTP{} })
}

func (h *HTTP) Publish(keys []string, cache *sync.Map) {
	path := h.Path
	if path == "" {
		logger.Infof("Using the default metrics path: ", defaultPath)
		path = defaultPath
	}
	port := h.Port
	if port <= 0 || port > 65535 {
		logger.Infof("Using the default port: %d", defaultPort)
		port = defaultPort
	}
	http.HandleFunc("/", h.handleRoot(path))
	http.HandleFunc(path, h.handleMetrics(keys, cache))

	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		ReadHeaderTimeout: 10 * time.Second, // hardcoded
	}

	logger.Info(fmt.Sprintf("Started listening on: \":%d\"", port))
	if err := server.ListenAndServe(); err != nil {
		logger.Fatal("Cannot start HTTP server: ", err)
	}
}

func (h *HTTP) handleRoot(metricsPath string) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		if _, err := io.WriteString(
			w, fmt.Sprintf("Welcome to Costs Exporter! Cost metrics are available at: %s", metricsPath),
		); err != nil {
			logger.Error(err)
		}
	}
}

func (h *HTTP) handleMetrics(keys []string, cache *sync.Map) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		logger.Info("Got request for metrics")
		var res bytes.Buffer
		for _, key := range keys {
			r, ok := cache.Load(key)
			if !ok {
				logger.Error("Cannot get metrics from cache")
				w.WriteHeader(http.StatusInternalServerError)
				_, err := w.Write([]byte("500 - Cannot get metrics from cache"))
				if err != nil {
					logger.Error(err)
				}
				return
			}
			rb, ok := r.([]byte)
			if !ok {
				logger.Error("Odd metrics format")
				w.WriteHeader(http.StatusInternalServerError)
				_, err := w.Write([]byte("500 - Odd metrics format"))
				if err != nil {
					logger.Error(err)
				}
				return
			}
			res.WriteString(fmt.Sprintf("# Metrics from %s\r\n", key))
			res.Write(rb)
		}

		logger.Debug(res.String())
		if _, err := io.WriteString(w, res.String()); err != nil {
			logger.Error(err)
			w.WriteHeader(http.StatusInternalServerError)
			_, err := w.Write([]byte(err.Error()))
			if err != nil {
				logger.Error(err)
			}
		}
	}
}
