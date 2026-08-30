package main

// Tests for `validate --fix` (ux-simplification 2.1).
//
// The four acceptance paths the task requires:
//
//   a) --fix on a drifted fixture copy regenerates the drifted artifacts and
//      then gates pass; a second --fix reports 0 regenerated (idempotent).
//   b) default validate on the same drifted copy still FAILs the same gates
//      (no accidental self-healing without the flag).
//   c) --fix does not rewrite anything under a synced slice (federated
//      fixture copy with a knowledge/ directory).
//   d) N counts only files whose bytes changed.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// copyFixtureTree is a recursive directory copy that preserves file content
// but not permissions (the test fixtures are read-only in places and the
// copy needs to be writable for --fix to regenerate).
func copyFixtureTree(t *testing.T, src, dst string) {
	t.Helper()
	if err := filepath.WalkDir(src, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o777)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o666)
	}); err != nil {
		t.Fatal(err)
	}
}

// TestValidateFixRegeneratesAndIsIdempotent is path (a): drift a tag and
// delete a per-directory index, run --fix, confirm the gates pass and the
// count is >0, then run --fix again and confirm the count is 0.
func TestValidateFixRegeneratesAndIsIdempotent(t *testing.T) {
	ws := filepath.Join(t.TempDir(), "ws")
	copyFixtureTree(t, fixturePath("workspace"), ws)

	// Drift a tag.
	brief := filepath.Join(ws, "teams", "customer-engagement", "product",
		"discovery", "2026-per-channel-quiet-hours", "brief.md")
	data, err := os.ReadFile(brief)
	if err != nil {
		t.Fatal(err)
	}
	drifted := strings.Replace(string(data),
		"tags: [kind/discovery, status/validated, team/customer-engagement]",
		"tags: [kind/discovery, status/validated, team/WRONG]", 1)
	if drifted == string(data) {
		t.Fatal("tag replacement did not match")
	}
	if err := os.WriteFile(brief, []byte(drifted), 0o666); err != nil {
		t.Fatal(err)
	}

	// Delete a per-directory index.
	idx := filepath.Join(ws, "teams", "customer-engagement", "standards", "index.md")
	if err := os.Remove(idx); err != nil {
		t.Fatal(err)
	}

	// First --fix: should pass and report N > 0.
	code, out, _ := runArgs(t, "--root", ws, "validate", "--fix")
	if code != 0 {
		t.Fatalf("first --fix exit %d:\n%s", code, out)
	}
	if !strings.Contains(out, "PASS") {
		t.Fatalf("first --fix did not PASS:\n%s", out)
	}
	if !strings.Contains(out, "validate --fix: 2 file(s) regenerated") {
		t.Fatalf("first --fix did not report 2 regenerated:\n%s", out)
	}

	// Second --fix: should pass and report 0 regenerated (idempotent).
	code, out, _ = runArgs(t, "--root", ws, "validate", "--fix")
	if code != 0 {
		t.Fatalf("second --fix exit %d:\n%s", code, out)
	}
	if !strings.Contains(out, "PASS") {
		t.Fatalf("second --fix did not PASS:\n%s", out)
	}
	if !strings.Contains(out, "validate --fix: 0 file(s) regenerated") {
		t.Fatalf("second --fix did not report 0 regenerated:\n%s", out)
	}
}

// TestValidateFixDefaultDoesNotHeal is path (b): the same drifted copy,
// without --fix, still fails the same gates.
func TestValidateFixDefaultDoesNotHeal(t *testing.T) {
	ws := filepath.Join(t.TempDir(), "ws")
	copyFixtureTree(t, fixturePath("workspace"), ws)

	// Drift a tag.
	brief := filepath.Join(ws, "teams", "customer-engagement", "product",
		"discovery", "2026-per-channel-quiet-hours", "brief.md")
	data, err := os.ReadFile(brief)
	if err != nil {
		t.Fatal(err)
	}
	drifted := strings.Replace(string(data),
		"tags: [kind/discovery, status/validated, team/customer-engagement]",
		"tags: [kind/discovery, status/validated, team/WRONG]", 1)
	if err := os.WriteFile(brief, []byte(drifted), 0o666); err != nil {
		t.Fatal(err)
	}

	// Default validate: should fail with the tag drift.
	code, out, _ := runArgs(t, "--root", ws, "validate")
	if code != 1 {
		t.Fatalf("expected exit 1, got %d:\n%s", code, out)
	}
	if !strings.Contains(out, "FAIL") {
		t.Fatalf("expected FAIL in output:\n%s", out)
	}
	if !strings.Contains(out, "committed tags drifted") {
		t.Fatalf("expected tag drift finding:\n%s", out)
	}
	// The fix summary line must NOT appear without --fix.
	if strings.Contains(out, "validate --fix:") {
		t.Fatalf("fix summary appeared without --fix:\n%s", out)
	}
}

// TestValidateFixDoesNotTouchSlices is path (c): --fix on a federated
// workspace does not rewrite anything under knowledge/.
func TestValidateFixDoesNotTouchSlices(t *testing.T) {
	ws := filepath.Join(t.TempDir(), "ws")
	copyFixtureTree(t, fixturePath("federated"), ws)

	// Snapshot the knowledge/ directory before --fix.
	snapBefore := snapshotDir(t, filepath.Join(ws, "knowledge"))

	// Run --fix. It may or may not pass (the federated fixture has its own
	// gate 9 checks), but it must not touch knowledge/.
	runArgs(t, "--root", ws, "validate", "--fix")

	snapAfter := snapshotDir(t, filepath.Join(ws, "knowledge"))
	if snapBefore != snapAfter {
		t.Fatal("--fix modified files under knowledge/ (synced slice)")
	}
}

// TestValidateFixCountsOnlyChangedBytes is path (d): N counts only files
// whose bytes actually changed. A clean workspace reports 0.
func TestValidateFixCountsOnlyChangedBytes(t *testing.T) {
	ws := filepath.Join(t.TempDir(), "ws")
	copyFixtureTree(t, fixturePath("workspace"), ws)

	code, out, _ := runArgs(t, "--root", ws, "validate", "--fix")
	if code != 0 {
		t.Fatalf("--fix on clean workspace exit %d:\n%s", code, out)
	}
	if !strings.Contains(out, "validate --fix: 0 file(s) regenerated") {
		t.Fatalf("clean workspace did not report 0 regenerated:\n%s", out)
	}
}

// TestValidateFixJSONCarriesCount verifies that --json output includes the
// fixRegenerated field when --fix is used.
func TestValidateFixJSONCarriesCount(t *testing.T) {
	ws := filepath.Join(t.TempDir(), "ws")
	copyFixtureTree(t, fixturePath("workspace"), ws)

	code, out, _ := runArgs(t, "--root", ws, "--json", "validate", "--fix")
	if code != 0 {
		t.Fatalf("--json --fix exit %d:\n%s", code, out)
	}
	if !strings.Contains(out, `"fixRegenerated": 0`) {
		t.Fatalf("JSON output missing fixRegenerated field:\n%s", out)
	}
}

// TestValidateDefaultJSONOmitsFixField verifies that --json output does NOT
// include the fixRegenerated field when --fix is not used.
func TestValidateDefaultJSONOmitsFixField(t *testing.T) {
	ws := filepath.Join(t.TempDir(), "ws")
	copyFixtureTree(t, fixturePath("workspace"), ws)

	code, out, _ := runArgs(t, "--root", ws, "--json", "validate")
	if code != 0 {
		t.Fatalf("--json validate exit %d:\n%s", code, out)
	}
	if strings.Contains(out, "fixRegenerated") {
		t.Fatalf("JSON output includes fixRegenerated without --fix:\n%s", out)
	}
}

// fixturePath resolves a path under examples/ from the test's working
// directory (company-os-starter/cmd/company-os/).
func fixturePath(parts ...string) string {
	args := append([]string{"..", "..", "..", "examples"}, parts...)
	return filepath.Join(args...)
}

// snapshotDir returns a content fingerprint of every file under dir, suitable
// for byte-identity comparison.
func snapshotDir(t *testing.T, dir string) string {
	t.Helper()
	var b strings.Builder
	if err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel, _ := filepath.Rel(dir, path)
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		b.WriteString(rel)
		b.WriteString(":")
		b.WriteString(string(data))
		b.WriteString("\n")
		return nil
	}); err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}
	return b.String()
}
