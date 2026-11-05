package main

import (
	"os"

	_ "golang.org/x/mod/semver"

	"github.com/phanivarun05/yamlchecker/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
