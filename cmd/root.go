package cmd

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
)

var logLevel string
var logFormat string

var rootCmd = &cobra.Command{
	Use:               "yamlchecker",
	Short:             "Validates YAML configs against required fields",
	PersistentPreRunE: setupLogging,
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
	return rootCmd.Execute()
}
