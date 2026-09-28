package cmd

import (
	"fmt"
	"os"

	"yamlchecker-cli/internal/scanner"

	"github.com/spf13/cobra"
)

var dirPath string

func runCheck(cmd *cobra.Command, args []string) error {
	fmt.Printf("Scanning Directory: %s", dirPath)
	files, err := scanner.FindYAMLFiles(dirPath)
	if err != nil {
		fmt.Printf("\nError Finding YAML files in %s directory", dirPath)
		return err
	}
	if len(files) == 0 {
		fmt.Printf("\nNo YAML files founded in directory %s", dirPath)
	}
	fmt.Printf("\nBelow are the YAML files from directory %s\n", dirPath)
	for _, file := range files {
		fmt.Println(file)
	}
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
