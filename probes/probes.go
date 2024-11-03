package probes

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/grem11n/aws-cost-meter/config"
)

const (
	defaultLivenessProbeEndpoint  = "/livez"
	defaultReadinessProbeEndpoint = "/readyz"
	defaultStartupProbeEndpoint   = "/startz"
)

type Probes struct {
	livenessProbeEndpoint  string
	readinessProbeEndpoint string
	startupProbeEndpoint   string
	server                 *http.Server
	mux                    *http.ServeMux
	cache                  *sync.Map
}

func New(conf *config.ProbeConfig, cache *sync.Map) *Probes {
	// Check if probes' endpoints are not empty
	livenessProbeEndpoint := conf.LivenssProbeEndpoint
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

	mux := http.NewServeMux()
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", conf.Port),
		Handler: mux,
	}
	return &Probes{
		livenessProbeEndpoint:  livenessProbeEndpoint,
		readinessProbeEndpoint: readinessProbeEndpoint,
		startupProbeEndpoint:   startupProbeEndpoint,
		server:                 server,
		mux:                    mux,
		cache:                  cache,
	}
}

func (p *Probes) Probe() error {
	p.mux.HandleFunc(p.livenessProbeEndpoint, p.livenessProbe)
	p.mux.HandleFunc(p.readinessProbeEndpoint, p.readinessProbe)
	p.mux.HandleFunc(p.startupProbeEndpoint, p.livenessProbe) // reuse Liveness for Startup
	return p.server.ListenAndServe()
}

func (p *Probes) livenessProbe(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
}

func (p *Probes) readinessProbe(w http.ResponseWriter, r *http.Request) {
	_, ok := p.cache.Load("raw_aws") // hardcoded
	if !ok {
		w.WriteHeader(503)
	}
	w.WriteHeader(200)
}
