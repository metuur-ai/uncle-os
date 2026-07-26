package render_test

// The renderer's four derived behaviours, isolated from any fixture. The five
// golden snapshots in internal/validate exercise all of these through real
// workspaces; these cases exist so that a break points at render.Validate rather
// than at whichever fixture happened to carry the shape.

import (
	"bytes"
	"strings"
	"testing"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/model"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/render"
)

func banner(root string, total int) model.GateResult {
	return model.GateResult{Slug: model.SlugWorkspace, Findings: []model.Finding{{
		Code:   model.CodeValidateRoot,
		Fields: model.Fields{"root": root, "complete": true, "gates": total},
	}}}
}

func renderReport(t *testing.T, sections ...model.GateResult) string {
	t.Helper()
	var buf bytes.Buffer
	if err := render.Validate(&buf, sections); err != nil {
		t.Fatalf("render.Validate: %v", err)
	}
	return buf.String()
}

// TestUniformPrefixRule is R-2.5. Subject is render-ready text and the rule is
// one rule: Subject, ": ", Message — or Message alone. The three values below
// are gate 1's three "prefix shapes" and gate 7's absent one, rendered by one
// code path.
func TestUniformPrefixRule(t *testing.T) {
	got := renderReport(t, banner("/ws", 1), model.GateResult{
		Ordinal: 1, Title: "ownership reconciliation",
		Findings: []model.Finding{
			{Severity: model.SevFail, Subject: "ghost", Message: "owns 'x' but no descriptor"},
			{Severity: model.SevFail, Subject: "'svc'", Message: "team 'ghost' claims accountable"},
			{Severity: model.SevOK, Subject: "svc", Message: "registry and descriptor agree (beta)"},
			{Severity: model.SevWarn, Message: "no prefix at all"},
		},
	})
	want := "validating workspace /ws\n\n" +
		"[1/1] ownership reconciliation\n" +
		"  [FAIL] ghost: owns 'x' but no descriptor\n" +
		"  [FAIL] 'svc': team 'ghost' claims accountable\n" +
		"  [ok] svc: registry and descriptor agree (beta)\n" +
		"  [warn] no prefix at all\n" +
		"\nFAIL — 2 problem(s)\n"
	if got != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
}

// TestBlankLineBelongsToTheHeader is R-2.6: it is present on every gate except
// the first, which is Ordinal > 1 and nothing else — so it survives a gate that
// produced no findings at all
// (examples/failing-federated-golden-validate.txt:3-4).
func TestBlankLineBelongsToTheHeader(t *testing.T) {
	got := renderReport(t, banner("/ws", 3),
		model.GateResult{Ordinal: 1, Title: "first"},
		model.GateResult{Ordinal: 2, Title: "second"},
		model.GateResult{Ordinal: 3, Title: "third", Findings: []model.Finding{
			{Severity: model.SevOK, Message: "only line"},
		}},
	)
	want := "validating workspace /ws\n\n" +
		"[1/3] first\n" +
		"\n[2/3] second\n" +
		"\n[3/3] third\n" +
		"  [ok] only line\n" +
		"\nPASS\n"
	if got != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
}

// TestDenominatorExcludesTheBanner: the banner is a section but not a gate, so
// seven gates render [n/7] with it present, and adding gate 8 moves every
// denominator without touching a gate record.
func TestDenominatorExcludesTheBanner(t *testing.T) {
	gates := make([]model.GateResult, 0, 8)
	sections := []model.GateResult{banner("/ws", 7)}
	for i := 1; i <= 7; i++ {
		gates = append(gates, model.GateResult{Ordinal: i, Title: "g"})
	}
	sections = append(sections, gates...)
	if got := renderReport(t, sections...); !strings.Contains(got, "[7/7] g\n") {
		t.Errorf("seven gates did not render [7/7]:\n%s", got)
	}
	sections[0] = banner("/ws", 8)
	sections = append(sections, model.GateResult{Ordinal: 8, Title: "g"})
	got := renderReport(t, sections...)
	if !strings.Contains(got, "[1/8] g\n") || !strings.Contains(got, "[8/8] g\n") {
		t.Errorf("eight gates did not render [n/8]:\n%s", got)
	}
}

// TestTrailerCountsFailsOnly: the failing-workspace oracle has 15 fails and 4
// warns and its trailer reads 15.
func TestTrailerCountsFailsOnly(t *testing.T) {
	got := renderReport(t, banner("/ws", 1), model.GateResult{
		Ordinal: 1, Title: "g",
		Findings: []model.Finding{
			{Severity: model.SevWarn, Message: "a"},
			{Severity: model.SevWarn, Message: "b"},
			{Severity: model.SevFail, Message: "c"},
			{Severity: model.SevOK, Message: "d"},
		},
	})
	if !strings.HasSuffix(got, "\nFAIL — 1 problem(s)\n") {
		t.Errorf("trailer counted warns:\n%s", got)
	}
	clean := renderReport(t, banner("/ws", 1), model.GateResult{
		Ordinal: 1, Title: "g",
		Findings: []model.Finding{{Severity: model.SevWarn, Message: "a"}},
	})
	if !strings.HasSuffix(clean, "\nPASS\n") {
		t.Errorf("a warn-only report did not PASS:\n%s", clean)
	}
}

// TestUnknownSeverityIsAnError: a producer/renderer mismatch must be a loud
// failure here rather than a silently mislabelled line, because the bracketed
// markers are what CI and every skill grep for.
func TestUnknownSeverityIsAnError(t *testing.T) {
	var buf bytes.Buffer
	err := render.Validate(&buf, []model.GateResult{banner("/ws", 1), {
		Ordinal: 1, Title: "g",
		Findings: []model.Finding{{Severity: model.Severity(99), Message: "x"}},
	}})
	if err == nil {
		t.Error("an unknown severity rendered without error")
	}
}

// TestAbortedRunHasNoTrailer: a record set whose banner says the run did not
// finish renders the banner and stops. The oracle's own behaviour on a malformed
// workspace.yaml is exactly that — banner on stdout, diagnostic on stderr, no
// verdict — and without the guard an aborted run would print PASS.
func TestAbortedRunHasNoTrailer(t *testing.T) {
	got := renderReport(t, model.GateResult{
		Slug: model.SlugWorkspace,
		Findings: []model.Finding{{
			Code:   model.CodeValidateRoot,
			Fields: model.Fields{"root": "/ws", "complete": false, "gates": 7},
		}},
	})
	if got != "validating workspace /ws\n\n" {
		t.Errorf("aborted run rendered %q", got)
	}
}

// TestAbortedRunKeepsTheCarriedDenominator is R-2.6a. A run that died inside
// gate 2 hands the renderer a banner and two gates; the oracle had already
// printed those two under `[1/7]` and `[2/7]`, so the short list must not be
// allowed to renumber them. This is the case the port previously answered by
// dropping the completed gates altogether.
func TestAbortedRunKeepsTheCarriedDenominator(t *testing.T) {
	got := renderReport(t, model.GateResult{
		Slug: model.SlugWorkspace,
		Findings: []model.Finding{{
			Code:   model.CodeValidateRoot,
			Fields: model.Fields{"root": "/ws", "complete": false, "gates": 7},
		}},
	},
		model.GateResult{Ordinal: 1, Title: "ownership reconciliation", Findings: []model.Finding{
			{Severity: model.SevOK, Subject: "svc", Message: "registry and descriptor agree (beta)"},
		}},
		model.GateResult{Ordinal: 2, Title: "deviation and exception expiry", Findings: []model.Finding{
			{Severity: model.SevOK, Subject: "t", Message: "deviation x current (review 2035-01-15)"},
		}},
	)
	want := "validating workspace /ws\n\n" +
		"[1/7] ownership reconciliation\n" +
		"  [ok] svc: registry and descriptor agree (beta)\n" +
		"\n[2/7] deviation and exception expiry\n" +
		"  [ok] t: deviation x current (review 2035-01-15)\n"
	if got != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
}
