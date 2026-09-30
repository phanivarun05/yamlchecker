package cmd

import (
	"fmt"
	"os"

	"yamlchecker-cli/internal/config"
	"yamlchecker-cli/internal/scanner"
	"yamlchecker-cli/internal/validator"

	"github.com/spf13/cobra"
)

var dirPath string

var required = []string{"runtime", "instance_class", "host"}

func runCheck(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true
	fmt.Printf("Scanning Directory: %s\n", dirPath)
	files, err := scanner.FindYAMLFiles(dirPath)
	if err != nil {
		return fmt.Errorf("scanning %s: %w", dirPath, err)
	}
	if len(files) == 0 {
		fmt.Printf("\nNo YAML files founded in directory %s", dirPath)
	}
	fmt.Printf("\nBelow are the YAML files mapped from directory %s\n", dirPath)
	for _, file := range files {
		result, err := config.Load(file)
		if err != nil {
			fmt.Println("ERROR:", err)
			continue
		}
		missing := validator.CheckRequired(result, required)
		if len(missing) > 0 {
			fmt.Printf("These are missing fields from YAML script: %s ---->", file)
			fmt.Println(missing)
		}
		fmt.Printf("%s -> %#v\n", file, result)
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
