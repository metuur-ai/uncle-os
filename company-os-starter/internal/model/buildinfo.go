package model

import "runtime"

// version and commit are stamped at link time by the Makefile:
//
//	-X <module>/internal/model.version=$(VERSION)
//	-X <module>/internal/model.commit=$(COMMIT)
//
// They live here rather than in package main because R-3.5 puts the same
// identifier in every --json payload, and the JSON encoder is an internal
// package: if the vars stayed in main, the encoder would need main to hand them
// down, and two call sites would be free to disagree about what "the build" is.
// One linker-visible pair, one accessor, both consumers read it.
//
// They are deliberately empty rather than pre-set to their fallbacks. `-X` on an
// already-initialised var works, but so does `make build VERSION=`, which stamps
// an empty string over a good default and yields `company-os  (commit , ...)`.
// Normalising in BuildInfo covers the un-stamped build and the badly stamped one
// with the same branch.
var (
	version = ""
	commit  = ""
)

// Fallbacks for a `go build` with no -ldflags. R-6.8 has to say something
// truthful about an unstamped binary, and "" is not it.
const (
	fallbackVersion = "dev"
	fallbackCommit  = "unknown"
)

// Build is the running binary's self-description: what R-6.8 reports through
// --version and what R-3.5 requires in every --json payload.
//
// It carries no wording. R-2.8 keeps sentence composition above the dispatch
// seam, so this is fields only and cmd/company-os decides how the --version line
// reads. The json tags are the JSON contract and are frozen by R-3.4 once
// published.
//
// There is no build *date*. A timestamp would make two builds of the same source
// differ, which costs the reproducibility R-6.10's checksums are worth having
// for, and buys nothing Commit does not already say precisely.
type Build struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	GoVersion string `json:"goVersion"`
	Platform  string `json:"platform"`
}

// BuildInfo returns the identifier for the running binary. It is the single
// source for --version and for the `build` object in --json (R-3.5) — neither
// re-derives it.
func BuildInfo() Build {
	b := Build{
		Version:   version,
		Commit:    commit,
		GoVersion: runtime.Version(),
		Platform:  runtime.GOOS + "/" + runtime.GOARCH,
	}
	if b.Version == "" {
		b.Version = fallbackVersion
	}
	if b.Commit == "" {
		b.Commit = fallbackCommit
	}
	return b
}
