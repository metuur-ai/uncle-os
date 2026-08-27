package render

import (
	"fmt"
	"io"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/model"
)

// Next writes the next-action view (ux-simplification 1.1).
//
// Default mode emits one action with its recommended command. `--all` mode
// emits grouped sections with headers. The renderer switches on Code to
// decide the layout; the records carry all the data.
//
// The leading blank line before each group header follows the convention
// `today` uses for platform/team blocks — a visual separator from whatever
// precedes it.
func Next(w io.Writer, sections []model.GateResult) error {
	for _, s := range sections {
		for _, f := range s.Findings {
			var err error
			switch f.Code {
			case model.CodeNextGroup:
				// Group header in --all mode.
				_, err = fmt.Fprintf(w, "\n%s\n", f.Message)
			case model.CodeNextExpiry:
				if isGrouped(f) {
					err = renderGroupedItem(w, f)
				} else {
					err = renderNextExpiry(w, f)
				}
			case model.CodeNextContract:
				if isGrouped(f) {
					err = renderGroupedItem(w, f)
				} else {
					err = renderNextContract(w, f)
				}
			case model.CodeNextDoneCheck:
				if isGrouped(f) {
					err = renderGroupedItem(w, f)
				} else {
					err = renderNextDoneCheck(w, f)
				}
			case model.CodeNextOutcome:
				if isGrouped(f) {
					err = renderGroupedItem(w, f)
				} else {
					err = renderNextOutcome(w, f)
				}
			case model.CodeNextEmpty:
				_, err = fmt.Fprintf(w, "%s\n", f.Message)
			default:
				err = fmt.Errorf("render: next: no rule for finding code %q", f.Code)
			}
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// renderNextExpiry writes one expired or soon-due escape hatch.
//
// Layout:
//
//	next: deviation for <rule> expired <date> — re-review or remove
//	  team: <team>
//	  run: company-os deviation declare <rule> --team <team>
func renderNextExpiry(w io.Writer, f model.Finding) error {
	kind := f.Fields.Str("kind")
	rule := f.Fields.Str("rule")
	date := f.Fields.Str("date")
	team := f.Fields.Str("team")
	status := f.Fields.Str("status")
	verb := "expired"
	if status == "soon-due" {
		verb = "due"
	}
	if _, err := fmt.Fprintf(w, "next: %s for %s %s %s — re-review or remove\n",
		kind, rule, verb, date); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "  team: %s\n", team); err != nil {
		return err
	}
	if next := f.Fields.Str(model.FieldNext); next != "" {
		if _, err := fmt.Fprintf(w, "  run: %s\n", next); err != nil {
			return err
		}
	}
	return nil
}

// renderNextContract writes one active PRD with contract issues.
func renderNextContract(w io.Writer, f model.Finding) error {
	prd := f.Fields.Str("prd")
	platform := f.Fields.Str("platform")
	issues := f.Fields.Int("issues")
	if _, err := fmt.Fprintf(w, "next: PRD %s has %d contract issue(s)\n",
		prd, issues); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "  platform: %s\n", platform); err != nil {
		return err
	}
	if next := f.Fields.Str(model.FieldNext); next != "" {
		if _, err := fmt.Fprintf(w, "  run: %s\n", next); err != nil {
			return err
		}
	}
	return nil
}

// renderNextDoneCheck writes one active PRD with done-check problems.
func renderNextDoneCheck(w io.Writer, f model.Finding) error {
	prd := f.Fields.Str("prd")
	platform := f.Fields.Str("platform")
	unchecked := f.Fields.Int("unchecked")
	stale := f.Fields.Int("stale")
	if _, err := fmt.Fprintf(w,
		"next: PRD %s has %d unchecked item(s) and %d stale/missing reality doc(s)\n",
		prd, unchecked, stale); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "  platform: %s\n", platform); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "  fix: update reality + check items, then complete\n"); err != nil {
		return err
	}
	if next := f.Fields.Str(model.FieldNext); next != "" {
		if _, err := fmt.Fprintf(w, "  run: %s\n", next); err != nil {
			return err
		}
	}
	return nil
}

// renderNextOutcome writes one pending outcome review.
func renderNextOutcome(w io.Writer, f model.Finding) error {
	prd := f.Fields.Str("prd")
	due := f.Fields.Str("due")
	platform := f.Fields.Str("platform")
	if _, err := fmt.Fprintf(w, "next: outcome review for %s due %s\n",
		prd, due); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "  platform: %s\n", platform); err != nil {
		return err
	}
	return nil
}

// isGrouped reports whether a finding belongs to a --all group rather than
// being the single-action view.
func isGrouped(f model.Finding) bool {
	v, _ := f.Fields["grouped"].(bool)
	return v
}

// renderGroupedItem writes one item inside a --all group as a dash-prefixed
// list entry, with the recommended command indented below it when present.
func renderGroupedItem(w io.Writer, f model.Finding) error {
	if _, err := fmt.Fprintf(w, "  - %s\n", f.Message); err != nil {
		return err
	}
	if next := f.Fields.Str(model.FieldNext); next != "" {
		if _, err := fmt.Fprintf(w, "    run: %s\n", next); err != nil {
			return err
		}
	}
	return nil
}
