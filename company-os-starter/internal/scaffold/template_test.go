package scaffold

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/workspace"
	"github.com/metuur-ai/uncle-os/company-os-starter/templates"
)

// --- oracle tests: they read the repository, so they run only in a checkout ---

// TestEmbeddedRealityTemplateMatchesDisk is the point of the embed: the bytes
// compiled into the binary must be the bytes of templates/reality-component.md,
// which is what _builtin_template read from disk (bin/company-os:526-529).
func TestEmbeddedRealityTemplateMatchesDisk(t *testing.T) {
	onDisk, err := os.ReadFile(filepath.Join("..", "..", "templates", "reality-component.md"))
	if err != nil {
		t.Fatalf("reading the template from disk: %v", err)
	}
	if string(onDisk) != templates.RealityComponent {
		t.Errorf("embedded reality-component.md differs from disk\n"+
			"disk     %d bytes\nembedded %d bytes", len(onDisk), len(templates.RealityComponent))
	}
}

// TestBuiltinsMatchPythonModuleStrings pins the two hand-ported constants to the
// oracle, so a transcription slip fails here rather than in a golden diff.
//
// It used to read the strings out of bin/company-os directly. R-9.3 deleted that
// file, so the oracle is now a frozen copy extracted from the python-cli-final
// tag — the last tree that contained it. Regenerate with:
//
//	git show python-cli-final:company-os-starter/bin/company-os |
//	  python3 -c 'import re,sys; s=sys.stdin.read(); \
//	    print(re.search(r"(?s)\nPRD_TEMPLATE = \"\"\"(.*?)\"\"\"", s).group(1), end="")' \
//	  > internal/scaffold/testdata/prd-template.python.txt
//
// Editing a testdata file to make this pass is a deliberate act: it means the Go
// template has intentionally diverged from what shipped, and the artifacts every
// adopter's `discover new` / `prd new` produces will change shape.
func TestBuiltinsMatchPythonModuleStrings(t *testing.T) {
	for _, tc := range []struct {
		symbol string
		golden string
		got    string
	}{
		{"DISCOVERY_TEMPLATE", "discovery-template.python.txt", DiscoveryTemplate},
		{"PRD_TEMPLATE", "prd-template.python.txt", PRDTemplate},
	} {
		want, err := os.ReadFile(filepath.Join("testdata", tc.golden))
		if err != nil {
			t.Fatalf("reading the frozen %s oracle: %v", tc.symbol, err)
		}
		if string(want) != tc.got {
			t.Errorf("%s differs from the Python module string frozen at python-cli-final\n"+
				"python %d bytes\ngo     %d bytes", tc.symbol, len(want), len(tc.got))
		}
	}
}

// --- self-contained tests: no repository file is read, so these also pass when
// the test binary is run alone in an empty directory (R-6.7) ---

// writeTemplate creates an override at dir/templates/<name>.md with a body that
// identifies which override it is.
func writeTemplate(t *testing.T, dir, name, body string) {
	t.Helper()
	full := filepath.Join(dir, "templates", name+".md")
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// TestResolveTemplateOverridePrecedence walks the chain top-down, deleting the
// winner each time, and asserts both the body and the printed label.
func TestResolveTemplateOverridePrecedence(t *testing.T) {
	ws := workspace.New(t.TempDir())
	teamDir := filepath.Join(ws.Teams, "engineering")
	platformDir := filepath.Join(ws.Platforms, "core")

	writeTemplate(t, teamDir, TemplatePRD, "team body\n")
	writeTemplate(t, platformDir, TemplatePRD, "platform body\n")
	writeTemplate(t, ws.Company, TemplatePRD, "company body\n")

	steps := []struct {
		remove     string
		wantBody   string
		wantSource string
	}{
		{"", "team body\n", "teams/engineering/templates/prd.md"},
		{filepath.Join(teamDir, "templates", "prd.md"),
			"platform body\n", "platforms/core/templates/prd.md"},
		{filepath.Join(platformDir, "templates", "prd.md"),
			"company body\n", "company-os/templates/prd.md"},
		{filepath.Join(ws.Company, "templates", "prd.md"),
			PRDTemplate, SourceBuiltinPRD},
	}
	for _, step := range steps {
		if step.remove != "" {
			if err := os.Remove(step.remove); err != nil {
				t.Fatal(err)
			}
		}
		body, source, err := ResolveTemplate(ws, TemplatePRD, "engineering", "core")
		if err != nil {
			t.Fatalf("ResolveTemplate: %v", err)
		}
		if source != step.wantSource {
			t.Errorf("source = %q, want %q", source, step.wantSource)
		}
		if body != step.wantBody {
			t.Errorf("body for %s = %q, want %q", step.wantSource, body, step.wantBody)
		}
	}
}

// TestResolveTemplateSkipsUnsuppliedScopes mirrors Python's `if team:` and
// `if platform:` guards: an override that exists is invisible when its scope was
// not passed, which is why `discover new` never picks up a platform override.
func TestResolveTemplateSkipsUnsuppliedScopes(t *testing.T) {
	ws := workspace.New(t.TempDir())
	writeTemplate(t, filepath.Join(ws.Teams, "engineering"), TemplatePRD, "team body\n")
	writeTemplate(t, filepath.Join(ws.Platforms, "core"), TemplatePRD, "platform body\n")

	cases := []struct {
		name              string
		team, platform    string
		wantBody, wantSrc string
	}{
		{"team only", "engineering", "", "team body\n", "teams/engineering/templates/prd.md"},
		{"platform only", "", "core", "platform body\n", "platforms/core/templates/prd.md"},
		{"neither", "", "", PRDTemplate, SourceBuiltinPRD},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			body, source, err := ResolveTemplate(ws, TemplatePRD, tc.team, tc.platform)
			if err != nil {
				t.Fatalf("ResolveTemplate: %v", err)
			}
			if source != tc.wantSrc || body != tc.wantBody {
				t.Errorf("got (%q, %q), want (%q, %q)", body, source, tc.wantBody, tc.wantSrc)
			}
		})
	}
}

// TestResolveTemplateBuiltinsWithoutWorkspace is the R-6.7 assertion: with no
// override anywhere and nothing on disk beside the binary, all three built-ins
// still resolve, and each carries the label the oracle prints.
func TestResolveTemplateBuiltinsWithoutWorkspace(t *testing.T) {
	ws := workspace.New(t.TempDir())
	cases := []struct {
		name     string
		wantBody string
		wantSrc  string
	}{
		{TemplateDiscoveryBrief, DiscoveryTemplate, SourceBuiltinDiscovery},
		{TemplatePRD, PRDTemplate, SourceBuiltinPRD},
		{TemplateRealityComponent, templates.RealityComponent, SourceBuiltinReality},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			body, source, err := ResolveTemplate(ws, tc.name, "engineering", "core")
			if err != nil {
				t.Fatalf("ResolveTemplate: %v", err)
			}
			if source != tc.wantSrc {
				t.Errorf("source = %q, want %q", source, tc.wantSrc)
			}
			if body != tc.wantBody {
				t.Errorf("body = %q, want %q", body, tc.wantBody)
			}
			if body == "" {
				t.Error("built-in body is empty")
			}
		})
	}
}

// TestEmbeddedRealityTemplateIsIntact checks the embedded bytes without reading
// the repository, so the check survives being run from an empty directory.
func TestEmbeddedRealityTemplateIsIntact(t *testing.T) {
	const wantPrefix = "---\nid: reality-<component-id>\ntype: component-reality\n"
	if !strings.HasPrefix(templates.RealityComponent, wantPrefix) {
		t.Errorf("embedded reality template does not start with the expected frontmatter, got %q",
			templates.RealityComponent)
	}
	for _, want := range []string{"## Business rules", "## Current limitations", "<Component Name>"} {
		if !strings.Contains(templates.RealityComponent, want) {
			t.Errorf("embedded reality template is missing %q", want)
		}
	}
}

// TestResolveTemplateUnknownName covers _builtin_template's (None, None) tail.
func TestResolveTemplateUnknownName(t *testing.T) {
	ws := workspace.New(t.TempDir())
	_, _, err := ResolveTemplate(ws, "no-such-template", "", "")
	if !errors.Is(err, ErrUnknownTemplate) {
		t.Fatalf("err = %v, want ErrUnknownTemplate", err)
	}
}
