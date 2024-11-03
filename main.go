/*
Copyright © 2024 Yurii Rochniak <yurii@rochniak.dev>
*/
package main

import (
	"sync"

	"github.com/grem11n/aws-cost-meter/client/aws"
	"github.com/grem11n/aws-cost-meter/config"
	"github.com/grem11n/aws-cost-meter/logger"
	httpOut "github.com/grem11n/aws-cost-meter/outputs/http"
	"github.com/grem11n/aws-cost-meter/probes"
	flag "github.com/spf13/pflag"
)

var (
	configPath *string = flag.StringP("config", "c", "./config.yaml", "Path to the configuration file")
	cache      sync.Map
)

func main() {
	flag.Parse()
	conf, err := config.New(*configPath)
	if err != nil {
		logger.Fatalf("Unable to read the config file: ", err)
	}
	awsClient, err := aws.New(conf.AWS, &cache)
	if err != nil {
		logger.Fatalf("Unable to create AWS client: %w", err)
	}

	// Initialize Kubernetes probes
	probeConf := &config.ProbeConfig{
		Port: 8999,
	}
	prober := probes.New(probeConf, &cache)
	go prober.Probe()

	// Get initial metrics
	awsClient.GetCostAndUsageMatrics()

	// Get AWS cost metrics in background
	go awsClient.GetCostAndUsageMatricsLoop()

	srv := httpOut.New(conf.Outputs[0], &cache)
	if err := srv.Output(); err != nil {
		logger.Fatalf("cannot start the web server", err)
	}
}
