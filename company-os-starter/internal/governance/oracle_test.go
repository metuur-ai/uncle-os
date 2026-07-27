package governance_test

// Shared fixture helpers for this package's tests.
//
// This file used to hold four differential tests — TestDeclareReproducesTheOracleBytes,
// TestRequestReproducesTheOracleBytes, TestResolveMatchesTheOracle and
// TestExplainMatchesTheOracle — which ran company-os-starter/bin/company-os over
// the committed fixtures and compared bytes at the library seam. R-9.3 deleted
// that binary, so all four could only SKIP: 22 subtests reporting green while
// asserting nothing, which is precisely what their own "a missing oracle must
// never look like agreement" comment was written to prevent.
//
// They were removed rather than frozen because internal/difftest covers the same
// ground end to end and more broadly: governance/resolve-* over all six committed
// workspaces plus the failure-path ones, governance/resolve-twice for
// idempotence, governance/explain-* including the close-match suggestion path,
// and deviation/* and exception/* for the read-modify-write these tests pinned —
// each freezing stdout, stderr, exit code AND a hash of every file in the tree,
// so a byte-level reflow regression still fails.
//
// What is lost is the claim "matches Python". What remains is "has not changed
// since a human reviewed it", which is all any test can assert once the
// reference implementation is gone.
//
// The helpers below outlived the tests: governance_test.go uses copyFixture,
// readFile, renderSections and repoRoot.

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/model"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/render"
)

func repoRoot(t *testing.T) string {
	t.Helper()
	abs, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatalf("resolving repo root: %v", err)
	}
	return abs
}

func copyFixture(t *testing.T, name string) string {
	t.Helper()
	src := filepath.Join(repoRoot(t), "examples", filepath.FromSlash(name))
	if _, err := os.Stat(src); err != nil {
		t.Skipf("fixture absent: %v", err)
	}
	dst := filepath.Join(t.TempDir(), "ws")
	err := filepath.WalkDir(src, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, p)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		if !d.Type().IsRegular() {
			return nil
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		return os.WriteFile(target, b, 0o644)
	})
	if err != nil {
		t.Fatalf("copying %s: %v", src, err)
	}
	return dst
}

func renderSections(t *testing.T, sections []model.GateResult) string {
	t.Helper()
	var b bytes.Buffer
	if err := render.Governance(&b, sections); err != nil {
		t.Fatalf("render.Governance: %v", err)
	}
	return b.String()
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading %s: %v", path, err)
	}
	return string(b)
}
