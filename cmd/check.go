package cmd

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"time"

	"yamlchecker-cli/checker"
	"yamlchecker-cli/internal/config"
	"yamlchecker-cli/internal/validator"

	"github.com/spf13/cobra"
)

var dirPath string

var requiredFields []string

var workers int

var interval time.Duration

// # helper
func targetFromEntry(entry map[string]interface{}) (checker.Target, error) {
	Type, ok := entry["type"].(string)
	if !ok || Type != "http" {
		return checker.Target{}, errors.New("No http type found")
	}
	name, ok := entry["name"].(string)
	if !ok {
		return checker.Target{}, errors.New("No name field found")
	}
	url, ok := entry["url"].(string)
	if !ok {
		return checker.Target{}, errors.New("No url found")
	}
	return checker.Target{Name: name, URL: url}, nil
}

func runRound(ctx context.Context, targets []checker.Target, workers int) bool {
	slog.Info("---- round started ----")
	results := checker.RunAll(ctx, targets, workers)
	if ctx.Err() != nil {
		return false
	}
	unhealthyTargets := false
	for _, result := range results {
		if result.Err != nil {
			slog.Error("endpoint unreachable", result.Name, result.URL, "duration", result.Duration, "err", result.Err)
		} else if !result.Success {
			slog.Warn("endpoint unhealthy", result.Name, result.URL, "status", result.Status, "duration", result.Duration)
		} else {
			slog.Info("endpoint healthy", result.Name, result.URL, "status", result.Status, "duration", result.Duration)
		}
		if !result.Success {
			unhealthyTargets = true
		}
	}
	return unhealthyTargets
}

func runCheck(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true
	configErrors := false
	fmt.Printf("Loading from Directory: %s\n", dirPath)
	loaded, err := config.LoadAll(dirPath)
	if err != nil {
		slog.Error("Error Loading from Directory %s:", "path", dirPath)
		return err
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
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	cmd.SilenceUsage = true
	configErrors := false
	unhealthy := false
	if workers < 1 {
		return fmt.Errorf("workers must be at least 1, got %d", workers)
	}
	if interval < 0 {
		return fmt.Errorf("negative interval is not accepted, got %v", interval)
	}
	loaded, err := config.LoadAll(dirPath)
	if err != nil {
		slog.Error("Error Loading from Directory %s:", "path", dirPath)
		return err
	}
	var targets []checker.Target
	for _, lf := range loaded {
		err := lf.Err
		if err != nil {
			fmt.Println("ERROR:", err)
			configErrors = true
			continue
		}
		endpoints, ok := lf.Data["endpoints"].([]interface{})
		if ok {
			for i, item := range endpoints {
				entry, ok := item.(map[string]interface{})
				if !ok {
					slog.Warn("invalid entry from %d file: %s\n", "item", i, "path", lf.Path)
					configErrors = true
					continue
				}
				target, err := targetFromEntry(entry)
				if err != nil {
					slog.Warn("\nSkipping entry in %d file %s: %v", "item", i, "path", lf.Path, "err", err)
					configErrors = true
					continue
				}
				targets = append(targets, target)
			}
		} else {
			if lf.Data["type"] != "http" {
				continue
			}
			target, err := targetFromEntry(lf.Data)
			if err != nil {
				fmt.Printf("\nSkipping entry of %s: %v", lf.Path, err)
				configErrors = true
				continue
			}
			targets = append(targets, target)
		}
	}
	if len(targets) == 0 {
		return fmt.Errorf("nothing to monitor, found %d targets", len(targets))
	}
	slog.Info("monitor started", "targets", len(targets))
	unhealthy = runRound(ctx, targets, workers)
	if ctx.Err() != nil {
		return nil
	}

	if interval == 0 {
		if configErrors && unhealthy {
			return fmt.Errorf("config errors found, and one or more endpoints are unhealthy")
		} else if configErrors {
			return fmt.Errorf("one or more config files could not be parsed")
		} else if unhealthy {
			return fmt.Errorf("one or more endpoints failed health check")
		}
		return nil
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.Info("shutting down...\n")
			return nil
		case <-ticker.C:
			start := time.Now()
			runRound(ctx, targets, workers)
			slog.Info("---- round finished ----")
			elapsed := time.Since(start)
			if elapsed > interval {
				slog.Warn("warning: round took %v, longer than --interval %v\n", "time_elapsed", elapsed, "interval", interval)
			}
		}
	}
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
	monitorCmd.Flags().IntVarP(&workers, "workers", "w", 10, "Number of workers to result targets")
	monitorCmd.Flags().DurationVarP(&interval, "interval", "i", 0*time.Minute, "Time interval to ping taregsts")
	if err := checkCmd.MarkFlagRequired("dir"); err != nil {
		fmt.Println("Error setting up required field", err)
		os.Exit(1)
	}

	rootCmd.AddCommand(checkCmd)
	rootCmd.AddCommand(monitorCmd)
}
