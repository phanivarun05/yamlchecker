package cmd

import "runtime/debug"

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func resolveVersion() string {
	if version != "dev" {
		return version
	}
	info, ok := debug.ReadBuildInfo()
	if ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}
	return version
}
