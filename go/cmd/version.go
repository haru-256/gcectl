package cmd

import (
	"github.com/haru-256/gcectl/internal/interface/presenter"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number and build info",
	Run: func(cmd *cobra.Command, args []string) {
		console := presenter.NewConsolePresenter()
		console.RenderVersion(appVersion, appCommit, appDate)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
