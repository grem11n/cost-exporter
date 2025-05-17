// Package probes implements Liveness, Readiness,
// and Startup probes for Kubernetes.
package probes

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/grem11n/cost-exporter/logger"
)

const (
	defaultPort                   = 8989
	defaultLivenessProbeEndpoint  = "/live"
	defaultReadinessProbeEndpoint = "/ready"
	defaultStartupProbeEndpoint   = "/start"
)

// Probes for K8s
type Probes struct {
	Port                   int
	LivenessProbeEndpoint  string `mapstructure:"liveness,omitempty"`
	ReadinessProbeEndpoint string `mapstructure:"readiness,omitempty"`
	StartupProbeEndpoint   string `mapstructure:"startup,omitempty"`
	Cache                  *sync.Map
}

// ProbeConfig stores configuration for K8s probes
type ProbeConfig struct {
	Port                   int
	LivenessProbeEndpoint  string `mapstructure:"liveness,omitempty"`
	ReadinessProbeEndpoint string `mapstructure:"readiness,omitempty"`
	StartupProbeEndpoint   string `mapstructure:"startup,omitempty"`
}

// New returns a pointer to a Probes instance
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

// Run the K8s probes server
func (p *Probes) Run() {
	http.HandleFunc(p.LivenessProbeEndpoint, p.livenessProbe)
	http.HandleFunc(p.ReadinessProbeEndpoint, p.readinessProbe)
	http.HandleFunc(p.StartupProbeEndpoint, p.livenessProbe) // reuse Liveness for Startup

	logger.Info("Starting the probes server")
	server := &http.Server{
		Addr:        fmt.Sprintf(":%d", p.Port),
		ReadTimeout: 10 * time.Second, // hardcoded
	}
	if err := server.ListenAndServe(); err != nil {
		logger.Fatal("Cannot start the probes server: ", err)
	}
}

func (p *Probes) livenessProbe(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(200)
	if _, err := w.Write([]byte("OK")); err != nil {
		logger.Error("LivenessProbe write error: ", err)
	}
}

func (p *Probes) readinessProbe(w http.ResponseWriter, _ *http.Request) {
	code := 200
	message := "OK"
	p.Cache.Range(func(key, _ any) bool {
		if key == nil {
			code = 503
			message = "503 - Metrics cache is empty"
			return false
		}
		return true
	})
	w.WriteHeader(code)
	if _, err := w.Write([]byte(message)); err != nil {
		logger.Error("ReadinessProbe write error: ", err)
	}
}
