package vcs

import "testing"

func TestVersionUsesBuildValue(t *testing.T) {
	original := buildVersion
	t.Cleanup(func() {
		buildVersion = original
	})

	buildVersion = "0.5.0"

	if got := Version(); got != buildVersion {
		t.Fatalf("Version() = %q, want %q", got, buildVersion)
	}
}

func TestRevisionUsesBuildValue(t *testing.T) {
	original := buildRevision
	t.Cleanup(func() {
		buildRevision = original
	})

	buildRevision = "0123456789abcdef"

	if got := Revision(); got != buildRevision {
		t.Fatalf("Revision() = %q, want %q", got, buildRevision)
	}
}
