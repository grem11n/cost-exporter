package main

import (
	"sync"

	"github.com/grem11n/cost-exporter/clients"
	"github.com/grem11n/cost-exporter/config"
	"github.com/grem11n/cost-exporter/converters"
	intmetrics "github.com/grem11n/cost-exporter/internal/metrics"
	"github.com/grem11n/cost-exporter/logger"
	"github.com/grem11n/cost-exporter/outputs"
	"github.com/grem11n/cost-exporter/probes"

	flag "github.com/spf13/pflag"
)

// App is a struct that holds parts for the application together
type App struct {
	Clients   map[string]clients.Client
	Converter converters.Converter
	Outputs   map[string]outputs.Output
}

const (
	internalMetricsKey = "prometheus-internal"
)

var (
	configPath *string = flag.StringP("config", "c", "./config.yaml", "Path to the configuration file")
	// Create new global cache as an exchange point
	cache sync.Map
	app   = App{
		Clients: make(map[string]clients.Client),
		Outputs: make(map[string]outputs.Output),
	}
)

func main() {
	flag.Parse()
	conf, err := config.New(*configPath)
	logger.Debug("Config: ", conf)
	if err != nil {
		logger.Fatalf("Unable to read the config file: ", err)
	}

	// Start the probes server
	probes := probes.New(&conf.Probes, &cache)
	go probes.Run()

	// Get cloud clients from the registry
	for clientName, clientConfig := range conf.Clients {
		constructor := clients.GetClient(clientName)
		if constructor == nil {
			logger.Fatalf("Client %s doesn't exist", clientName)
		}
		logger.Debug("Client config: ", clientConfig)
		client := constructor(clientConfig)
		app.Clients[clientName] = client
	}

	// Populate the cache with raw metrics
	for _, cl := range app.Clients {
		go cl.GetMetrics(&cache)
	}

	if len(conf.MetricsFormat) != 1 {
		logger.Fatalf("only a single metrics format is supported")
	}

	var converterName string
	var converterConfig converters.ConverterConfig
	for k, v := range conf.MetricsFormat {
		converterName, converterConfig = k, v
		break

	}

	constructor := converters.GetConverter(converterName)
	if constructor == nil {
		logger.Fatalf("Converter %s doesn't exist", conf.MetricsFormat)
	}
	converter := constructor(converterConfig)
	app.Converter = converter

	// Convert metrics from the input to the output format
	// Cache key prefix is hardcoded, because only AWS is supported for now
	go app.Converter.Convert(&cache, "aws_")

	// Collect the internal metrics
	go intmetrics.Publish(internalMetricsKey, &cache)

	// Get the outputs from the registry
	for outputName, outputConfig := range conf.Outputs {
		constructor := outputs.GetOutput(outputName)
		logger.Debug("Outputs constructor: ", constructor)
		if constructor == nil {
			logger.Fatalf("Output %s doesn't exist", outputName)
		}
		logger.Debug("Output config: ", outputConfig)
		output := constructor(outputConfig)
		app.Outputs[outputName] = output
	}

	// Output the metrics + append the internal metrics
	for _, out := range app.Outputs {
		out.Publish(&cache, []string{converterName, internalMetricsKey})
	}
}
