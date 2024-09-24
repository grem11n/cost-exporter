/*
Copyright © 2024 Yurii Rochniak <yurii@rochniak.dev>
*/
package main

import (
	"github.com/grem11n/aws-cost-meter/actions/serve"
	"github.com/grem11n/aws-cost-meter/config"
	"github.com/grem11n/aws-cost-meter/logger"
	flag "github.com/spf13/pflag"
)

var configPath *string = flag.StringP("config", "c", "./config.yaml", "Path to the configuration file")

func main() {
	flag.Parse()
	config, err := config.New(*configPath)
	if err != nil {
		logger.Fatalf("Unable to read the config file: ", err)
	}
	if err := serve.Run(config); err != nil {
		logger.Fatalf("Unable to start the server: ", err)
	}
}
