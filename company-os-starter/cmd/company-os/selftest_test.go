package main

// Inherited from examples/selftest.py (task 6.1) — the assertions that only mean
// something at the process boundary, where selftest.py drove them with
// subprocess.run(). run() is that boundary: argv in, exit code out, both streams
// captured.
//
// Covered here: ST-014, ST-015 (`:99-101`), ST-018..ST-025 (`:127-145`).
//
// Package note. The inventory files ST-018..ST-025 under internal/scaffold, and
// most of their halves are covered there already (see the ruling column). The
// two that are NOT — "a fresh workspace validates green" and "it still validates
// green after add + reality" — are the whole point of the group, and they need
// `validate` as well as the scaffolders. internal/scaffold must not import
// internal/validate to say so, so the pair lives here, one layer up, where
// selftest.py had them anyway.

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/model"
)

// runArgs is wantExit's plumbing without the assertion: it returns everything
// the caller might want to assert on.
func runArgs(t *testing.T, argv ...string) (code int, stdout, stderr string) {
	t.Helper()
	var out, errb bytes.Buffer
	code = run(argv, &out, &errb)
	return code, out.String(), errb.String()
}

// --------------------------------------------- root resolution (GPF-R-8.1)

// TestValidateOutsideRootExitsNonZero is selftest.py:99 (ST-014). Fail fast: a
// workspace-scoped command pointed at a directory that is not a workspace must
// refuse rather than validate an empty set of nothing and report success.
func TestValidateOutsideRootExitsNonZero(t *testing.T) {
	code, _, _ := runArgs(t, "--root", t.TempDir(), "validate")
	if code == 0 {
		t.Fatal("validate outside a workspace root exited 0")
	}
	if code != int(model.ExitWorkspace) {
		t.Errorf("exit code = %d, want %d (workspace)", code, model.ExitWorkspace)
	}
}

// TestValidateOutsideRootNamesResolutionOrder is selftest.py:100-101 (ST-015).
// The exit code alone tells a user nothing about which of three inputs the
// binary actually consulted. internal/workspace's TestRequireRootMessage pins
// the same sentence on the error VALUE; this pins that the value reaches stderr
// rather than being swallowed by a dispatch branch that only reads the code.
func TestValidateOutsideRootNamesResolutionOrder(t *testing.T) {
	const want = "--root -> $COMPANY_OS_WORKSPACE_ROOT -> current directory"
	_, stdout, stderr := runArgs(t, "--root", t.TempDir(), "validate")
	if !strings.Contains(stderr, want) {
		t.Errorf("stderr does not name the resolution order\nwant substring: %s\ngot stderr: %s\ngot stdout: %s",
			want, stderr, stdout)
	}
}

// -------------------------------------------------- scaffolding (GPF-R-1.x)

// TestInitHeadlessFlagsExitZero is selftest.py:127 (ST-018). Flags only, no
// TTY — the `go test` process has no terminal, so a prompt regression hangs or
// fails here rather than in someone's CI six weeks later (GPF-R-1.3).
func TestInitHeadlessFlagsExitZero(t *testing.T) {
	root := filepath.Join(t.TempDir(), "ws")
	code, _, stderr := runArgs(t, "--root", root, "init",
		"--company", "Acme", "--team", "core", "--platform", "payments")
	if code != 0 {
		t.Fatalf("init exited %d: %s", code, stderr)
	}
	if _, err := os.Stat(filepath.Join(root, "company-os", "company.yaml")); err != nil {
		t.Errorf("init exited 0 but scaffolded nothing: %v", err)
	}
}

// TestInitFreshWorkspaceValidatesGreen is selftest.py:128-129 (ST-019), GPF-R-1.7
// — the load-bearing one in this group. Every template, every seeded id, and
// every gate meet here: a scaffold that does not validate is a first-run
// experience where the tool immediately reports its own output as broken.
func TestInitFreshWorkspaceValidatesGreen(t *testing.T) {
	root := filepath.Join(t.TempDir(), "ws")
	if code, _, errb := runArgs(t, "--root", root, "init",
		"--company", "Acme", "--team", "core", "--platform", "payments"); code != 0 {
		t.Fatalf("init exited %d: %s", code, errb)
	}
	code, stdout, stderr := runArgs(t, "--root", root, "validate")
	if code != 0 {
		t.Fatalf("a freshly initialized workspace does not validate green (exit %d)\n%s\n%s",
			code, stdout, stderr)
	}
}

// TestInitRefusesReinit is selftest.py:131-132 (ST-020), GPF-R-1.2, at the
// dispatch level. internal/scaffold's TestInitRefusesInsideAWorkspace proves the
// refusal and the no-mutation half against a hand-made root; this proves the
// refusal survives a REAL prior init and that run() turns it into a non-zero
// exit rather than a printed warning.
func TestInitRefusesReinit(t *testing.T) {
	root := filepath.Join(t.TempDir(), "ws")
	if code, _, errb := runArgs(t, "--root", root, "init",
		"--company", "Acme", "--team", "core", "--platform", "payments"); code != 0 {
		t.Fatalf("first init exited %d: %s", code, errb)
	}
	before := treeOf(t, root)

	code, _, _ := runArgs(t, "--root", root, "init",
		"--company", "X", "--team", "y", "--platform", "z")
	if code == 0 {
		t.Fatal("re-init inside an existing workspace exited 0")
	}
	if got := treeOf(t, root); got != before {
		t.Error("the refused re-init mutated the workspace")
	}
}

// TestAddComponentExitZeroThenRefusesOverwrite is selftest.py:134-138
// (ST-021 + ST-022), GPF-R-1.9. internal/scaffold covers the duplicate refusal
// for `add platform`; the component path is a different branch (it also touches
// the platform's feature index and the id registry) and had no duplicate test.
func TestAddComponentExitZeroThenRefusesOverwrite(t *testing.T) {
	root := initFixture(t)

	code, _, errb := runArgs(t, "--root", root, "add", "component",
		"checkout-service", "--platform", "payments")
	if code != 0 {
		t.Fatalf("add component exited %d: %s", code, errb)
	}
	desc := filepath.Join(root, "platforms", "payments", "components", "checkout-service.yaml")
	if _, err := os.Stat(desc); err != nil {
		t.Fatalf("add component exited 0 but wrote no descriptor: %v", err)
	}
	before, err := os.ReadFile(desc)
	if err != nil {
		t.Fatal(err)
	}

	code, _, _ = runArgs(t, "--root", root, "add", "component",
		"checkout-service", "--platform", "payments")
	if code == 0 {
		t.Fatal("adding the same component twice exited 0 — the descriptor was overwritten")
	}
	after, err := os.ReadFile(desc)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, after) {
		t.Error("the refused add rewrote the existing descriptor")
	}
}

// TestRealityNewExitZeroThenRefusesOverwrite is selftest.py:139-142
// (ST-023 + ST-024), Unit 2, at the dispatch level.
func TestRealityNewExitZeroThenRefusesOverwrite(t *testing.T) {
	root := initFixture(t)
	if code, _, errb := runArgs(t, "--root", root, "add", "component",
		"checkout-service", "--platform", "payments"); code != 0 {
		t.Fatalf("add component exited %d: %s", code, errb)
	}

	code, _, errb := runArgs(t, "--root", root, "reality", "new",
		"checkout-service", "--platform", "payments")
	if code != 0 {
		t.Fatalf("reality new exited %d: %s", code, errb)
	}
	doc := filepath.Join(root, "platforms", "payments", "reality", "components", "checkout-service.md")
	if _, err := os.Stat(doc); err != nil {
		t.Fatalf("reality new exited 0 but wrote no document: %v", err)
	}

	if code, _, _ = runArgs(t, "--root", root, "reality", "new",
		"checkout-service", "--platform", "payments"); code == 0 {
		t.Fatal("re-scaffolding the same reality doc exited 0")
	}
}

// TestWorkspaceValidatesGreenAfterAddAndReality is selftest.py:143-145 (ST-025),
// GPF-R-1.7. Distinct from ST-019: init alone leaves no component, so a
// descriptor/ownership/reality reconciliation bug is invisible until something
// has been added. This is the assertion that the GROWN workspace, not just the
// empty one, stays green.
func TestWorkspaceValidatesGreenAfterAddAndReality(t *testing.T) {
	root := initFixture(t)
	for _, argv := range [][]string{
		{"--root", root, "add", "component", "checkout-service", "--platform", "payments"},
		{"--root", root, "reality", "new", "checkout-service", "--platform", "payments"},
	} {
		if code, _, errb := runArgs(t, argv...); code != 0 {
			t.Fatalf("%v exited %d: %s", argv, code, errb)
		}
	}
	code, stdout, stderr := runArgs(t, "--root", root, "validate")
	if code != 0 {
		t.Fatalf("the workspace stopped validating green after add + reality (exit %d)\n%s\n%s",
			code, stdout, stderr)
	}
}

// initFixture is selftest.py's shared `with tempfile.TemporaryDirectory()` block
// header (`:126-127`): one initialized workspace with a team and a platform.
func initFixture(t *testing.T) string {
	t.Helper()
	root := filepath.Join(t.TempDir(), "ws")
	if code, _, errb := runArgs(t, "--root", root, "init",
		"--company", "Acme", "--team", "core", "--platform", "payments"); code != 0 {
		t.Fatalf("init exited %d: %s", code, errb)
	}
	return root
}

// treeOf renders the relative path of every file under root, so "mutated
// nothing" can be asserted rather than assumed.
func treeOf(t *testing.T, root string) string {
	t.Helper()
	var b strings.Builder
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, relErr := filepath.Rel(root, path)
		if relErr != nil {
			return relErr
		}
		b.WriteString(rel)
		if !info.IsDir() {
			data, readErr := os.ReadFile(path)
			if readErr != nil {
				return readErr
			}
			b.WriteString("\x00")
			b.Write(data)
		}
		b.WriteString("\n")
		return nil
	})
	if err != nil {
		t.Fatalf("walking %s: %v", root, err)
	}
	return b.String()
}
