package render

import (
	"fmt"
	"io"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/ids"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/model"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/roles"
)

// IDs writes `ids list` as the Python CLI writes it (bin/company-os:1275-1302).
//
// Every sentence in the command is composed here and nowhere else (R-2.8):
// internal/ids emits a code and typed fields per line, and this function is the
// only place that knows the wording, the two-space indent, or the four-space
// "  ->  " separator between an ID and its defining file.
func IDs(w io.Writer, sections []model.GateResult) error {
	for _, s := range sections {
		if s.Slug == roles.SlugGlossary {
			if err := glossary(w, s); err != nil {
				return err
			}
			continue
		}
		for _, f := range s.Findings {
			var err error
			switch f.Code {
			case ids.CodeRegistryEmpty:
				_, err = fmt.Fprintf(w,
					"no canonical IDs found — %s is missing or empty "+
						"(seed it with: company-os init / add)\n", f.Fields.Str("registry"))
			case ids.CodeListingHeader:
				_, err = fmt.Fprintf(w, "canonical IDs (%s):\n", f.Fields.Str("registry"))
			case ids.CodeEntry:
				err = idEntry(w, f)
			case ids.CodeCount:
				err = idCount(w, f)
			default:
				err = fmt.Errorf("render: ids: no rule for finding code %q", f.Code)
			}
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// idEntry is `:1300`: the "  ->  " tail is suppressed for an entry with no
// definedIn, so a registry row missing that key renders as a bare id.
func idEntry(w io.Writer, f model.Finding) error {
	line := "  " + f.Fields.Str("id")
	if defined := f.Fields.Str("definedIn"); defined != "" {
		line += "  ->  " + defined
	}
	_, err := fmt.Fprintln(w, line)
	return err
}

// idCount is `:1301-1302`. The "of N" suffix appears only when a filter removed
// something, which is a comparison of the two counts rather than a flag: an
// active filter that removes nothing prints the bare tally, as Python's
// `len(rows) != len(entries)` does.
func idCount(w io.Writer, f model.Finding) error {
	matched, total := f.Fields.Int("matched"), f.Fields.Int("total")
	suffix := ""
	if matched != total {
		suffix = fmt.Sprintf(" of %d", total)
	}
	_, err := fmt.Fprintf(w, "%d id(s)%s\n", matched, suffix)
	return err
}
