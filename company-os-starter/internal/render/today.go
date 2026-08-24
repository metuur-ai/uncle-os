package render

import (
	"fmt"
	"io"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/model"
)

// Today writes the role-aware daily view as the Python CLI writes it
// (bin/company-os:1168-1203).
//
// The leading blank lines are part of this renderer, not of the records: in
// Python they are literal "\n" prefixes inside three f-strings (`:1177`, `:1194`,
// `:1203`) and are absent from the header, the legend, and the warn at `:1192`.
// There is no derivable rule — `today` is not `validate`, where a blank belongs
// to every gate header after the first — so each line states its own.
func Today(w io.Writer, sections []model.GateResult) error {
	for _, s := range sections {
		if s.Slug == model.SlugGlossary {
			if err := glossary(w, s); err != nil {
				return err
			}
			continue
		}
		for _, f := range s.Findings {
			var err error
			switch f.Code {
			case model.CodeHeader:
				_, err = fmt.Fprintf(w, "== today (%s) ==\n", f.Fields.Str("role"))
			case model.CodePlatform:
				_, err = fmt.Fprintf(w, "\nplatform %s: %d active PRD(s)\n",
					f.Fields.Str("platform"), f.Fields.Int("activePRDs"))
			case model.CodeActivePRD:
				_, err = fmt.Fprintf(w, "  - %s [%s] team=%s\n",
					f.Fields.Str("prd"), f.Fields.Str("status"), f.Fields.Str("team"))
			case model.CodeOutcomeReview:
				_, err = fmt.Fprintf(w, "  - outcome review due %s: %s\n",
					f.Fields.Str("due"), f.Fields.Str("prd"))
			case model.CodeGovernanceMissing:
				// warn() writes to stdout with a two-space indent
				// (bin/company-os:50-51), so this line is part of the view.
				_, err = fmt.Fprintf(w,
					"  [warn] %s: no effective-governance.yaml — run governance resolve\n",
					f.Fields.Str("team"))
			case model.CodeTeam:
				_, err = fmt.Fprintf(w, "\nteam %s (governance generated %s)\n",
					f.Fields.Str("team"), f.Fields.Str("generatedAt"))
			case model.CodeComponent:
				_, err = fmt.Fprintf(w,
					"  - %s: %d platform requirement(s), %d company control(s)\n",
					f.Fields.Str("component"),
					f.Fields.Int("platformRequirements"),
					f.Fields.Int("companyControls"))
			// R-9.1. The leading blank line follows the platform/team blocks'
			// convention (`:1177`, `:1194`) rather than deriving one: this is a
			// new block in the same view, and it separates itself from whatever
			// precedes it exactly as they do.
			case model.CodeDraftsHeader:
				_, err = fmt.Fprintf(w, "\nteam %s: %d open draft(s)\n",
					f.Fields.Str("team"), f.Fields.Int("drafts"))
			// title and platform are already resolved to "unset" by
			// roles.draftsSection (R-9.3), so this line never has to know what a
			// placeholder looks like.
			case model.CodeDraft:
				_, err = fmt.Fprintf(w, "  - %s: %s -> %s\n",
					f.Fields.Str("draft"), f.Fields.Str("title"),
					f.Fields.Str("platform"))
			case model.CodeOnboarding:
				_, err = fmt.Fprintf(w, "\nonboarding: %s\n", f.Fields.Str("guide"))
			default:
				err = fmt.Errorf("render: today: no rule for finding code %q", f.Code)
			}
			if err != nil {
				return err
			}
		}
	}
	return nil
}
