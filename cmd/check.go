package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var dirPath string

func runCheck(cmd *cobra.Command, args []string) error {
	fmt.Printf("Scanning Directory: %s", dirPath)
	return nil
}

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Checks the provided Directory path",
	RunE:  runCheck,
}

func init() {
	checkCmd.Flags().StringVarP(&dirPath, "dir", "d", "", "Provide the Directory Path")

	if err := checkCmd.MarkFlagRequired("dir"); err != nil {
		fmt.Println("Error setting up required field", err)
		os.Exit(1)
	}

	rootCmd.AddCommand(checkCmd)
}
