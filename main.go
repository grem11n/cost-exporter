/*
Copyright © 2024 Yurii Rochniak <yurii@rochniak.dev>
*/
package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/grem11n/aws-cost-meter/client/aws"
	"github.com/grem11n/aws-cost-meter/config"
	"github.com/grem11n/aws-cost-meter/logger"
	flag "github.com/spf13/pflag"
)

var (
	configPath   *string = flag.StringP("config", "c", "./config.yaml", "Path to the configuration file")
	rawCache     sync.Map
	metricsCache sync.Map
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

	// Get AWS cost metrics in background
	go awsClient.GetCostAndUsageMatrics(&rawCache)

	time.Sleep(5 * time.Second)
	got, ok := rawCache.Load("aws")
	if !ok {
		fmt.Println("chache is empty")
	}
	fmt.Printf("%+v", got)

	//// Get the raw metrics cache
	//if err := prometheus.ConvertRawMetrics(rawMetrics); err != nil {
	//	logger.Errorf(fmt.Sprintf("%s", err))
	//}

	//// TODO: Move it to a registry
	//srv := httpOut.New(conf.Outputs[0])
	//srv.Output()
}
