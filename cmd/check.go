package cmd

import (
	"fmt"
	"os"

	"yamlchecker-cli/checker"
	"yamlchecker-cli/internal/config"
	"yamlchecker-cli/internal/scanner"
	"yamlchecker-cli/internal/validator"

	"github.com/spf13/cobra"
)

var dirPath string

var requiredFields []string

var anyErrors bool

var configErrors bool

var unhealthyTargets bool

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

func runMonitor(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true
	files, err := scanner.FindYAMLFiles(dirPath)
	if err != nil {
		return fmt.Errorf("scanning %s: %w", dirPath, err)
	}
	var targets []checker.Target
	for _, file := range files {
		result, err := config.Load(file)
		if err != nil {
			fmt.Println("ERROR:", err)
			configErrors = true
			continue
		}
		if result["type"] != "http" {
			continue // not an HTTP target, skip silently
		}

		name, ok := result["name"].(string)
		if !ok {
			fmt.Printf("skipping %s: missing or invalid 'name' field\n", file)
			continue
		}

		url, ok := result["url"].(string)
		if !ok {
			fmt.Printf("skipping %s: missing or invalid 'url' field\n", file)
			continue
		}

		targets = append(targets, checker.Target{Name: name, URL: url})
	}
	fmt.Printf("Monitoring %d HTTP endpoints\n", len(targets))
	results := checker.RunAll(targets)
	for _, result := range results {
		fmt.Printf("Name: %s, URL: %s, Success: %t, Status: %d, Duration: %s, Err: %v\n",
			result.Name,
			result.URL,
			result.Success,
			result.Status,
			result.Duration,
			result.Err,
		)
		if !result.Success {
			unhealthyTargets = true
		}
	}
	if configErrors && unhealthyTargets {
		return fmt.Errorf("config errors found, and one or more endpoints are unhealthy")
	} else if configErrors {
		return fmt.Errorf("one or more config files could not be parsed")
	} else if unhealthyTargets {
		return fmt.Errorf("one or more endpoints failed health check")
	}
	return nil
}

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Checks the provided Directory path",
	RunE:  runCheck,
}

var monitorCmd = &cobra.Command{
	Use:   "monitor",
	Short: "Monitors the provided HTTP endpoints",
	RunE:  runMonitor,
}

func init() {
	checkCmd.Flags().StringVarP(&dirPath, "dir", "d", "", "Provide the Directory Path")
	checkCmd.Flags().StringSliceVarP(&requiredFields, "required", "r", []string{}, "Provide the required fields comma separated")
	monitorCmd.Flags().StringVarP(&dirPath, "dir", "d", "", "Provide the Directory to Report HTTP EndPoints")
	if err := checkCmd.MarkFlagRequired("dir"); err != nil {
		fmt.Println("Error setting up required field", err)
		os.Exit(1)
	}

	rootCmd.AddCommand(checkCmd)
	rootCmd.AddCommand(monitorCmd)
}
