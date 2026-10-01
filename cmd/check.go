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

var requiredFields []string

var anyErrors bool

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
			anyErrors = true
			continue
		}
		missing := validator.CheckRequired(result, requiredFields)
		if len(missing) > 0 {
			fmt.Printf("These are missing fields from YAML script: %s ---->", file)
			fmt.Println(missing)
		} else {
			if len(result) > 0 {
				fmt.Printf("YAML Script parsed without errors:%s -> %v\n", file, result)
			} else {
				fmt.Printf("YAML Script is empty:%s -> %v\n", file, result)
			}
		}
	}
	if anyErrors {
		return fmt.Errorf("errors found in YAML scripts")
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
	checkCmd.Flags().StringSliceVarP(&requiredFields, "required", "r", []string{}, "Provide the required fields comma separated")
	if err := checkCmd.MarkFlagRequired("dir"); err != nil {
		fmt.Println("Error setting up required field", err)
		os.Exit(1)
	}

	rootCmd.AddCommand(checkCmd)
}
