package main

import (
	"runtime/debug"
	"testing"
)

func TestResolvedVersionPrefersInjectedVersion(t *testing.T) {
	restoreVersionResolver(t, "1.2.3", debug.BuildInfo{
		Main: debug.Module{Version: "v9.9.9"},
	}, true)

	if got, want := resolvedVersion(), "1.2.3"; got != want {
		t.Fatalf("resolvedVersion() = %q, want %q", got, want)
	}
}

func TestResolvedVersionUsesModuleVersion(t *testing.T) {
	restoreVersionResolver(t, defaultVersion, debug.BuildInfo{
		Main: debug.Module{Version: "v3.5.1"},
	}, true)

	if got, want := resolvedVersion(), "3.5.1"; got != want {
		t.Fatalf("resolvedVersion() = %q, want %q", got, want)
	}
}

func TestResolvedVersionFallsBackToDev(t *testing.T) {
	restoreVersionResolver(t, defaultVersion, debug.BuildInfo{
		Main: debug.Module{Version: "(devel)"},
	}, true)

	if got, want := resolvedVersion(), defaultVersion; got != want {
		t.Fatalf("resolvedVersion() = %q, want %q", got, want)
	}
}

func TestResolvedVersionFallsBackToDevWithoutBuildInfo(t *testing.T) {
	restoreVersionResolver(t, defaultVersion, debug.BuildInfo{}, false)

	if got, want := resolvedVersion(), defaultVersion; got != want {
		t.Fatalf("resolvedVersion() = %q, want %q", got, want)
	}
}

func restoreVersionResolver(t *testing.T, testVersion string, buildInfo debug.BuildInfo, ok bool) {
	t.Helper()

	previousVersion := version
	previousReadBuildInfo := readBuildInfo
	version = testVersion
	readBuildInfo = func() (*debug.BuildInfo, bool) {
		return &buildInfo, ok
	}

	t.Cleanup(func() {
		version = previousVersion
		readBuildInfo = previousReadBuildInfo
	})
}
