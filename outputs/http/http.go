package httpOut

import (
	"fmt"
	"io"
	"net/http"

	"github.com/grem11n/aws-cost-meter/cache"
	"github.com/grem11n/aws-cost-meter/config"
	"github.com/grem11n/aws-cost-meter/logger"
)

type httpOut struct {
	config config.OutputsConfig
	server *http.Server
	mux    *http.ServeMux
}

func New(config config.OutputsConfig) *httpOut {
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
	}
}

func (h *httpOut) Output() error {
	h.mux.HandleFunc("/", h.handleRoot)
	h.mux.HandleFunc("/metrics", h.handleMetrics)

	logger.Info(fmt.Sprintf("Started listening on: %s", h.server.Addr))
	return h.server.ListenAndServe()
}

func (h *httpOut) handleRoot(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Welcome to Costs Exporter! Check out the /metrics endpoint")
}

func (h *httpOut) handleMetrics(w http.ResponseWriter, _ *http.Request) {
	mainCache := cache.GetMainCache()
	fmt.Printf("%+v\n", mainCache.Cache)
	res := mainCache.Cache[string(h.config.Converter)]
	logger.Info(res.String())
	io.WriteString(w, res.String())
}
