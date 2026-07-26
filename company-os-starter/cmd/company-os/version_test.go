package main

// R-6.8: `company-os --version` reports the version AND the build identifier.
//
// The test asserts the shape rather than a literal string, because the literal
// changes with every commit. What must not rot is that all four facts reach the
// line and that an unstamped build says something truthful instead of leaving
// holes — the failure mode here is silent (a renamed linker symbol still links,
// it just stops stamping) and only the content check catches it.

import (
	"bytes"
	"regexp"
	"strings"
	"testing"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/model"
)

// versionShape is the published form:
//
//	company-os <version> (commit <sha>, go<x.y.z>, <goos>/<goarch>)
var versionShape = regexp.MustCompile(
	`^company-os \S+ \(commit \S+, go\S+, [a-z0-9]+/[a-z0-9]+\)\n$`)

func TestVersionFlagReportsVersionAndBuildID(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := run([]string{"--version"}, &stdout, &stderr); code != 0 {
		t.Fatalf("--version exited %d, want 0 (stderr: %s)", code, stderr.String())
	}
	if stderr.Len() != 0 {
		t.Errorf("--version wrote to stderr: %q", stderr.String())
	}
	got := stdout.String()
	if !versionShape.MatchString(got) {
		t.Fatalf("--version printed %q, want %v", got, versionShape)
	}

	// The four facts have to come from BuildInfo, not from a second derivation
	// that can drift away from the one --json will use (R-3.5).
	b := model.BuildInfo()
	for _, want := range []string{b.Version, b.Commit, b.GoVersion, b.Platform} {
		if !strings.Contains(got, want) {
			t.Errorf("--version output %q omits %q from model.BuildInfo()", got, want)
		}
	}
}

// TestBuildInfoNeverEmpty covers `go build` with no -ldflags, and the sharper
// case of `make build VERSION=` stamping an empty string over a good default.
// Both reach BuildInfo as "", and R-6.8 is not satisfied by a blank.
func TestBuildInfoNeverEmpty(t *testing.T) {
	b := model.BuildInfo()
	for name, v := range map[string]string{
		"Version":   b.Version,
		"Commit":    b.Commit,
		"GoVersion": b.GoVersion,
		"Platform":  b.Platform,
	} {
		if strings.TrimSpace(v) == "" {
			t.Errorf("model.BuildInfo().%s is empty", name)
		}
	}
	// `go test` links without the Makefile's -X flags, so this run is the
	// unstamped case and the fallbacks are what must be showing.
	if b.Version != "dev" || b.Commit != "unknown" {
		t.Errorf("unstamped build reported %q/%q, want the dev/unknown fallbacks",
			b.Version, b.Commit)
	}
}
