package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/phanivarun05/yamlchecker/internal/update"
	"github.com/spf13/cobra"
)

var logLevel string
var logFormat string

var rootCmd = &cobra.Command{
	SilenceErrors:     true,
	Use:               "yamlchecker",
	Short:             "Validates YAML configs against required fields",
	PersistentPreRunE: setupLogging,
	Version:           resolveVersion(),
}

func init() {
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "info", "pass the log level")
	rootCmd.PersistentFlags().StringVar(&logFormat, "log-format", "text", "pass the logging format")
}

func setupLogging(cmd *cobra.Command, args []string) error {
	var parsedLevel slog.Level
	if err := parsedLevel.UnmarshalText([]byte(logLevel)); err != nil {
		return fmt.Errorf("unknown log level %s, defaulting to Info\n", logLevel)
	}
	options := &slog.HandlerOptions{
		Level: parsedLevel,
	}
	var handler slog.Handler
	switch logFormat {
	case "json":
		handler = slog.NewJSONHandler(os.Stderr, options)
	case "text":
		handler = slog.NewTextHandler(os.Stderr, options)
	default:
		return fmt.Errorf("invalid log format %s", logFormat)
	}

	slog.SetDefault(slog.New(handler))
	return nil
}

func Execute() error {
	version := resolveVersion()

	var notice <-chan update.Release
	if update.Eligible(version) {
		notice = startUpdateCheck(version)
	}

	err := rootCmd.Execute()

	if notice != nil {
		select {
		case rel, ok := <-notice:
			if ok {
				fmt.Fprintf(os.Stderr,
					"\nA new version of yamlchecker is available: %s -> %s\n%s\n",
					version, rel.Tag, rel.URL)
			}
		case <-time.After(time.Second):
			// the check was too slow; skip it silently
		}
	}
	return err
}

func startUpdateCheck(current string) <-chan update.Release {
	ch := make(chan update.Release, 1)
	go func() {
		defer close(ch)

		cacheFile, err := update.DefaultCacheFile()
		if err != nil {
			slog.Debug("update check skipped", "err", err)
			return
		}
		checker := &update.Checker{
			Client:    &http.Client{Timeout: 3 * time.Second},
			URL:       update.LatestReleaseURL,
			CacheFile: cacheFile,
		}

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		rel, err := checker.Latest(ctx)
		if err != nil {
			slog.Debug("update check failed", "err", err)
			return
		}
		if update.IsNewer(current, rel.Tag) {
			ch <- rel
		}
	}()
	return ch
}
