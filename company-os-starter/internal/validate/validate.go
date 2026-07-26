package validate

// cmd_validate (bin/company-os:922-1107), as records.
//
// This is composition and ordering, and deliberately almost nothing else. Six
// clusters already answer their own gate — internal/governance gates 1 and 2,
// internal/product gate 3, internal/graph gates 5 and 6, internal/skills gate 7,
// internal/federation gate 8 — and each returns a finished model.GateResult, so
// Run's body is a list of calls in the oracle's order.
//
// Gate 4 is the exception and is built here, because it is the one gate that
// spans two clusters: it walks internal/graph's document iterator and applies
// internal/product's core_field_errors to each document. internal/product
// already imports internal/graph, so a gate 4 living in either package would
// need an import the other direction. It composes no prose of its own: each
// finding's sentence comes from the Message of whichever package owns the fact.

import (
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/federation"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/governance"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/graph"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/model"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/product"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/skills"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/workspace"
)

// FrontmatterGateSlug and FrontmatterGateTitle name gate 4. The other seven
// gates' identities belong to the packages that produce them.
const (
	FrontmatterGateSlug  = "frontmatter-tags"
	FrontmatterGateTitle = "frontmatter core and tag derivation (interop contract)"
)

// Run is cmd_validate: the workspace banner followed by seven or eight gates.
//
// The first section is the banner (`:924`) rather than a gate. It carries the
// workspace root, which is the one thing in the output that no gate can derive,
// and the [N/M] denominator — 7 without a federation manifest, 8 with (`:930`).
// The renderer identifies it by model.SlugWorkspace so it stays out of the gate
// list. Gate 8 is self-suppressing on purpose: in monorepo mode it does not
// exist at all, which is what keeps the monorepo golden byte-identical.
//
// The denominator is CARRIED on that record rather than derived from the length
// of the gate list (R-2.6a). `:930` decides it from manifest presence before any
// gate runs, and the oracle then prints every header against that one number, so
// the count is a property of the run and not of how far the run got.
//
// Every gate runs even when an earlier one failed. There is no early return in
// the oracle and adding one would hide problems behind problems — a gate
// FAILURE is a finding, not an error.
//
// A malformed artifact is different: it aborts the run. Even then the banner is
// returned WITH the error, because `:924` prints it before `:929` loads the
// manifest, so the oracle has already written that line to stdout by the time it
// dies — measured on a workspace.yaml with a non-list `repos:`. The banner
// carries `complete: false` in that case, which is what stops the renderer
// printing a PASS trailer over an aborted run.
func Run(ws *workspace.Workspace) ([]model.GateResult, error) {
	banner := func(complete bool, total int) model.GateResult {
		return model.GateResult{
			Ordinal: 0,
			Slug:    model.SlugWorkspace,
			Findings: []model.Finding{{
				Severity: model.SevOK,
				Code:     model.CodeValidateRoot,
				Fields: model.Fields{
					"root": ws.Root, "complete": complete, "gates": total,
				},
			}},
		}
	}

	manifest, err := federation.LoadManifest(ws)
	if err != nil {
		// The manifest load is what decides 7-versus-8, so a run that dies here
		// has no denominator to report. It also printed no gate header: `:930`
		// precedes `[1/ng]`, so there is nothing for the count to number.
		return []model.GateResult{banner(false, 0)}, err
	}

	steps := []func(int) (model.GateResult, error){
		func(n int) (model.GateResult, error) { return governance.OwnershipGate(ws, n) },
		func(n int) (model.GateResult, error) { return governance.ExpiryGate(ws, n) },
		func(n int) (model.GateResult, error) { return product.Gate(ws, n) },
		func(n int) (model.GateResult, error) { return frontmatterGate(ws, n) },
		func(n int) (model.GateResult, error) { return graph.NodeGate(ws, n) },
		func(n int) (model.GateResult, error) { return graph.FeatureIndexGate(ws, n) },
		func(n int) (model.GateResult, error) { return skills.Gate(ws, n) },
	}
	if manifest != nil {
		steps = append(steps, func(n int) (model.GateResult, error) {
			return federation.Gate(ws, manifest, n)
		})
	}

	total := len(steps)
	out := []model.GateResult{banner(true, total)}
	for i, step := range steps {
		g, err := step(i + 1)
		if err != nil {
			// A mid-run abort keeps everything the oracle had already printed: the
			// banner, every completed gate, and the aborting gate's own header and
			// findings-so-far. Each gate prints its header before its first check
			// (`:936`, `:954`, `:975`, …) and every producer returns its partial
			// GateResult alongside the error, so the record set is exactly the
			// oracle's truncated stdout.
			//
			// This is only correct because `total` is carried (R-2.6a). Derived
			// from `out`, an abort inside gate 2 would render `[1/2]`/`[2/2]`
			// against the oracle's `[1/7]`/`[2/7]`; the port used to drop the
			// completed gates rather than print that, which lost six human-facing
			// lines and collided with R-0.8. Neither is needed — the denominator
			// was decided before gate 1 ran.
			//
			// `exception/garbage-expires` is the corpus case: the oracle crashes
			// inside gate 2 at `:970` having printed gate 1 and three of gate 2's
			// four lines, and the port now reproduces that stdout byte-for-byte.
			// Only its stderr stays a declared divergence, Python's being a
			// ValueError traceback.
			//
			// The banner is rewritten to `complete: false` — it was built optimistic
			// and the run has now not reached a verdict, which is what suppresses the
			// PASS/FAIL trailer the oracle never printed.
			out[0] = banner(false, total)
			return append(out, g), err
		}
		out = append(out, g)
	}
	return out, nil
}

// frontmatterGate is gate 4 (`:996-1013`): the OKF/Obsidian interop contract,
// checked over every graph document. Body format is never inspected here.
//
// Three shapes in it are easy to lose and each is fixed by a golden:
//
//   - The `[ok]` is CONDITIONAL. A document carrying core-field errors emits
//     those and no ok line (`:1003-1008`). No model support is needed — the
//     producer just does not append the record.
//   - The pointer check is a LOOP, not a multi-line finding. One document with
//     four malformed pointers yields four separate one-line warnings, emitted
//     after that document's own ok/fail and before the next document's.
//   - Severities interleave in DOCUMENT order within the one Findings slice and
//     must never be bucketed.
func frontmatterGate(ws *workspace.Workspace, ordinal int) (model.GateResult, error) {
	g := model.GateResult{Ordinal: ordinal, Slug: FrontmatterGateSlug, Title: FrontmatterGateTitle}
	docs, err := graph.IterGraphDocs(ws)
	if err != nil {
		return g, err
	}
	for _, d := range docs {
		core := product.CoreFieldErrors(d.Meta)
		for _, issue := range core {
			// The declared code for this render site is one code (model/codes.go
			// § gate 4). The decomposed issue code rides in Fields so a machine
			// consumer keeps the resolution the sentence has, and the sentence
			// itself is product's, not this package's.
			fields := model.Fields{"path": d.Rel, "issue": issue.Code}
			for k, v := range issue.Fields {
				fields[k] = v
			}
			g.Findings = append(g.Findings, model.Finding{
				Severity: model.SevFail,
				Code:     model.CodeFrontmatterCoreField,
				Subject:  d.Rel,
				Path:     d.Rel,
				Message:  product.Message(issue.Code, issue.Fields),
				Fields:   fields,
			})
		}

		inSync, err := graph.TagsInSync(d.Meta, d.Tags)
		if err != nil {
			return g, err
		}
		switch {
		case !inSync:
			g.Findings = append(g.Findings, docFinding(model.SevFail,
				model.CodeTagsDrift, d.Rel, model.Fields{"path": d.Rel, "tags": d.Tags}))
		case len(core) == 0:
			g.Findings = append(g.Findings, docFinding(model.SevOK,
				model.CodeFrontmatterInSync, d.Rel, model.Fields{"path": d.Rel}))
		}

		// Pointer well-formedness is guidance-tier and never blocks (R-1.3), so
		// these are warns and are excluded from the trailer count.
		for _, pe := range graph.PointerErrors(d.Meta) {
			g.Findings = append(g.Findings, docFinding(model.SevWarn,
				model.CodePointerGuidance, d.Rel, model.Fields{"path": d.Rel, "problem": pe}))
		}
	}
	return g, nil
}

// docFinding builds one gate-4 record whose sentence internal/graph owns.
func docFinding(sev model.Severity, code, rel string, f model.Fields) model.Finding {
	return model.Finding{
		Severity: sev,
		Code:     code,
		Subject:  rel,
		Path:     rel,
		Message:  graph.Message(code, f),
		Fields:   f,
	}
}
