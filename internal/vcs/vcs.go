package vcs

import (
	"runtime/debug"
)

// returns application version set during build from release tag
func Version() string {
	bi, ok := debug.ReadBuildInfo()
	if ok {
		return bi.Main.Version
	}

	return ""
}
