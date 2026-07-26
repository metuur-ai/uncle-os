package render

import (
	"fmt"
	"io"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/model"
)

// Product writes the six product-lifecycle commands as the Python CLI writes
// them: `discover new` (bin/company-os:411-424), `discover validate`
// (`:425-463`), `prd new` (`:574-623`), `prd validate` (`:625-669`),
// `prd complete` (`:671-711`) and `check ready|done` (`:714-733`).
//
// One renderer serves all six because they share one line grammar: a flat block,
// severity markers from the three ok/warn/fail helpers, and per-code indents.
// What they do NOT share with `validate` is gate headers, which is why nothing
// here reads GateResult.Ordinal or Title.
//
// `prd complete` splices rebuild_generated's sections between its own two
// halves, so a section carrying graph codes is delegated to Graph rather than
// rejected. That delegation is what makes the oracle's ordering — archived,
// outcome, appended, THEN the derived lines, THEN the next step — expressible as
// one ordered record set instead of two interleaved writers.
func Product(w io.Writer, sections []model.GateResult) error {
	for _, s := range sections {
		switch s.Slug {
		case model.SectionTags, model.SectionFeatureIndexes, model.SectionClaudeNodes:
			if err := Graph(w, []model.GateResult{s}); err != nil {
				return err
			}
			continue
		}
		for _, f := range s.Findings {
			if err := productLine(w, f); err != nil {
				return err
			}
		}
	}
	return nil
}

func productLine(w io.Writer, f model.Finding) error {
	switch f.Code {

	// Plain print() lines, no indent.
	case model.CodeDiscoveryCreated,
		model.CodePRDCreated,
		model.CodePRDArchived,
		model.CodeOutcomeScheduled,
		model.CodeLogAppended,
		model.CodeDoneCheckHeader,
		model.CodeCheckBaselineHeader,
		model.CodeCheckBaselineText,
		model.CodeCheckChecklist:
		return writeLines(w, f.Message)

	// `print(f"\n== Applicable governance ...")` — the blank line is a literal
	// "\n" inside the f-string (`:725`), so it belongs to this line and to no
	// other in the command.
	case model.CodeCheckGovernanceHeader:
		return writeLines(w, "", f.Message)

	// The next-step lines (R-1.8). Each is one print() whose text begins with
	// the literal "next: ".
	case model.CodeDiscoveryNext,
		model.CodeDiscoveryValidateNext,
		model.CodePRDNext,
		model.CodePRDValidateNext,
		model.CodePRDCompleteNext:
		return writeLines(w, "next: "+f.Message)

	// Two-space indented notes that are not ok()/warn()/fail() lines.
	case model.CodeTemplateSource:
		return writeLines(w, "  "+f.Message)
	case model.CodePRDRealityNote:
		return writeLines(w, "  note: "+f.Message)
	case model.CodeDoneFix:
		return writeLines(w, "  fix: "+f.Message)

	// ok() (`:47-48`).
	case model.CodeDiscoveryValidated, model.CodePRDContractOK:
		return writeLines(w, "  [ok] "+f.Message)

	// warn() (`:50-51`) — stdout, two-space indent, so these lines are part of
	// the view rather than a diagnostic stream.
	case model.CodeCheckBaselineMissing,
		model.CodeCheckUnresolved,
		model.CodePRDGovernanceUnresolved:
		return writeLines(w, "  [warn] "+f.Message)
	}

	// The contract issues are the one family whose severity, not whose code,
	// picks the marker: `section 'X' is empty` renders as a warn when the team
	// has not opted into enforcement and as a fail when it has, off the same
	// record (`:448-453`, `:647-652`).
	switch f.Code {
	case model.CodeCoreTypeMissing,
		model.CodeCoreIdentityMissing,
		model.CodeCoreStatusMissing,
		model.CodeCoreUpdatedMissing,
		model.CodeCoreRoleMissing,
		model.CodeSectionHeadingMissing,
		model.CodeSectionEmpty,
		model.CodePRDProcessField,
		model.CodeDoneChecklistUnchecked,
		model.CodeDoneRealityMissing,
		model.CodeDoneRealityStale,
		model.CodeDoneRealityDateInvalid:
		switch f.Severity {
		case model.SevWarn:
			return writeLines(w, "  [warn] "+f.Message)
		default:
			return writeLines(w, "  [FAIL] "+f.Message)
		}
	}
	return fmt.Errorf("render: product: no rule for finding code %q", f.Code)
}
