package skills_test

// Starter-kit skill conformance (docs/lld/starter-kit-skills.md): every skill
// shipped under company-os-starter/skills/ must be discoverable and gate-4
// clean when copied, flat and unchanged, into a workspace's company-os/skills/.
// These files are read by nothing else, so this test is the entire safety net
// against the frontmatter drifting back — no `type:`, hand-written tags — the
// defect the LLD repairs.

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/model"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/skills"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/validate"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/workspace"
)

// starterKitSkillsDir is the shipped skills directory, resolved from the
// package's own location so the test works from any checkout layout.
func starterKitSkillsDir(t *testing.T) string {
	t.Helper()
	dir, err := filepath.Abs(filepath.Join("..", "..", "skills"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(dir); err != nil {
		t.Skipf("shipped skills directory unavailable: %v", err)
	}
	return dir
}

func TestStarterKitSkillsDiscoverAndPassGate4(t *testing.T) {
	dir := starterKitSkillsDir(t)
	entries, err := fs.Glob(os.DirFS(dir), "*.SKILL.md")
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) == 0 {
		t.Fatalf("no shipped skills found in %s", dir)
	}

	root := t.TempDir()
	dest := filepath.Join(root, "company-os", "skills")
	if err := os.MkdirAll(dest, 0o777); err != nil {
		t.Fatal(err)
	}
	for _, name := range entries {
		data, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dest, name), data, 0o666); err != nil {
			t.Fatal(err)
		}
	}

	ws := workspace.New(root)

	// Discovery: every shipped file is found, flat, at the company layer.
	got, err := skills.Discover(ws)
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}
	if len(got) != len(entries) {
		t.Fatalf("discovered %d skills, want %d", len(got), len(entries))
	}

	assertSkillsGate4Clean(t, ws)
}

// assertSkillsGate4Clean runs the full gate set and asserts that every skill
// under company-os/skills/ drew an ok and no core-field or tag-drift failure.
//
// The ok is asserted as well as the absence of failures: gate 4 must be
// WALKING the files for their silence to mean anything, and a traversal that
// stopped visiting skills would otherwise turn this green by looking away.
func assertSkillsGate4Clean(t *testing.T, ws *workspace.Workspace) {
	t.Helper()
	dir := filepath.Join(ws.Root, "company-os", "skills")
	names, err := fs.Glob(os.DirFS(dir), "*"+skills.Suffix)
	if err != nil {
		t.Fatal(err)
	}
	if len(names) == 0 {
		t.Fatalf("no skills at %s to check", dir)
	}

	sections, err := validate.Run(ws)
	if err != nil {
		t.Fatalf("validate.Run: %v", err)
	}
	want := make(map[string]bool, len(names))
	for _, n := range names {
		want[filepath.ToSlash(filepath.Join("company-os", "skills", n))] = true
	}
	sawOK := make(map[string]bool, len(names))
	for _, s := range sections {
		for _, f := range s.Findings {
			if !want[f.Path] {
				continue
			}
			switch f.Code {
			case model.CodeFrontmatterInSync:
				sawOK[f.Path] = true
			case model.CodeFrontmatterCoreField, model.CodeTagsDrift:
				t.Errorf("%s failed gate 4: [%s] %s", f.Path, f.Code, f.Message)
			}
		}
	}
	for rel := range want {
		if !sawOK[rel] {
			t.Errorf("gate 4 reported no ok for %s — is it being walked?", rel)
		}
	}
}
