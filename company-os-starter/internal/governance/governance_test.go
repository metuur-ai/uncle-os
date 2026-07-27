package governance_test

// What oracle_test.go cannot measure.
//
// Three groups: the R-0.7c write guard, whose whole point is that it produces
// LESS filesystem churn than the oracle and therefore cannot be proven by
// comparing against it; the mandatory-rule invariant, which is a fact about the
// generated document rather than about stdout; and the wrong-shape refusals,
// where the oracle raises a traceback and R-0.7a(j) requires a clean exit 4 with
// nothing written.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/governance"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/model"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/workspace"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/yamlio"
)

// canonicalize re-emits a document key-sorted, which is the same comparison
// form the guard uses and a layout that shares no line with the committed,
// insertion-ordered one.
func canonicalize(text string) (string, error) {
	v, err := yamlio.PyLoadBytes([]byte(text), "reflow")
	if err != nil {
		return "", err
	}
	return yamlio.PyDumpCanonical(v)
}

// ---------------------------------------------------------------- R-0.7c

// TestResolveLeavesAnUnchangedGeneratedFileAlone is R-0.7c's third site and the
// second half of R-0.10.
//
// The committed effective-governance.yaml is byte-stable under safe_dump today
// EXCEPT for generatedAt, so an unguarded writer rewrites one line on every
// invocation and `git status` is never clean after `governance resolve`. The
// assertion is the strongest available: not "the content is equivalent" but
// "the file was not touched at all", down to the modification time.
func TestResolveLeavesAnUnchangedGeneratedFileAlone(t *testing.T) {
	ws := copyFixture(t, "workspace")
	out := filepath.Join(ws, "teams", "customer-engagement", "generated",
		"effective-governance.yaml")

	before := readFile(t, out)
	beforeStat, err := os.Stat(out)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := governance.ResolveSections(workspace.New(ws), "customer-engagement"); err != nil {
		t.Fatalf("ResolveSections: %v", err)
	}
	if after := readFile(t, out); after != before {
		t.Errorf("the guard rewrote an unchanged derived file\nbefore:\n%s\nafter:\n%s",
			before, after)
	}
	afterStat, err := os.Stat(out)
	if err != nil {
		t.Fatal(err)
	}
	if !afterStat.ModTime().Equal(beforeStat.ModTime()) {
		t.Error("the guard rewrote identical bytes; the file must not be opened for writing")
	}
}

// TestResolveGuardIsSemanticNotByteWise proves the guard compares the parsed
// STRUCTURE, in both directions.
//
// Re-laying the committed file with sorted keys and a different indent leaves
// every byte different and the structure identical: an unchanged derivation
// must still be left alone. Dropping a requirement from it changes the
// structure and must be regenerated. A byte compare gets the first case wrong;
// comparing only "does the file exist" gets the second wrong.
func TestResolveGuardIsSemanticNotByteWise(t *testing.T) {
	t.Run("relaid-but-equivalent-is-left-alone", func(t *testing.T) {
		ws := copyFixture(t, "workspace")
		out := filepath.Join(ws, "teams", "customer-engagement", "generated",
			"effective-governance.yaml")
		// Flow style throughout and a four-space indent: no byte in common with
		// the committed layout, same document.
		relaid := reflow(t, readFile(t, out))
		writeFile(t, out, relaid)

		if _, err := governance.ResolveSections(workspace.New(ws), "customer-engagement"); err != nil {
			t.Fatalf("ResolveSections: %v", err)
		}
		if got := readFile(t, out); got != relaid {
			t.Errorf("a structurally identical relayout was rewritten:\n%s", got)
		}
	})

	t.Run("drifted-is-regenerated", func(t *testing.T) {
		ws := copyFixture(t, "workspace")
		out := filepath.Join(ws, "teams", "customer-engagement", "generated",
			"effective-governance.yaml")
		drifted := strings.Replace(readFile(t, out),
			"deviationsApplied:", "deviationsApplied: []\nstale: true\n#", 1)
		writeFile(t, out, drifted)

		if _, err := governance.ResolveSections(workspace.New(ws), "customer-engagement"); err != nil {
			t.Fatalf("ResolveSections: %v", err)
		}
		got := readFile(t, out)
		if got == drifted {
			t.Error("a drifted derived file was left alone")
		}
		if strings.Contains(got, "stale: true") {
			t.Errorf("the regenerated file still carries the drift:\n%s", got)
		}
	})
}

// TestResolveWritesWhenTheGeneratedFileIsAbsent is the guard's other half: it
// must not turn "unchanged" into "never write".
func TestResolveWritesWhenTheGeneratedFileIsAbsent(t *testing.T) {
	ws := copyFixture(t, "standalone-team")
	out := filepath.Join(ws, "teams", "solo", "generated", "effective-governance.yaml")
	if _, err := os.Stat(out); err == nil {
		t.Skip("fixture already carries a generated file")
	}
	res, err := governance.Resolve(workspace.New(ws), "solo")
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if !res.Written {
		t.Error("Written = false with no committed file to compare against")
	}
	if _, err := os.Stat(out); err != nil {
		t.Fatalf("the generated file was not created: %v", err)
	}
}

// TestResolveRewritesAMalformedGeneratedFile: the oracle never reads the file it
// is about to overwrite, so a file the guard cannot parse must not become a file
// the guard refuses to replace.
func TestResolveRewritesAMalformedGeneratedFile(t *testing.T) {
	ws := copyFixture(t, "workspace")
	out := filepath.Join(ws, "teams", "customer-engagement", "generated",
		"effective-governance.yaml")
	writeFile(t, out, "this: [is not: valid: yaml\n")

	res, err := governance.Resolve(workspace.New(ws), "customer-engagement")
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if !res.Written {
		t.Error("Written = false over an unparseable file")
	}
	if strings.Contains(readFile(t, out), "is not") {
		t.Error("the malformed file survived")
	}
}

// ---------------------------------------------------- the tier invariant

// TestDeviationAimedAtAMandatoryRuleIsRecordedNotRefused is invariant #1 of the
// methodology, and the reason `deviation declare` has no validation in it.
//
// The oracle records `deviationRejected` into the generated file and CONTINUES
// (`bin/company-os:317-319`); it is not an exit site, and `cmd_deviation`
// (`:1112-1125`) always exits 0. The refusal surfaces later as a validate gate
// failure. Adding enforcement at declare time would be a behaviour change, not
// a port — see .devlocal/go-port/exit-code-map.md § "Code 5's third example".
func TestDeviationAimedAtAMandatoryRuleIsRecordedNotRefused(t *testing.T) {
	ws := copyFixture(t, "workspace")
	w := workspace.New(ws)

	// delivery-reliability is `level: mandatory` in the communications catalog.
	if _, err := governance.DeclareSections(w, "customer-engagement",
		"platform-standard://communications/delivery-reliability", ""); err != nil {
		t.Fatalf("declare must not refuse a mandatory rule: %v", err)
	}
	res, err := governance.Resolve(w, "customer-engagement")
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	text := readFile(t, res.Path)
	if !strings.Contains(text, "mandatory rules cannot be deviated; use an exception") {
		t.Errorf("deviationRejected was not recorded:\n%s", text)
	}
	// And the rejected rule must NOT appear under deviationsApplied.
	applied := text[strings.Index(text, "deviationsApplied:"):]
	if strings.Contains(applied, "delivery-reliability") {
		t.Errorf("a rejected deviation was counted as applied:\n%s", applied)
	}
}

// TestDeclaringTheSameRuleTwiceKeepsTheLast is `{d["rule"]: d for d in …}`
// (`:270`): the dict comprehension lets the LAST entry win, so declaring twice
// behaves as one declaration rather than doubling it.
func TestDeclaringTheSameRuleTwiceKeepsTheLast(t *testing.T) {
	ws := copyFixture(t, "workspace")
	w := workspace.New(ws)
	rule := "platform-standard://communications/prd-structure"

	if _, err := governance.Declare(w, "customer-engagement", rule, "first"); err != nil {
		t.Fatal(err)
	}
	if _, err := governance.Declare(w, "customer-engagement", rule, "second"); err != nil {
		t.Fatal(err)
	}
	res, err := governance.Resolve(w, "customer-engagement")
	if err != nil {
		t.Fatal(err)
	}
	text := readFile(t, res.Path)
	if !strings.Contains(text, "rationale: second") {
		t.Errorf("the last declaration did not win:\n%s", text)
	}
	if strings.Contains(text, "rationale: first") {
		t.Errorf("an overwritten declaration survived into the resolution:\n%s", text)
	}
	// Both entries are still in the AUTHORED file — declare appends, it never
	// deduplicates.
	dev := readFile(t, filepath.Join(ws, "teams", "customer-engagement",
		"governance", "deviations.yaml"))
	if strings.Count(dev, rule) != 3 {
		t.Errorf("declare should have appended twice to the committed entry:\n%s", dev)
	}
}

// ------------------------------------------------------------ R-0.7a(j)

// TestWrongShapeArtifactsRefuseAndWriteNothing covers the well-formed-YAML,
// wrong-shape paths where the oracle raises AttributeError/TypeError. The
// filesystem effect is NOT carved out: Python writes nothing, so neither may
// this. Each case asserts exit 4 AND that the target file is untouched.
func TestWrongShapeArtifactsRefuseAndWriteNothing(t *testing.T) {
	cases := []struct {
		name    string
		file    string // workspace-relative, replaced before the run
		content string
		run     func(w *workspace.Workspace) error
	}{
		{
			name:    "ownership-components-is-a-scalar",
			file:    "teams/customer-engagement/ownership/components.yaml",
			content: "components: not-a-list\n",
			run: func(w *workspace.Workspace) error {
				_, err := governance.Resolve(w, "customer-engagement")
				return err
			},
		},
		{
			name:    "ownership-component-entry-is-a-string",
			file:    "teams/customer-engagement/ownership/components.yaml",
			content: "components: [just-an-id]\n",
			run: func(w *workspace.Workspace) error {
				_, err := governance.Resolve(w, "customer-engagement")
				return err
			},
		},
		{
			name:    "deviations-root-is-a-list",
			file:    "teams/customer-engagement/governance/deviations.yaml",
			content: "- rule: x\n",
			run: func(w *workspace.Workspace) error {
				_, err := governance.Declare(w, "customer-engagement", "x", "")
				return err
			},
		},
		{
			name:    "deviations-key-is-a-mapping",
			file:    "teams/customer-engagement/governance/deviations.yaml",
			content: "deviations: {a: 1}\n",
			run: func(w *workspace.Workspace) error {
				_, err := governance.Declare(w, "customer-engagement", "x", "")
				return err
			},
		},
		{
			name:    "exceptions-key-is-a-scalar",
			file:    "teams/customer-engagement/governance/exceptions.yaml",
			content: "exceptions: 7\n",
			run: func(w *workspace.Workspace) error {
				_, err := governance.Request(w, "customer-engagement", "r", "c", "", "2035-01-01")
				return err
			},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ws := copyFixture(t, "workspace")
			target := filepath.Join(ws, filepath.FromSlash(c.file))
			writeFile(t, target, c.content)

			err := c.run(workspace.New(ws))
			if err == nil {
				t.Fatal("a wrong-shape artifact was accepted")
			}
			if got := model.CodeOf(err); got != model.ExitArtifact {
				t.Errorf("exit code = %d, want %d (%v)", got, model.ExitArtifact, err)
			}
			if got := readFile(t, target); got != c.content {
				t.Errorf("the refused path still wrote:\n%s", got)
			}
		})
	}
}

// TestResolveRefusesAnUnknownPlatform is not R-0.7a(j) but its neighbour:
// `ws.platform_dir(pid)` (`:302`) is inside the relationship loop, so a
// descriptor naming a platform that is not in the catalog dies with exit 3 and
// no generated file is written — even though the earlier components resolved
// cleanly.
func TestResolveRefusesAnUnknownPlatform(t *testing.T) {
	ws := copyFixture(t, "workspace")
	desc := filepath.Join(ws, "platforms", "communications", "components",
		"customer-notification-service.yaml")
	writeFile(t, desc, strings.Replace(readFile(t, desc),
		"platform://communications", "platform://ghost-platform", 1))
	out := filepath.Join(ws, "teams", "customer-engagement", "generated",
		"effective-governance.yaml")
	before := readFile(t, out)

	_, err := governance.Resolve(workspace.New(ws), "customer-engagement")
	if err == nil {
		t.Fatal("an unknown platform was accepted")
	}
	if got := model.CodeOf(err); got != model.ExitWorkspace {
		t.Errorf("exit code = %d, want %d (%v)", got, model.ExitWorkspace, err)
	}
	if readFile(t, out) != before {
		t.Error("the generated file was written on a path that dies")
	}
}

// TestDuplicatePlatformRelationshipTalliesLikeTheOracle pins the one place the
// resolve block's two numbers are computed differently from each other.
//
// `platforms:` is a LIST and is appended to per relationship;
// `requirements.platform[pid]` is a DICT KEY and is overwritten. So a descriptor
// naming the same platform twice renders `platforms [communications,
// communications]` with the requirement count UNCHANGED. Accumulating the count
// as the loop runs doubles it, which is why the tally is read back out of the
// finished entry.
// It used to assert this by running the Python reference over the same
// duplicated fixture and diffing stdout. R-9.3 deleted that binary, so the test
// could only skip. The property does not need an oracle to state, so it is
// asserted directly instead: the platform appears TWICE in the list and the
// requirement counts are UNCHANGED from the unduplicated resolve. That is the
// whole of what the differential was checking here, minus the claim that Python
// agrees — which nothing can check any more.
func TestDuplicatePlatformRelationshipTalliesLikeTheOracle(t *testing.T) {
	resolve := func(ws string) string {
		t.Helper()
		sections, err := governance.ResolveSections(workspace.New(ws), "customer-engagement")
		if err != nil {
			t.Fatalf("ResolveSections: %v", err)
		}
		return renderSections(t, sections)
	}

	// Baseline: the fixture as committed.
	baseline := resolve(copyFixture(t, "workspace"))
	if !strings.Contains(baseline, "platforms [communications]") {
		t.Fatalf("fixture no longer renders the expected single relationship:\n%s", baseline)
	}

	// Same fixture with the belongs-to relationship stated twice.
	dupWS := copyFixture(t, "workspace")
	desc := filepath.Join(dupWS, "platforms", "communications", "components",
		"customer-notification-service.yaml")
	text := readFile(t, desc)
	const block = "- platform: platform://communications\n  relationship: belongs-to\n"
	if !strings.Contains(text, block) {
		t.Fatalf("fixture no longer carries the expected relationship block:\n%s", text)
	}
	writeFile(t, desc, strings.Replace(text, block, block+block, 1))
	got := resolve(dupWS)

	// `platforms:` is a list, so the duplicate shows up.
	if !strings.Contains(got, "[communications, communications]") {
		t.Fatalf("duplicate relationship did not reach the platforms list:\n%s", got)
	}
	// ...but the counts are dict keys, so they must not double. Comparing the
	// whole tail after the platforms list catches both counts at once.
	countsOf := func(s string) string {
		_, after, ok := strings.Cut(s, "platforms [")
		if !ok {
			t.Fatalf("no platforms list in:\n%s", s)
		}
		_, counts, ok := strings.Cut(after, "], ")
		if !ok {
			t.Fatalf("no counts after the platforms list in:\n%s", s)
		}
		return counts
	}
	if a, b := countsOf(baseline), countsOf(got); a != b {
		t.Errorf("duplicating the relationship changed the tally\n"+
			" once  : %q\n twice : %q\nthe counts are dict keys and must be overwritten, "+
			"not accumulated", a, b)
	}
}

// TestExplainUnresolvedCarriesTheContract pins the failure record `governance
// explain` returns. `:367` is exit 3 per .devlocal/go-port/exit-code-map.md § B,
// and the map is explicit that it COLLAPSES "unknown id" with "resolve was never
// run". That collapse is reproduced deliberately: both shapes below get the same
// code and the same sentence, and only the suggestion list differs.
func TestExplainUnresolvedCarriesTheContract(t *testing.T) {
	ws := workspace.New(filepath.Join(repoRoot(t), "examples", "workspace"))

	for _, c := range []struct {
		name, component string
		wantSuggestions []string
	}{
		{"near-miss", "customer-notification-servic",
			[]string{"component://customer-notification-service"}},
		{"nothing-close", "zzzzzzzz", nil},
	} {
		t.Run(c.name, func(t *testing.T) {
			_, err := governance.Explain(ws, c.component)
			if err == nil {
				t.Fatal("an unresolved component was accepted")
			}
			if got := model.CodeOf(err); got != model.ExitWorkspace {
				t.Errorf("exit code = %d, want %d", got, model.ExitWorkspace)
			}
			var ue *governance.UnresolvedError
			if !asUnresolved(err, &ue) {
				t.Fatalf("error is not *UnresolvedError: %T", err)
			}
			if ue.Component != c.component {
				t.Errorf("Component = %q, want %q", ue.Component, c.component)
			}
			if strings.Join(ue.Suggestions, ",") != strings.Join(c.wantSuggestions, ",") {
				t.Errorf("Suggestions = %v, want %v", ue.Suggestions, c.wantSuggestions)
			}
		})
	}
}

func asUnresolved(err error, target **governance.UnresolvedError) bool {
	ue, ok := err.(*governance.UnresolvedError)
	if ok {
		*target = ue
	}
	return ok
}

// --------------------------------------------------------------- helpers

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// reflow re-emits a document through the canonical (key-sorted) form, which
// shares no line with the insertion-ordered committed layout while parsing to
// the same structure.
func reflow(t *testing.T, text string) string {
	t.Helper()
	out, err := canonicalize(text)
	if err != nil {
		t.Fatalf("reflowing: %v", err)
	}
	if out == text {
		t.Fatal("the reflow produced identical bytes; the test would prove nothing")
	}
	return out
}
