/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"time"

	"github.com/paniccaaa/stresstea/internal/config"
	"github.com/paniccaaa/stresstea/internal/engine"
	"github.com/paniccaaa/stresstea/internal/parser"
	"github.com/spf13/cobra"
)

var (
	target     string
	duration   time.Duration
	rate       int
	concurrent int
	configFile string
	protocol   string
	cpus       int
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run [target]",
	Short: "Run load testing",
	Long: `Runs load testing against the specified service.
Supports HTTP and gRPC protocols.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 {
			target = args[0]
		}

		if target == "" && configFile == "" {
			return fmt.Errorf("target or config file must be specified")
		}

		var cfg *parser.Config
		var err error

		if configFile != "" {
			cfg, err = parser.LoadFromFile(configFile)
			if err != nil {
				return fmt.Errorf("failed to load configuration: %w", err)
			}
		} else {
			cfg = &parser.Config{
				App: config.DefaultAppConfig(),
				Test: &parser.TestRunConfig{
					Target:     target,
					Duration:   duration,
					Rate:       rate,
					Concurrent: concurrent,
					Protocol:   protocol,
					CPUs:       cpus,
				},
			}
		}

		return engine.Run(cfg)
	},
}

func init() {
	rootCmd.AddCommand(runCmd)

	// Define flags for the run command
	runCmd.Flags().StringVarP(&target, "target", "t", "", "Target URL or gRPC endpoint")
	runCmd.Flags().DurationVarP(&duration, "duration", "d", 30*time.Second, "Test duration")
	runCmd.Flags().IntVarP(&rate, "rate", "r", 100, "Requests per second")
	runCmd.Flags().IntVarP(&concurrent, "concurrent", "c", 10, "Number of concurrent connections")
	runCmd.Flags().StringVarP(&configFile, "config", "f", "", "Path to YAML configuration file")
	runCmd.Flags().StringVarP(&protocol, "protocol", "p", "http", "Protocol (http or grpc)")
	runCmd.Flags().IntVarP(&cpus, "cpus", "", 0, "Number of CPUs to use (0 = all available)")
}
