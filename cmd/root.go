// Copyright © 2018 Dhananjay Balan <dhananjay@cliqz.com>
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package cmd

import (
	"fmt"
	"os"

	"github.com/cliqz/cloudwatch-writer/prom_cloudwatch_writer"
	"github.com/spf13/cobra"
)

var (
	serverAddr string
	namespace  string
	region     string
	debug      bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "cloudwatch-writer",
	Short: "Export prometheus metrics to cloudwatch",
	Long: `cloudwatch-writer uses the remote writer interface in prometheus to
export metrics to cloudwatch.

please note cloudwatch only allows at most 10 dimensions with a metric, and all metrices with more
dimensions will be automatically ignored by this application`,
	Run: func(cmd *cobra.Command, args []string) {
		err := prom_cldwatch_writer.StartMetricExporter(serverAddr, namespace, region, debug)
		fmt.Fprintf(os.Stderr, "error: %v", err)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "enable debug output")
	rootCmd.Flags().StringVarP(&serverAddr, "serveraddr", "s", "0.0.0.0:1234", "server address listen for prometehus remote writes")
	rootCmd.Flags().StringVarP(&namespace, "namespace", "n", "", "namespace for cloudwatch metrics")
	rootCmd.MarkFlagRequired("namespace")
	rootCmd.Flags().StringVar(&region, "region", "", "aws region")
	rootCmd.MarkFlagRequired("region")
}
