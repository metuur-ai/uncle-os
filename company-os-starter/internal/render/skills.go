package render

import (
	"fmt"
	"io"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/model"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/skills"
)

// Skills writes `skills list` as the Python CLI writes it
// (bin/company-os:869-917).
//
// Every sentence, indent and blank line in the command is composed here and
// nowhere else (R-2.8): internal/skills emits a code and typed fields per line.
// The blank lines are treated the way R-2.6 treats validate's — as a property
// of the header that follows, derived rather than stored, so the record set
// carries no formatting.
func Skills(w io.Writer, sections []model.GateResult) error {
	for _, s := range sections {
		for _, f := range s.Findings {
			if err := skillLine(w, f); err != nil {
				return err
			}
		}
	}
	return nil
}

func skillLine(w io.Writer, f model.Finding) error {
	switch f.Code {
	case model.CodeBanner:
		// The banner's own trailing blank line is print("...\n") at `:871`.
		return writeLines(w, "== agent skills (merged view across 4 layers) ==", "")

	case model.CodeLayersHeader:
		return writeLines(w, "layers (origin-labeled):")

	case model.CodeLayerEmpty:
		return writeLines(w, fmt.Sprintf("  [%s] <none>", f.Fields.Str("layer")))

	case model.CodeLayerEntry:
		return writeLines(w, layerEntry(f))

	case model.CodeMergedHeader:
		return writeLines(w, "",
			"merged guidance (canonical steps first; personal rules last, non-overriding):")

	case model.CodeSkillHeader:
		return writeLines(w, "", fmt.Sprintf("  %s [%s%s, authority=%s]",
			f.Fields.Str("name"), f.Fields.Str("layer"),
			scopeSuffix(f.Fields.Str("scope")), f.Fields.Str("authority")))

	case model.CodeBaseHeader:
		return writeLines(w, fmt.Sprintf("    layered on base %s:", f.Fields.Str("extends")))

	case model.CodeBaseStep:
		return writeLines(w, "      [base] "+f.Fields.Str("step"))

	case model.CodeDanglingExtendsWarning:
		return writeLines(w, fmt.Sprintf(
			"    [warn] extends %s does not resolve (dangling — validate will fail)",
			f.Fields.Str("extends")))

	case model.CodeStep:
		return writeLines(w, "      "+f.Fields.Str("step"))

	case model.CodePersonalHeader:
		return writeLines(w, "",
			"  personal rules (non-overriding — canonical mandatory steps always win):")

	case model.CodePersonalEntry:
		return writeLines(w, fmt.Sprintf("    [personal:%s] %s",
			f.Fields.Str("team"), f.Fields.Str("name")))

	case model.CodeSummary:
		return writeLines(w, "", fmt.Sprintf("%d skill(s) across %d populated layer(s).",
			f.Fields.Int("skills"), f.Fields.Int("layers")))
	}
	return fmt.Errorf("render: skills: no rule for finding code %q", f.Code)
}

// layerEntry is `:880-889`. The origin tag gains its scope only when the skill
// has one — a company-layer skill belongs to no platform and no team — and a
// personal rule shows a fixed notice in place of the metadata the shared layers
// display.
func layerEntry(f model.Finding) string {
	layer := f.Fields.Str("layer")
	tag := layer
	if scope := f.Fields.Str("scope"); scope != "" {
		tag = layer + ":" + scope
	}
	extra := "  personal rule (non-overriding; canonical mandatory steps win)"
	if layer != string(skills.LayerPersonal) {
		extra = fmt.Sprintf("  id=%s authority=%s",
			f.Fields.Str("id"), f.Fields.Str("authority"))
		// An absent `extends` is absent from the line, not rendered as None —
		// unlike id and authority, which always print (`:886-888`).
		if ext := f.Fields.Str("extends"); ext != "" {
			extra += " extends=" + ext
		}
	}
	return fmt.Sprintf("  [%s] %s%s", tag, f.Fields.Str("name"), extra)
}

func scopeSuffix(scope string) string {
	if scope == "" {
		return ""
	}
	return ":" + scope
}

func writeLines(w io.Writer, lines ...string) error {
	for _, line := range lines {
		if _, err := fmt.Fprintln(w, line); err != nil {
			return err
		}
	}
	return nil
}
