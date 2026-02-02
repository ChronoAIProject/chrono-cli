package main

import (
	cmd "github.com/aelfdevops/chrono/cmd"
)

// Version information set by build flags (e.g., -ldflags "-X main.version=1.0.0")
var (
	version   = "dev"
	buildTime = "unknown"
	commit    = "unknown"
)

func init() {
	// Set version info for display
	cmd.Version = version
	cmd.BuildTime = buildTime
	cmd.Commit = commit
}

func main() {
	cmd.Execute()
}
