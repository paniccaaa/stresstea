/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// versionCmd represents the version command
const Version = "0.1.0"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show Stresstea version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Stresstea v%s\n", Version)
		fmt.Println("CLI tool for load testing HTTP and gRPC services")
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// versionCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// versionCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
