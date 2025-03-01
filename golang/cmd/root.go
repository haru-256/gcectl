/*
Copyright © 2025 yohei.kuro48@gmail.com

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"os"

	"github.com/haru-256/gce-commands/cmd/set"
	"github.com/haru-256/gce-commands/pkg/log"
	"github.com/spf13/cobra"
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "gce-commands",
	Short: "Google Compute Engine commands",
	Long:  `Google Compute Engine commands such as listing vm and update vm-spec, add vm into stop-scheduler.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Logger.Debug("run root command")
		if err := cmd.Help(); err != nil {
			log.Logger.Fatal(err)
			os.Exit(1)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the RootCmd.
func Execute() {
	err := RootCmd.Execute()
	if err != nil {
		log.Logger.Fatal(err)
		os.Exit(1)
	}
}

var (
	project string
	zone    string
	CnfPath string
)

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	RootCmd.PersistentFlags().StringVar(&project, "project", "haru256-sandbox-20250224", "GCP Project ID")
	RootCmd.PersistentFlags().StringVarP(&zone, "zone", "z", "asia-northeast1-a", "zone or location in GCP")
	RootCmd.PersistentFlags().StringVarP(&CnfPath, "config", "c", "~/gce-commands.yaml", "config file path")

	// set sub command
	RootCmd.AddCommand(set.SetCmd)
}
