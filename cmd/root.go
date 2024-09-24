/*
Copyright © 2024 Yurii Rochniak <yurii@rochniak.dev>
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "aws-cost-meter",
	Short: "Expose AWS Cost Explorer metrics in observability-tools friendly format",
	Long: `Export AWS Cost Explorer metrics in a format that external monitoring systems can use.

	Currently, only Prometheus format is supported.

	AWS Cost Meter can expose metrics on an HTTP endpoint or push them into a configured destination, or output to stdout.

	Currently, only HTTP endpoint is implemented.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Push is not implemented yet. Use `aws-cost-meter serve` for to expose metrics on an HTTP endpoint.")
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringP(
		"config",
		"c",
		"config.yaml",
		"Help message for toggle",
	)
}
