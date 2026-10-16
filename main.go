package main

import (
	"log/slog"
	"os"
	"yamlchecker-cli/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		slog.Error("command failed", "err", err)
		os.Exit(1)
	}
}
