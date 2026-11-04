package main

import (
	"log/slog"
	"os"

	_ "golang.org/x/mod/semver"

	"github.com/phanivarun05/yamlchecker/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		slog.Error("command failed", "err", err)
		os.Exit(1)
	}
}
