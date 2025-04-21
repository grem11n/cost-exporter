package main

import (
	"fmt"
	"sync"

	"github.com/grem11n/cost-exporter/clients"
	"github.com/grem11n/cost-exporter/config"
	"github.com/grem11n/cost-exporter/converters"
	"github.com/grem11n/cost-exporter/logger"
	"github.com/grem11n/cost-exporter/outputs"
	"github.com/grem11n/cost-exporter/probes"

	flag "github.com/spf13/pflag"
)

type App struct {
	MetricsFormat string
	Clients       map[string]clients.Client
	Converters    map[string]converters.Conveter
	Outputs       map[string]outputs.Output
}

var (
	configPath *string = flag.StringP("config", "c", "./config.yaml", "Path to the configuration file")
	// Create new global cache as an exchange point
	cache sync.Map
	app   = App{
		MetricsFormat: "prometheus", // only Prometheus is supported for now
		Clients:       make(map[string]clients.Client),
		Converters:    make(map[string]converters.Conveter),
		Outputs:       make(map[string]outputs.Output),
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

	// Convert raw metrics into a format suitable for output
	// Only Prometheus is supported for now
	convertedKeys := getConvertedKeys()
	for _, converterName := range convertedKeys {
		constructor := converters.GetConverter(converterName)
		if constructor == nil {
			logger.Fatalf("Converter %s doesn't exist", converterName)
		}
		converter := constructor()
		app.Converters[converterName] = converter
	}

	for _, cv := range app.Converters {
		go cv.Convert(&cache)
	}

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

	// Output the metrics
	fmt.Println(app.Outputs)
	for _, out := range app.Outputs {
		out.Publish(convertedKeys, &cache)
	}
}

// Get a slice of cache keys that store converted metrics ready for outputs
func getConvertedKeys() []string {
	var convertedKeys = []string{}
	for clientName := range app.Clients {
		convertedKeys = append(
			convertedKeys, fmt.Sprintf(
				"%s-%s", app.MetricsFormat, clientName,
			),
		)
	}
	return convertedKeys
}
