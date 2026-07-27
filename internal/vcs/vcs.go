package vcs

import (
	"runtime/debug"
)

var buildVersion string

// returns application version set during build from release tag
func Version() string {
	if buildVersion != "" {
		return buildVersion
	}

	bi, ok := debug.ReadBuildInfo()
	if ok {
		return bi.Main.Version
	}

	return ""
}
