// Copyright © 2018 Dhananjay Balan <dhananjay@cliqz.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
	rootCmd.Flags().StringVar(&region, "region", "", "aws reguion")
	rootCmd.MarkFlagRequired("region")
}
