/*
Copyright © 2024 Yurii Rochniak <yurii@rochniak.dev>
*/
package cmd

import (
	"github.com/grem11n/aws-cost-meter/actions/serve"
	"github.com/grem11n/aws-cost-meter/config"
	"github.com/grem11n/aws-cost-meter/logger"
	"github.com/spf13/cobra"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Run AWS Cost Meter as a daemon.",
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := config.New(cmd.Flags())
		if err != nil {
			logger.Fatalf("Unable to read the config file: ", err)
		}
		if err := serve.Run(config); err != nil {
			logger.Fatalf("Unable to start the server: ", err)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
}
