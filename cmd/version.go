package cmd

import "fmt"

// Version information (set by main.go build flags)
var (
	Version   = "dev"
	BuildTime = "unknown"
	Commit    = "unknown"
)

// GetFullVersion returns the full version string with build info
func GetFullVersion() string {
	version := Version
	if BuildTime != "unknown" {
		version += fmt.Sprintf("\nBuilt at: %s", BuildTime)
	}
	if Commit != "unknown" {
		version += fmt.Sprintf("\nCommit: %s", Commit)
	}
	return version
}
