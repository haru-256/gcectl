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
	"fmt"
	"os"

	"github.com/haru-256/gcectl/cmd/set"
	infraLog "github.com/haru-256/gcectl/internal/infrastructure/log"
	"github.com/haru-256/gcectl/internal/interface/presenter"
	"github.com/spf13/cobra"
)

var (
	// CnfPath is the path to the configuration file
	CnfPath string
	// Package-level variables to store values passed from main.
	appVersion string
	appCommit  string
	appDate    string
)

// SetVersionInfo is called from main.go to set the version information.
// It's an exported function so that its variables can be overwritten by GoReleaser's ldflags.
func SetVersionInfo(version, commit, date string) {
	appVersion = version
	appCommit = commit
	appDate = date
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "gcectl [command]",
	Short: "Google Compute Engine commands to control VMs",
	Long:  `Google Compute Engine commands to control VMs such as listing vm and updating vm-spec, attach vm with stop-scheduler.`,
	Run: func(cmd *cobra.Command, args []string) {
		infraLog.DefaultLogger.Debugf("run root command")
		if err := cmd.Help(); err != nil {
			infraLog.DefaultLogger.Fatalf("failed to show help: %v", err)
			os.Exit(1)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	// mainでSetVersionInfoが呼び出されてから実行
	rootCmd.Version = appVersion
	err := rootCmd.Execute()
	if err != nil {
		infraLog.DefaultLogger.Fatalf("failed to execute command: %v", err)
		os.Exit(1)
	}
}

func init() {
	console := presenter.NewConsolePresenter()
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	home, err := os.UserHomeDir()
	if err != nil {
		console.Error(fmt.Sprintf("failed to get user home directory: %v", err))
		os.Exit(1)
	}
	defaultCnfPath := home + "/.config/gcectl/config.yaml"
	rootCmd.PersistentFlags().StringVarP(&CnfPath, "config", "c", defaultCnfPath, "config file path")

	// set sub command
	rootCmd.AddCommand(set.SetCmd)
}
