package product

// The one function in this package that produces human prose.
//
// Message is a pure function of (code, Fields) and nothing else (R-2.8, R-2.12):
// every number, path and identifier in a sentence is read back out of Fields
// here, so a field that stops being populated changes the rendered line rather
// than going quietly missing. internal/render calls it; nothing above composes a
// sentence.

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/model"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/yamlio"
)

// uncheckedRe is `r"- \[ \] .*"` (`bin/company-os:678`) — the governance
// checklist items still carrying no evidence. `.` excludes newlines in Python
// without re.DOTALL, and Go's default is the same, so the two agree per line.
var uncheckedRe = regexp.MustCompile(`- \[ \] .*`)

// ErrDoneCheck is `prd complete`'s done-gate refusal (`:699-707`).
//
// It is QUIET: the oracle prints the whole refusal block to stdout and writes
// nothing to stderr, so the dispatcher renders PRDComplete's records and adds no
// `error: …` line of its own. The code is 5, not 1 — see the package comment.
var ErrDoneCheck = model.Quiet(model.Errorf(model.ExitPrecondition,
	"done-check failed — a change is not done until reality is updated"))

func okFinding(code, subject, path string, f model.Fields) model.Finding {
	if f == nil {
		f = model.Fields{}
	}
	return model.Finding{
		Severity: model.SevOK,
		Code:     code,
		Subject:  subject,
		Path:     path,
		Message:  Message(code, f),
		Fields:   f,
	}
}

func warnFinding(code, subject string, f model.Fields) model.Finding {
	if f == nil {
		f = model.Fields{}
	}
	return model.Finding{
		Severity: model.SevWarn,
		Code:     code,
		Subject:  subject,
		Message:  Message(code, f),
		Fields:   f,
	}
}

// Message composes one product line's sentence from its code and its typed
// fields, and from nothing else.
func Message(code string, f model.Fields) string {
	switch code {

	// ---------------------------------------------------- core_field_errors
	case model.CodeCoreTypeMissing:
		return "frontmatter core: 'type' missing"
	case model.CodeCoreIdentityMissing:
		return "frontmatter core: no identity field ('id', or 'prd' for outcome reviews)"
	case model.CodeCoreStatusMissing:
		return fmt.Sprintf("frontmatter core: '%s' doc has no 'status' (lifecycle)", f.Str("type"))
	case model.CodeCoreUpdatedMissing:
		return "frontmatter core: reality doc has no 'updated' date (the done-gate reads it)"
	case model.CodeCoreRoleMissing:
		return "frontmatter core: onboarding-guide has no 'role'"

	// ------------------------------------------------------ section contract
	case model.CodeSectionHeadingMissing:
		return fmt.Sprintf("required section heading '## %s' missing", f.Str("section"))
	case model.CodeSectionEmpty:
		bare := fmt.Sprintf("section '%s' is empty", f.Str("section"))
		if enforced, _ := f["enforced"].(bool); enforced {
			return bare
		}
		return bare + " — format guidance only; the team may use its own " +
			"structure (opt in via standards/doc-formats.yaml)"

	// ------------------------------------------------------------- discover
	case model.CodeDiscoveryCreated:
		return "created " + f.Str("path")
	case model.CodeTemplateSource:
		return "template: " + f.Str("source")
	case model.CodeDiscoveryNext:
		return fmt.Sprintf("fill Problem signal, Hypothesis, Success criteria, then run: "+
			"company-os discover validate --team %s %s", f.Str("team"), f.Str("brief"))
	case model.CodeDiscoveryValidated:
		return fmt.Sprintf("brief '%s' validated (status: validated)", f.Str("brief"))
	case model.CodeDiscoveryValidateNext:
		return fmt.Sprintf("company-os prd new --team %s --from-discovery %s "+
			"--platform <platform-id> --components <comp-id,...>",
			f.Str("team"), f.Str("brief"))

	// ------------------------------------------------------------------ prd
	case model.CodePRDGovernanceUnresolved:
		return fmt.Sprintf("component '%s' has no resolved governance "+
			"(not in team ownership? run governance resolve)", f.Str("component"))
	case model.CodePRDCreated:
		return "created " + f.Str("path")
	case model.CodePRDRealityNote:
		return fmt.Sprintf("component '%s' has no reality doc yet — scaffold it: "+
			"company-os reality new --platform %s %s",
			f.Str("component"), f.Str("platform"), f.Str("component"))
	case model.CodePRDNext:
		return fmt.Sprintf("fill Proposed change + decisionOwner, then: "+
			"company-os prd validate --platform %s %s", f.Str("platform"), f.Str("prd"))
	case model.CodePRDProcessField:
		return fmt.Sprintf("process contract field '%s' missing or TODO", f.Str("field"))
	case model.CodePRDContractOK:
		return fmt.Sprintf("PRD '%s' passes the process contract", f.Str("prd"))
	case model.CodePRDValidateNext:
		return fmt.Sprintf("deliver the change, update the reality doc for each component, "+
			"then: company-os prd complete --platform %s %s", f.Str("platform"), f.Str("prd"))

	// ------------------------------------------------------------ done-gate
	case model.CodeDoneCheckHeader:
		return "done-check failed — a change is not done until reality is updated:"
	case model.CodeDoneChecklistUnchecked:
		return fmt.Sprintf("%d governance checklist item(s) unchecked "+
			"(check them off with linked evidence)", f.Int("count"))
	case model.CodeDoneRealityMissing:
		return fmt.Sprintf("no reality doc for component '%s' (%s)",
			f.Str("component"), f.Str("path"))
	case model.CodeDoneRealityStale:
		return fmt.Sprintf("reality doc for '%s' not updated since PRD created "+
			"(reality updated: %s)", f.Str("component"), f.Str("updated"))
	case model.CodeDoneRealityDateInvalid:
		// R-1.14 / OKF v0.2 R-0.2: name the file AND the offending value, so a
		// malformed date is actionable rather than an invisible pass.
		return fmt.Sprintf("cannot compare dates: %s has '%s: %s', which is not an "+
			"ISO-8601 date (YYYY-MM-DD)", f.Str("path"), f.Str("field"), f.Str("value"))
	case model.CodeDoneFix:
		return fmt.Sprintf("company-os reality new --platform %s %s",
			f.Str("platform"), f.Str("component"))
	case model.CodePRDArchived:
		return "archived -> " + f.Str("path")
	case model.CodeOutcomeScheduled:
		return fmt.Sprintf("outcome review scheduled (due %s)", f.Str("due"))
	case model.CodeLogAppended:
		return "appended " + f.Str("path")
	case model.CodePRDCompleteNext:
		return "company-os validate"

	// ---------------------------------------------------------------- check
	case model.CodeCheckBaselineHeader:
		return fmt.Sprintf("== Team baseline (%s) ==", f.Str("file"))
	case model.CodeCheckBaselineText:
		return f.Str("text")
	case model.CodeCheckBaselineMissing:
		return "no team baseline file"
	case model.CodeCheckGovernanceHeader:
		return fmt.Sprintf("== Applicable governance (%s) ==",
			strings.Join(f.Strs("components"), ", "))
	case model.CodeCheckChecklist:
		return f.Str("markdown")
	case model.CodeCheckUnresolved:
		return fmt.Sprintf("'%s' not in team governance — run governance resolve",
			f.Str("component"))

	// -------------------------------------------------- validate, gate 3
	case model.CodePRDFrontmatterMissing:
		// The list renders as a PYTHON list literal — `['team', 'components']`
		// — because the oracle interpolates the list object into an f-string
		// (`:990`). golden-validate line 16 freezes those quotes and that
		// ", " spacing.
		return "missing frontmatter " + yamlio.PyStrings(f.Strs("missing"))
	case model.CodePRDContractPresent:
		return "contract fields present"
	}
	return ""
}
