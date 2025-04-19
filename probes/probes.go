package probes

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/grem11n/cost-exporter/logger"
)

const (
	defaultPort                   = 8989
	defaultLivenessProbeEndpoint  = "/live"
	defaultReadinessProbeEndpoint = "/ready"
	defaultStartupProbeEndpoint   = "/start"
)

type Probes struct {
	Port                   int
	LivenessProbeEndpoint  string `mapstructure:"liveness,omitempty"`
	ReadinessProbeEndpoint string `mapstructure:"readiness,omitempty"`
	StartupProbeEndpoint   string `mapstructure:"startup,omitempty"`
	Cache                  *sync.Map
}

type ProbeConfig struct {
	Port                   int
	LivenessProbeEndpoint  string `mapstructure:"liveness,omitempty"`
	ReadinessProbeEndpoint string `mapstructure:"readiness,omitempty"`
	StartupProbeEndpoint   string `mapstructure:"startup,omitempty"`
}

func New(conf *ProbeConfig, cache *sync.Map) *Probes {
	// Check if probes' endpoints are not empty
	livenessProbeEndpoint := conf.LivenessProbeEndpoint
	if livenessProbeEndpoint == "" {
		livenessProbeEndpoint = defaultLivenessProbeEndpoint
	}

	readinessProbeEndpoint := conf.ReadinessProbeEndpoint
	if readinessProbeEndpoint == "" {
		readinessProbeEndpoint = defaultReadinessProbeEndpoint
	}

	startupProbeEndpoint := conf.StartupProbeEndpoint
	if startupProbeEndpoint == "" {
		startupProbeEndpoint = defaultStartupProbeEndpoint
	}

	port := conf.Port
	if port <= 0 || port > 65535 {
		logger.Infof("Using the default port: %d", defaultPort)
		port = defaultPort
	}

	return &Probes{
		Port:                   port,
		LivenessProbeEndpoint:  livenessProbeEndpoint,
		ReadinessProbeEndpoint: readinessProbeEndpoint,
		StartupProbeEndpoint:   startupProbeEndpoint,
		Cache:                  cache,
	}
}

func (p *Probes) Run() {
	http.HandleFunc(p.LivenessProbeEndpoint, p.livenessProbe)
	http.HandleFunc(p.ReadinessProbeEndpoint, p.readinessProbe)
	http.HandleFunc(p.StartupProbeEndpoint, p.livenessProbe) // reuse Liveness for Startup
	logger.Info("Starting the probes server")
	if err := http.ListenAndServe(fmt.Sprintf(":%d", p.Port), nil); err != nil {
		logger.Fatal("Cannot start the probes server: ", err)
	}
}

func (p *Probes) livenessProbe(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	w.Write([]byte("OK"))
}

func (p *Probes) readinessProbe(w http.ResponseWriter, r *http.Request) {
	code := 200
	message := "OK"
	p.Cache.Range(func(key, value any) bool {
		if key == nil {
			code = 503
			message = "503 - Metrics cache is empty"
			return false
		}
		return true
	})
	w.WriteHeader(code)
	w.Write([]byte(message))
}
