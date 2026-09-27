package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "yamlchecker",
	Short: "Validates YAML configs against required fields",
}

func Execute() error {
	return rootCmd.Execute()
}
