/*
Copyright © 2024 Yurii Rochniak <yurii@rochniak.dev>
*/
package main

import (
	"sync"

	"github.com/grem11n/aws-cost-meter/client/aws"
	"github.com/grem11n/aws-cost-meter/config"
	"github.com/grem11n/aws-cost-meter/converters/prometheus"
	"github.com/grem11n/aws-cost-meter/logger"
	httpOut "github.com/grem11n/aws-cost-meter/outputs/http"
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
	awsClient, err := aws.New(conf.AWS)
	if err != nil {
		logger.Errorf("Unable to create AWS client: %w", err)
	}

	// Get initial metrics
	awsClient.GetCostAndUsageMatrics(&cache)

	// Get AWS cost metrics in background
	go awsClient.GetCostAndUsageMatricsConcurrently(&cache)

	//// TODO: Move it to a registry
	// Convert metrics in another goroutine
	go prometheus.ConvertAWSMetrics(&cache)

	srv := httpOut.New(conf.Outputs[0], &cache)
	if err := srv.Output(); err != nil {
		logger.Fatalf("cannot start the web server", err)
	}
}
