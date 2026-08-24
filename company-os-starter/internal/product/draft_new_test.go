package product

// `prd new --draft` (R-2.1 … R-2.10). The cases are table-driven per R-ID so a
// rule that stops holding names itself in the failure.

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/graph"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/model"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/workspace"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/yamlio"
)

// draftID is the id DraftNew derives for a title, so the tests assert against
// the same derivation the command uses rather than a hard-coded year.
func draftID(slug string) string {
	return strconv.Itoa(today().Year()) + "-" + slug
}

// readDraft returns the draft's frontmatter, its body, and its raw text.
func readDraft(t *testing.T, path string) (yamlio.PyMap, string, string) {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	meta, body, err := graph.ReadFrontmatter(path)
	if err != nil {
		t.Fatal(err)
	}
	return meta, string(body), string(raw)
}

// TestDraftNewFrontmatter is R-2.1, R-2.2, R-2.4, R-2.5 and R-2.10: one draft,
// one assertion per rule over its frontmatter.
func TestDraftNewFrontmatter(t *testing.T) {
	for _, tc := range []struct {
		name     string
		platform string
		// want is the frontmatter key -> exact value expected.
		want map[string]string
	}{
		{
			name:     "no platform leaves both platform fields placeholders",
			platform: "",
			want: map[string]string{
				"type":               "prd",
				"status":             "draft",
				"team":               "core",
				"title":              "TODO",
				"platform":           "TODO",
				"governanceSnapshot": "TODO",
				"decisionOwner":      "TODO",
			},
		},
		{
			// R-2.5.
			name:     "--platform sets promoteTo.platform and platform",
			platform: "payments",
			want: map[string]string{
				"type":     "prd",
				"status":   "draft",
				"team":     "core",
				"platform": "payments",
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ws := fixture(t)
			if _, err := DraftNew(ws, "core", tc.platform, "Faster refunds", "", nil); err != nil {
				t.Fatal(err)
			}
			id := draftID("faster-refunds")
			dir, err := DraftDir(ws, "core")
			if err != nil {
				t.Fatal(err)
			}
			path := DraftPath(dir, id)
			meta, _, raw := readDraft(t, path)

			// R-2.1: the derived id and today's date.
			if got := strOf(meta, "id"); got != id {
				t.Errorf("id = %q, want %q (R-2.1)", got, id)
			}
			if got := strOf(meta, "created"); got != today().Format(isoDate) {
				t.Errorf("created = %q, want today (R-2.1)", got)
			}
			for k, want := range tc.want {
				if got := strOf(meta, k); got != want {
					t.Errorf("%s = %q, want %q", k, got, want)
				}
			}
			// R-2.4 / R-2.5: the routing field.
			wantTarget := tc.platform
			if wantTarget == "" {
				wantTarget = "TODO"
			}
			pt, ok := meta.Get("promoteTo").(yamlio.PyMap)
			if !ok {
				t.Fatalf("promoteTo missing or not a mapping (R-2.4)")
			}
			if got := yamlio.PyString(pt.Get("platform")); got != wantTarget {
				t.Errorf("promoteTo.platform = %q, want %q (R-2.4/R-2.5)", got, wantTarget)
			}
			// R-2.4: components is a placeholder, not an invented list.
			if seq, ok := meta.Get("components").(yamlio.PySeq); !ok || len(seq) != 0 {
				t.Errorf("components = %v, want an empty list (R-2.4)", meta.Get("components"))
			}
			// R-2.2: `level` means rule tier elsewhere and must not appear.
			if meta.Get("level") != nil {
				t.Error("draft carries a 'level' key (R-2.2)")
			}
			// R-2.10: tags are derived by graph build, never authored.
			if meta.Get("tags") != nil {
				t.Error("draft carries a hand-written 'tags' key (R-2.10)")
			}
			// The frontmatter parser contract: `^---\n...\n---\n`.
			if !strings.HasPrefix(raw, "---\ntype: prd\n") {
				t.Errorf("draft does not open with the frontmatter fence:\n%.40q", raw)
			}
		})
	}
}

// TestDraftNewBodyFromTemplate is R-2.3: the body comes from ResolveTemplate
// under the existing team -> platform -> company -> built-in probe order, and
// the template's own frontmatter does not survive into the draft.
func TestDraftNewBodyFromTemplate(t *testing.T) {
	for _, tc := range []struct {
		name       string
		override   string // "" for none
		body       string
		wantSource string
		wantIn     string
	}{
		{
			name:       "built-in",
			wantSource: "built-in PRD_TEMPLATE",
			wantIn:     "## Problem statement",
		},
		{
			name:       "team override wins",
			override:   filepath.Join("teams", "core", "templates", "prd.md"),
			body:       "---\ntype: prd\nstatus: proposed\n---\n\n# team body {title}\n",
			wantSource: "teams/core/templates/prd.md",
			wantIn:     "# team body Faster refunds",
		},
		{
			name:       "company override when the team has none",
			override:   filepath.Join("company-os", "templates", "prd.md"),
			body:       "# company body\n",
			wantSource: "company-os/templates/prd.md",
			wantIn:     "# company body",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ws := fixture(t)
			if tc.override != "" {
				write(t, filepath.Join(ws.Root, tc.override), tc.body)
			}
			out, err := DraftNew(ws, "core", "", "Faster refunds", "", nil)
			if err != nil {
				t.Fatal(err)
			}
			dir, _ := DraftDir(ws, "core")
			_, body, raw := readDraft(t, DraftPath(dir, draftID("faster-refunds")))
			if !strings.Contains(body, tc.wantIn) {
				t.Errorf("body does not contain %q (R-2.3):\n%s", tc.wantIn, body)
			}
			// The override's own `status: proposed` block must not survive.
			if strings.Contains(raw, "status: proposed") {
				t.Errorf("template frontmatter leaked into the draft (R-2.1):\n%s", raw)
			}
			if got := fieldOf(out, model.CodeTemplateSource, "source"); got != tc.wantSource {
				t.Errorf("template source = %q, want %q (R-2.3)", got, tc.wantSource)
			}
		})
	}
}

// TestDraftNewRefusals is R-2.6 plus the two argument preconditions: every case
// refuses without writing, and the conflict leaves the existing draft alone.
func TestDraftNewRefusals(t *testing.T) {
	for _, tc := range []struct {
		name     string
		team     string
		platform string
		title    string
		discover string
		// seed writes a draft at this id before the run.
		seed     string
		wantCode model.ExitCode
		wantIn   string
	}{
		{
			name: "no team", team: "", title: "T",
			wantCode: model.ExitUsage, wantIn: "--team",
		},
		{
			name: "unknown team", team: "nope", title: "T",
			wantCode: model.ExitWorkspace,
		},
		{
			name: "unknown platform", team: "core", platform: "ghost", title: "T",
			wantCode: model.ExitWorkspace,
		},
		{
			name: "no title and no discovery", team: "core",
			wantCode: model.ExitUsage, wantIn: "--title required",
		},
		{
			// R-2.6.
			name: "existing derived id", team: "core", title: "Faster refunds",
			seed: draftID("faster-refunds"),
			// The refusal is a conflict, not an artifact or workspace fault.
			wantCode: model.ExitConflict, wantIn: "already exists",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ws := fixture(t)
			const sentinel = "---\ntype: prd\nid: seeded\nstatus: draft\n---\n\nmine\n"
			var seeded string
			if tc.seed != "" {
				dir, err := DraftDir(ws, "core")
				if err != nil {
					t.Fatal(err)
				}
				seeded = DraftPath(dir, tc.seed)
				write(t, seeded, sentinel)
			}
			out, err := DraftNew(ws, tc.team, tc.platform, tc.title, tc.discover, nil)
			if err == nil {
				t.Fatal("expected a refusal")
			}
			if out != nil {
				t.Errorf("a refused run returned records: %v", out)
			}
			if got := model.CodeOf(err); got != tc.wantCode {
				t.Errorf("exit = %d, want %d (%v)", got, tc.wantCode, err)
			}
			if tc.wantIn != "" && !strings.Contains(err.Error(), tc.wantIn) {
				t.Errorf("error = %q, want it to mention %q", err, tc.wantIn)
			}
			if seeded != "" {
				raw, readErr := os.ReadFile(seeded)
				if readErr != nil || string(raw) != sentinel {
					t.Error("the existing draft was overwritten (R-2.6)")
				}
			}
		})
	}
}

// TestDraftNewFromDiscovery is R-2.7: the brief must be validated, and its
// Problem and Success sections are copied forward.
func TestDraftNewFromDiscovery(t *testing.T) {
	brief := func(status string) string {
		return "---\ntype: discovery-brief\nid: 2026-refunds\ntitle: Refund latency\n" +
			"status: " + status + "\n---\n\n" +
			"# Discovery\n\n## Problem signal\nRefunds take four days.\n\n" +
			"## Success criteria\nUnder one day at p95.\n"
	}

	t.Run("draft copies the sections forward", func(t *testing.T) {
		ws := fixture(t)
		write(t, filepath.Join(ws.Root, "teams", "core", "product", "discovery",
			"2026-refunds", "brief.md"), brief("validated"))

		if _, err := DraftNew(ws, "core", "", "", "2026-refunds", nil); err != nil {
			t.Fatal(err)
		}
		dir, _ := DraftDir(ws, "core")
		// The title came from the brief, so the id derives from it too.
		_, body, _ := readDraft(t, DraftPath(dir, draftID("refund-latency")))
		for _, want := range []string{"Refunds take four days.", "Under one day at p95."} {
			if !strings.Contains(body, want) {
				t.Errorf("body is missing %q (R-2.7):\n%s", want, body)
			}
		}
	})

	t.Run("an unvalidated brief is refused", func(t *testing.T) {
		ws := fixture(t)
		write(t, filepath.Join(ws.Root, "teams", "core", "product", "discovery",
			"2026-refunds", "brief.md"), brief("draft"))

		_, err := DraftNew(ws, "core", "", "", "2026-refunds", nil)
		if err == nil {
			t.Fatal("expected a refusal (R-2.7)")
		}
		if got := model.CodeOf(err); got != model.ExitPrecondition {
			t.Errorf("exit = %d, want %d (%v)", got, model.ExitPrecondition, err)
		}
		dir, _ := DraftDir(ws, "core")
		if ids := DraftIDs(dir); ids != nil {
			t.Errorf("a refused run wrote drafts: %v (R-2.7)", ids)
		}
	})
}

// TestDraftNewRebuilds is R-2.8: the derived artifacts are rebuilt before the
// command returns, and the rebuild's records are spliced into the output.
func TestDraftNewRebuilds(t *testing.T) {
	ws := fixture(t)
	called := 0
	rebuild := Rebuild(func(*workspace.Workspace) ([]model.GateResult, error) {
		called++
		return []model.GateResult{{Ordinal: 9, Slug: model.SectionTags}}, nil
	})
	out, err := DraftNew(ws, "core", "", "Faster refunds", "", rebuild)
	if err != nil {
		t.Fatal(err)
	}
	if called != 1 {
		t.Errorf("rebuild called %d times, want 1 (R-2.8)", called)
	}
	// The rebuild's section sits between the creation lines and the closing
	// pointer, so the pointer is genuinely the last thing said.
	if len(out) != 3 || out[1].Slug != model.SectionTags {
		t.Fatalf("records = %v, want created / derived / next (R-2.8)", out)
	}
	if out[2].Findings[0].Code != model.CodePRDDraftNext {
		t.Errorf("last finding = %q, want the draft pointer", out[2].Findings[0].Code)
	}
}

// TestDraftNewGuidance is R-2.9: the closing line names the draft file, and
// names no promote command.
func TestDraftNewGuidance(t *testing.T) {
	ws := fixture(t)
	out, err := DraftNew(ws, "core", "", "Faster refunds", "", nil)
	if err != nil {
		t.Fatal(err)
	}
	var next model.Finding
	for _, s := range out {
		for _, f := range s.Findings {
			if f.Code == model.CodePRDDraftNext {
				next = f
			}
		}
	}
	if next.Code == "" {
		t.Fatal("no closing guidance line (R-2.9)")
	}
	rel := filepath.ToSlash(filepath.Join("teams", "core", "product",
		"change-records", "draft", draftID("faster-refunds"), "prd.md"))
	if !strings.Contains(next.Message, rel) {
		t.Errorf("guidance = %q, want it to name %q (R-2.9)", next.Message, rel)
	}
	// R-2.9: promotion is not the immediate next command.
	for _, s := range out {
		for _, f := range s.Findings {
			if strings.Contains(f.Message, "prd promote") {
				t.Errorf("guidance names `prd promote` (R-2.9): %q", f.Message)
			}
		}
	}
	// R-3.6 lifts next-commands out of Fields; a draft offers none, because the
	// next move is editing the file rather than running anything.
	if got := model.NextCommands(out); got != nil {
		t.Errorf("NextCommands = %v, want none (R-2.9)", got)
	}
}

// fieldOf returns the named Fields value of the first finding carrying code.
func fieldOf(sections []model.GateResult, code, key string) string {
	for _, s := range sections {
		for _, f := range s.Findings {
			if f.Code == code {
				return f.Fields.Str(key)
			}
		}
	}
	return ""
}
