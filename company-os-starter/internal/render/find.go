package render

import (
	"fmt"
	"io"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/model"
)

// Find writes the unified-search view (ux-simplification 3.1).
//
// Output is grouped by source with a header per section, then one line per
// hit carrying the path, title (when present), and match reason. The
// graphify section (when present) is appended last and uses a simpler layout
// — its output is either graphify's own lines or a single hint/error.
//
// The renderer switches on Code to choose the layout; the records carry all
// the data. Deterministic ordering is the producer's responsibility — results
// arrive sorted by path within each section.
func Find(w io.Writer, sections []model.GateResult) error {
	for i, s := range sections {
		// Blank line between sections (not before the first).
		if i > 0 {
			if _, err := fmt.Fprintln(w); err != nil {
				return err
			}
		}
		switch s.Slug {
		case model.SlugFindGraphify:
			if err := renderFindGraphify(w, s); err != nil {
				return err
			}
		default:
			if err := renderFindSection(w, s); err != nil {
				return err
			}
		}
	}
	return nil
}

// renderFindSection writes one source section: a header line followed by
// each hit indented under it. The "no matches" finding is rendered as a bare
// message line without a section header.
func renderFindSection(w io.Writer, s model.GateResult) error {
	// "no matches" is a bare message, no header.
	if len(s.Findings) == 1 && s.Findings[0].Code == model.CodeFindNoMatches {
		_, err := fmt.Fprintf(w, "%s\n", s.Findings[0].Message)
		return err
	}
	if _, err := fmt.Fprintf(w, "== %s ==\n", s.Title); err != nil {
		return err
	}
	for _, f := range s.Findings {
		// Format: "  <path>  [<title>]  <reason>"
		// Title is omitted when empty or identical to path.
		title := f.Fields.Str("title")
		reason := f.Fields.Str("reason")
		path := f.Path
		if title != "" && title != path {
			if _, err := fmt.Fprintf(w, "  %s  [%s]  %s\n", path, title, reason); err != nil {
				return err
			}
		} else {
			if _, err := fmt.Fprintf(w, "  %s  %s\n", path, reason); err != nil {
				return err
			}
		}
	}
	return nil
}

// renderFindGraphify writes the graphify hook section. It uses a simpler
// layout: a "graphify:" header followed by each output line or the single
// hint/error message.
func renderFindGraphify(w io.Writer, s model.GateResult) error {
	if len(s.Findings) == 0 {
		return nil
	}
	// Hints and errors are single-line; graphify output lines are prefixed
	// "graphify:" already in the finding message.
	first := s.Findings[0]
	switch first.Code {
	case model.CodeFindGraphifyHint, model.CodeFindGraphifyError:
		_, err := fmt.Fprintf(w, "%s\n", first.Message)
		return err
	case model.CodeFindGraphify:
		if _, err := fmt.Fprintln(w, "graphify:"); err != nil {
			return err
		}
		for _, f := range s.Findings {
			if _, err := fmt.Fprintf(w, "  %s\n", f.Message); err != nil {
				return err
			}
		}
		return nil
	}
	// Fallback: print each finding message.
	for _, f := range s.Findings {
		if _, err := fmt.Fprintf(w, "  %s\n", f.Message); err != nil {
			return err
		}
	}
	return nil
}
