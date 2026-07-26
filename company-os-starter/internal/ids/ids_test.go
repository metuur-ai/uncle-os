package ids_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/ids"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/model"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/workspace"
)

// exampleWorkspace is examples/workspace, the fixture selftest.py drove.
func exampleWorkspace(t *testing.T) *workspace.Workspace {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", "..", "..", "examples", "workspace"))
	if err != nil {
		t.Fatalf("resolving fixture: %v", err)
	}
	if _, err := os.Stat(root); err != nil {
		t.Skipf("fixture unavailable: %v", err)
	}
	return workspace.New(root)
}

// TestLoadRegistry_ReadsRegistryEntries ports selftest.py ST-026 (:150):
// canonical IDs come from ids/registry.yaml and from no parallel index.
func TestLoadRegistry_ReadsRegistryEntries(t *testing.T) {
	entries, err := ids.Load(exampleWorkspace(t))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	got := map[string]string{}
	for _, e := range entries {
		got[e.ID] = e.DefinedIn
	}
	for _, want := range []string{
		"component://customer-notification-service",
		"team://customer-engagement",
	} {
		if _, ok := got[want]; !ok {
			t.Errorf("registry is missing %q; got %v", want, got)
		}
	}
	if defined := got["team://customer-engagement"]; defined != "teams/customer-engagement/team.yaml" {
		t.Errorf("definedIn = %q, want teams/customer-engagement/team.yaml", defined)
	}
}

// TestLoadRegistry_MissingRegistryReturnsEmpty ports ST-027 (:155): an absent or
// empty registry yields no entries and no error.
func TestLoadRegistry_MissingRegistryReturnsEmpty(t *testing.T) {
	for name, seed := range map[string]string{
		"absent": "",
		"empty":  "",
		"null":   "null\n",
		// A registry whose `ids:` key is missing entirely is still well-formed.
		"no-ids-key": "schemaVersion: '1.0'\nkind: IdRegistry\n",
		// `or []` is truthiness, so an explicitly null list is an empty one.
		"null-ids": "kind: IdRegistry\nids:\n",
		// A row without a truthy id is dropped by the comprehension.
		"blank-id": "kind: IdRegistry\nids:\n- {id: '', definedIn: a.yaml}\n",
	} {
		t.Run(name, func(t *testing.T) {
			ws := workspace.New(t.TempDir())
			if name != "absent" {
				dir := filepath.Join(ws.Root, "company-ontology", "ids")
				if err := os.MkdirAll(dir, 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(dir, "registry.yaml"),
					[]byte(seed), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			entries, err := ids.Load(ws)
			if err != nil {
				t.Fatalf("Load: %v", err)
			}
			if len(entries) != 0 {
				t.Errorf("got %d entries, want 0: %v", len(entries), entries)
			}
		})
	}
}

// TestSuggestIDs_ClosestMatchFirst ports ST-028 (:161): a typo'd component id
// yields the real id as the top closest match.
func TestSuggestIDs_ClosestMatchFirst(t *testing.T) {
	got, err := ids.Suggest(exampleWorkspace(t), "customer-notifcation-servce", "component")
	if err != nil {
		t.Fatalf("Suggest: %v", err)
	}
	if len(got) == 0 || got[0] != "component://customer-notification-service" {
		t.Errorf("Suggest = %v, want component://customer-notification-service first", got)
	}
}

// TestSuggestIDs_ScopedAndCapped ports ST-029 (:163): suggestions never exceed
// three, and the scheme filter admits nothing outside its own scheme.
func TestSuggestIDs_ScopedAndCapped(t *testing.T) {
	ws := exampleWorkspace(t)
	got, err := ids.Suggest(ws, "customer-notifcation-servce", "component")
	if err != nil {
		t.Fatalf("Suggest: %v", err)
	}
	if len(got) > 3 {
		t.Errorf("Suggest returned %d results, want <= 3: %v", len(got), got)
	}
	for _, s := range got {
		if !hasPrefix(s, "component://") {
			t.Errorf("Suggest leaked a non-component id: %q", s)
		}
	}
	// An unscoped call over the same registry is free to reach other schemes,
	// which is what makes the scoping above meaningful rather than vacuous.
	all, err := ids.Suggest(ws, "communications", "")
	if err != nil {
		t.Fatalf("Suggest: %v", err)
	}
	if len(all) > 3 {
		t.Errorf("unscoped Suggest returned %d results, want <= 3: %v", len(all), all)
	}
}

func hasPrefix(s, p string) bool { return len(s) >= len(p) && s[:len(p)] == p }

// TestList_FilterCombinationsAreConjunctive pins the three filters at `:1289`
// as three independent `continue`s: --team and --platform together admit
// nothing, because no id is defined under both roots.
func TestList_FilterCombinationsAreConjunctive(t *testing.T) {
	ws := exampleWorkspace(t)
	cases := []struct {
		name    string
		filter  ids.Filter
		matched int
	}{
		{"none", ids.Filter{}, 7},
		{"team", ids.Filter{Team: "customer-engagement"}, 1},
		{"platform", ids.Filter{Platform: "communications"}, 4},
		{"prefix", ids.Filter{Prefix: "component://"}, 1},
		{"prefix-nomatch", ids.Filter{Prefix: "zzzz"}, 0},
		{"team-and-platform",
			ids.Filter{Team: "customer-engagement", Platform: "communications"}, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			sections, err := ids.List(ws, "", tc.filter)
			if err != nil {
				t.Fatalf("List: %v", err)
			}
			count := findingByCode(t, sections, model.CodeCount)
			if got := count.Fields.Int("matched"); got != tc.matched {
				t.Errorf("matched = %d, want %d", got, tc.matched)
			}
			if got := count.Fields.Int("total"); got != 7 {
				t.Errorf("total = %d, want 7", got)
			}
		})
	}
}

// TestList_EntryCarriesSchemeAndLocalName pins the two fields the text renderer
// never prints. They exist so a listing UI can group by scheme without
// re-parsing a display string (R-2.3).
func TestList_EntryCarriesSchemeAndLocalName(t *testing.T) {
	sections, err := ids.List(exampleWorkspace(t), "", ids.Filter{Prefix: "component://"})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	e := findingByCode(t, sections, model.CodeEntry)
	if got := e.Fields.Str("scheme"); got != "component" {
		t.Errorf("scheme = %q, want component", got)
	}
	if got := e.Fields.Str("localName"); got != "customer-notification-service" {
		t.Errorf("localName = %q, want customer-notification-service", got)
	}
	if e.Path != "platforms/communications/components/customer-notification-service.yaml" {
		t.Errorf("Path = %q, want the descriptor path", e.Path)
	}
}

// TestList_NoRegistryEmitsTheEmptyRecord covers the `:1280` early return: a
// workspace with no ontology produces one finding and no listing.
func TestList_NoRegistryEmitsTheEmptyRecord(t *testing.T) {
	sections, err := ids.List(workspace.New(t.TempDir()), "", ids.Filter{})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(sections) != 1 || len(sections[0].Findings) != 1 {
		t.Fatalf("got %d section(s), want one section with one finding", len(sections))
	}
	if code := sections[0].Findings[0].Code; code != model.CodeRegistryEmpty {
		t.Errorf("code = %q, want %q", code, model.CodeRegistryEmpty)
	}
}

// TestList_GlossaryPrecedesAnEmptyRegistry pins the print order at `:1276-1283`:
// the --role legend is emitted before the emptiness check, so it survives a
// workspace with no registry at all.
func TestList_GlossaryPrecedesAnEmptyRegistry(t *testing.T) {
	sections, err := ids.List(workspace.New(t.TempDir()), "team-lead", ids.Filter{})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(sections) != 2 {
		t.Fatalf("got %d section(s), want glossary + registry", len(sections))
	}
	if sections[0].Slug != "glossary" {
		t.Errorf("first section = %q, want glossary", sections[0].Slug)
	}
}

func findingByCode(t *testing.T, sections []model.GateResult, code string) model.Finding {
	t.Helper()
	for _, s := range sections {
		for _, f := range s.Findings {
			if f.Code == code {
				return f
			}
		}
	}
	t.Fatalf("no finding with code %q", code)
	return model.Finding{}
}
