package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "fstimeline",
	Short: "A file system timeline monitor",
	Long:  `Monitor file system changes over time and query historical events.`,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(watchCmd)
	rootCmd.AddCommand(queryCmd)
	rootCmd.AddCommand(exportCmd)
}
