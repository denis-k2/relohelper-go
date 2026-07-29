package vcs

import (
	"runtime/debug"
)

var (
	buildVersion  string
	buildRevision string
)

// Version returns the application version set at build time.
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

// Revision returns the source control revision used to build the application.
func Revision() string {
	if buildRevision != "" {
		return buildRevision
	}

	bi, ok := debug.ReadBuildInfo()
	if !ok {
		return ""
	}

	for _, setting := range bi.Settings {
		if setting.Key == "vcs.revision" {
			return setting.Value
		}
	}

	return ""
}
