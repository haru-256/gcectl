package cmd

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/haru-256/gce-commands/pkg/config"
	"github.com/haru-256/gce-commands/pkg/log"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(listCmd)
}

var (
	purple = lipgloss.Color("99")
	gray   = lipgloss.Color("#fbfcfc ")

	headerStyle  = lipgloss.NewStyle().Foreground(purple).Bold(true).Align(lipgloss.Center)
	baseRowStyle = lipgloss.NewStyle().Padding(0, 1).Foreground(gray)
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all VM in settings",
	Long:  `All software has versions. This is Hugo's`,
	Run: func(cmd *cobra.Command, args []string) {
		cnf, err := config.ParseConfig(cnfPath)
		if err != nil {
			log.Logger.Fatal(err)
			os.Exit(1)
		}
		log.Logger.Debug(fmt.Sprintf("Config: %+v", cnf))

		// TODO: get status of VMs

		var rows [][]string
		for _, vm := range cnf.VMs {
			rows = append(rows, []string{
				vm.Name,
				vm.Project,
				vm.Zone,
				"Running",
			})
		}

		// render all VMs in settings
		t := table.New().
			Border(lipgloss.NormalBorder()).
			BorderStyle(lipgloss.NewStyle().Foreground(purple)).
			Headers("NAME", "PROJECT", "ZONE", "STATUS").
			Rows(rows...).
			StyleFunc(func(row, col int) lipgloss.Style {
				switch {
				case row == 0:
					return headerStyle
				case col == 3: // status
					return baseRowStyle.Align(lipgloss.Center)
				default:
					return baseRowStyle.Align(lipgloss.Left)
				}
			})
		fmt.Println(t)
	},
}
