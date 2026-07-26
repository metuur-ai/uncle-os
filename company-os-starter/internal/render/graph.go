package render

import (
	"fmt"
	"io"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/model"
)

// Graph writes `graph build` as the Python CLI writes it
// (bin/company-os:1787-1797), and — over the subset of sections
// rebuild_generated produces — the derived-artifact lines every scaffolding
// command prints before its own output (`:1803-1810`).
//
// One renderer serves both callers on purpose. The alternative, a second
// line-composing function for the injected scaffold -> graph seam, would put
// the same four sentences in two places and let them drift; here the seam is
// this function writing into a buffer, so a reworded line is reworded once.
func Graph(w io.Writer, sections []model.GateResult) error {
	for _, s := range sections {
		for _, f := range s.Findings {
			var err error
			switch f.Code {
			case model.CodeGraphTagged:
				_, err = fmt.Fprintf(w, "  tagged %s\n", f.Fields.Str("path"))
			case model.CodeGraphIndexWritten:
				_, err = fmt.Fprintf(w, "  wrote index %s\n", f.Fields.Str("path"))
			case model.CodeGraphNodeWritten:
				_, err = fmt.Fprintf(w, "  node %s\n", f.Fields.Str("path"))
			case model.CodeGraphNodeMarkersUnbalanced:
				// warn() writes to stdout with a two-space indent
				// (bin/company-os:50-51). The path here is ABSOLUTE while the
				// three lines above are workspace-relative — warn() interpolates
				// the Path it was handed (`:1664`) and the others call
				// relative_to() first. R-0.8 keeps that asymmetry.
				_, err = fmt.Fprintf(w,
					"  [warn] %s: unbalanced/duplicated company-os:generated markers "+
						"(%d start, %d end) — not rewriting\n",
					f.Fields.Str("path"), f.Fields.Int("starts"), f.Fields.Int("ends"))
			case model.CodeGraphSummary:
				_, err = fmt.Fprintf(w, "graph build: %d doc(s) scanned, %d updated\n",
					f.Fields.Int("scanned"), f.Fields.Int("updated"))
			default:
				err = fmt.Errorf("render: graph: no rule for finding code %q", f.Code)
			}
			if err != nil {
				return err
			}
		}
	}
	return nil
}
