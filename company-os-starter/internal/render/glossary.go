package render

import (
	"fmt"
	"io"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/model"
)

// glossaryIntro is the legend's first line (bin/company-os:1268-1269), wrapped
// here only because the Python source wraps it — the emitted line is one line.
const glossaryIntro = "terms (plain-language, display-only — canonical terms " +
	"are unchanged in artifacts, IDs, tags, and validation):"

// glossary renders the --role legend shared by `today` and `ids list`.
//
// The intro belongs to the section, not to a term: role_glossary_lines returns
// [] for an unmapped role, so an absent section is the whole of "no glossary, no
// error" (GPF-R-3.3) and the caller never has to test the role again.
func glossary(w io.Writer, s model.GateResult) error {
	if len(s.Findings) == 0 {
		return nil
	}
	if _, err := fmt.Fprintln(w, glossaryIntro); err != nil {
		return err
	}
	for _, f := range s.Findings {
		if f.Code != model.CodeGlossaryTerm {
			return fmt.Errorf("render: glossary: no rule for finding code %q", f.Code)
		}
		// The quotes are literal, not %q: Python's f-string interpolates the
		// label raw, while %q would escape a non-ASCII byte in it.
		if _, err := fmt.Fprintf(w, "  %s — \"%s\"\n",
			f.Fields.Str("canonical"), f.Fields.Str("plain")); err != nil {
			return err
		}
	}
	return nil
}
