package render

import (
	"fmt"
	"io"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/model"
)

// Validate writes `validate` as the Python CLI writes it
// (bin/company-os:922-1107).
//
// There is ONE prefix rule, not a per-gate discriminator: emit Subject, then
// ": ", then Message — and Message alone when Subject is empty. The LLD's table
// of "seven distinct prefix shapes" describes seven distinct Subject VALUES
// chosen by the producers, and Subject is render-ready text that may carry
// punctuation for exactly that reason: gate 1 alone emits a team id (`:941`), a
// single-quoted component id (`:946`) and a bare component id (`:951`), and gate
// 5 emits `<root>/team.yaml`, a bare `<root>` and `<root>/CLAUDE.md`. A renderer
// that branched per gate would have to be re-derived every time a producer
// changed its mind about a prefix; this one cannot go out of date.
//
// Three things are DERIVED here rather than carried in the records:
//
//   - The blank line before a gate header belongs to the header and is present
//     on every gate except the first (R-2.6) — exactly Ordinal > 1. It survives a
//     gate with zero findings, which examples/failing-federated-golden-validate.txt
//     lines 3-4 fix.
//   - The trailer counts SevFail only. examples/failing-workspace-golden-validate.txt
//     has 15 fails and 4 warns and its trailer reads 15.
//
// The banner section is not a gate: it carries the workspace root, the
// completeness flag, and the [N/M] denominator, and is identified by
// model.SlugWorkspace, which is what keeps it out of the gate list.
//
// The denominator is READ from the banner, never counted off the gate list
// (R-2.6a). On a complete run the two agree; on a run that aborted mid-gate the
// list is short and only the carried number is the one the oracle printed.
func Validate(w io.Writer, sections []model.GateResult) error {
	report := model.Report{}
	complete := false
	// fixCount is -1 when --fix was not requested. A non-negative value means
	// --fix ran and that many files changed; the summary line is printed after
	// the PASS/FAIL trailer.
	fixCount := -1
	for _, s := range sections {
		switch s.Slug {
		case model.SlugWorkspace:
			for _, f := range s.Findings {
				if f.Code == model.CodeValidateRoot {
					report.Root = f.Fields.Str("root")
					report.Total = f.Fields.Int("gates")
					complete, _ = f.Fields["complete"].(bool)
				}
			}
		case model.SlugFixSummary:
			// ux-simplification 2.1: extract the regenerated count. The
			// section is not a gate — it rides after the trailer, not in the
			// numbered list.
			for _, f := range s.Findings {
				if f.Code == model.CodeFixRegenerated {
					fixCount = f.Fields.Int("regenerated")
				}
			}
		default:
			report.Gates = append(report.Gates, s)
		}
	}

	if _, err := fmt.Fprintf(w, "validating workspace %s\n\n", report.Root); err != nil {
		return err
	}
	for _, g := range report.Gates {
		if g.Ordinal > 1 {
			if _, err := fmt.Fprintln(w); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprintf(w, "[%d/%d] %s\n", g.Ordinal, report.Total, g.Title); err != nil {
			return err
		}
		for _, f := range g.Findings {
			marker, err := severityMarker(f.Severity)
			if err != nil {
				return err
			}
			line := f.Message
			if f.Subject != "" {
				line = f.Subject + ": " + f.Message
			}
			if _, err := fmt.Fprintf(w, "  [%s] %s\n", marker, line); err != nil {
				return err
			}
		}
	}
	// The trailer is a VERDICT, and a run that aborted on a malformed artifact
	// reached no verdict. The oracle dies mid-command there, having printed the
	// banner and nothing else (bin/company-os:924 precedes the manifest load at
	// `:929`), so the banner's `complete` field is what separates "checked
	// everything and found nothing" from "never finished checking". Without it an
	// aborted run would print PASS.
	if !complete {
		return nil
	}
	if n := report.Problems(); n > 0 {
		if _, err := fmt.Fprintf(w, "\nFAIL — %d problem(s)\n", n); err != nil {
			return err
		}
		return writeFixSummary(w, fixCount)
	}
	if _, err := fmt.Fprint(w, "\nPASS\n"); err != nil {
		return err
	}
	return writeFixSummary(w, fixCount)
}

// writeFixSummary emits the `validate --fix: N file(s) regenerated` line when
// --fix was requested. fixCount < 0 means the flag was absent and nothing is
// printed — the default path stays byte-identical.
func writeFixSummary(w io.Writer, fixCount int) error {
	if fixCount < 0 {
		return nil
	}
	_, err := fmt.Fprintf(w, "validate --fix: %d file(s) regenerated\n", fixCount)
	return err
}

// severityMarker is the bracketed text of ok()/warn()/fail()
// (bin/company-os:47-55). The capitalization is not cosmetic — `[FAIL]` is what
// CI and every skill grep for — so it lives here rather than on Severity, whose
// String() is the JSON-facing lowercase name.
func severityMarker(s model.Severity) (string, error) {
	switch s {
	case model.SevOK:
		return "ok", nil
	case model.SevWarn:
		return "warn", nil
	case model.SevFail:
		return "FAIL", nil
	}
	return "", fmt.Errorf("render: validate: unknown severity %d", int(s))
}
