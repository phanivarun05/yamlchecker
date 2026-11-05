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
	var updater *update.Checker
	var refreshed <-chan struct{}

	if update.Eligible(version) {
		updater = newUpdateChecker()
	}

	if updater != nil && updater.Stale() {
		refreshed = refreshInBackground(updater)
	}

	err := rootCmd.Execute()

	if updater != nil && refreshed != nil {
		<-refreshed
	}

	if err != nil {
		slog.Error("command failed:", "err", err)
	}

	// Print the update notice last with styling
	if updater != nil {
		if rel, ok := updater.Cached(); ok && update.IsNewer(version, rel.Tag) {
			// Define ANSI color sequences
			const (
				colorYellow = "\033[33m"
				styleBold   = "\033[1m"
				colorReset  = "\033[0m"
			)

			// Wrap the notice text in Bold + Yellow, then clear it with colorReset
			fmt.Fprintf(os.Stderr, "\n%s%sA new version of yamlchecker is available: %s -> %s\n%s%s\n",
				styleBold, colorYellow, version, rel.Tag, rel.URL, colorReset)
		}
	}

	if err != nil {
		os.Exit(1)
	}

	return nil
}

func newUpdateChecker() *update.Checker {
	cacheFile, err := update.DefaultCacheFile()
	if err != nil {
		return nil
	}
	return &update.Checker{
		Client:    &http.Client{Timeout: 2 * time.Second},
		URL:       update.LatestReleaseURL,
		CacheFile: cacheFile,
	}
}

func refreshInBackground(c *update.Checker) <-chan struct{} {
	done := make(chan struct{})
	go func() {
		defer close(done)
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if _, err := c.Latest(ctx); err != nil {
			slog.Debug("update check failed", "err", err)
		}
	}()
	return done
}
