package cmd

import (
	"fmt"
	"os"

	"yamlchecker-cli/checker"
	"yamlchecker-cli/internal/config"
	"yamlchecker-cli/internal/validator"

	"github.com/spf13/cobra"
)

var dirPath string

var requiredFields []string

func runCheck(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true
	configErrors := false
	fmt.Printf("Loading from Directory: %s\n", dirPath)
	loaded, err := config.LoadAll(dirPath)
	if err != nil {
		return fmt.Errorf("Error Loading configs from %s: %w", dirPath, err)
	}
	for _, lf := range loaded {
		result, err := lf.Data, lf.Err
		if err != nil {
			fmt.Println("ERROR:", err)
			configErrors = true
			continue
		}
		missing := validator.CheckRequired(result, requiredFields)
		if len(missing) > 0 {
			fmt.Printf("These are missing fields from YAML script: %s ---->", lf.Path)
			fmt.Println(missing)
		} else {
			if len(result) > 0 {
				fmt.Printf("YAML Script parsed without errors:%s -> %v\n", lf.Path, result)
			} else {
				fmt.Printf("YAML Script is empty:%s -> %v\n", lf.Path, result)
			}
		}
	}
	if configErrors {
		return fmt.Errorf("errors found in YAML scripts")
	}
	return nil
}

func runMonitor(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true
	configErrors := false
	unhealthyTargets := false
	loaded, err := config.LoadAll(dirPath)
	if err != nil {
		return fmt.Errorf("Error Loading from Directory %s: %w", dirPath, err)
	}
	var targets []checker.Target
	for _, lf := range loaded {
		result, err := lf.Data, lf.Err
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
			fmt.Printf("skipping %s: missing or invalid 'name' field\n", lf.Path)
			continue
		}

		url, ok := result["url"].(string)
		if !ok {
			fmt.Printf("skipping %s: missing or invalid 'url' field\n", lf.Path)
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
