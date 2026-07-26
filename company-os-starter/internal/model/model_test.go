package model_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/model"
)

// renderText is a TEST-LOCAL renderer. The real one is task 3.2; this one exists
// only to prove that the record model carries enough information to reproduce a
// golden without any out-of-band state. If a shape cannot be rendered from a
// model.Report alone, this file stops compiling or the comparison fails — which
// is what makes task 1.7 verifiable rather than speculative.
//
// It deliberately implements ONE uniform prefix rule (Subject, then ": ") to
// prove the claim that the seven "distinct prefix shapes" in the LLD are seven
// distinct Subject *values*, not seven renderer branches.
func renderText(r model.Report) string {
	var b strings.Builder
	fmt.Fprintf(&b, "validating workspace %s\n\n", r.Root)
	for _, g := range r.Gates {
		// The leading blank line belongs to the gate header and is present on
		// every gate except the first (R-2.6) — derived from Ordinal, not stored.
		if g.Ordinal > 1 {
			b.WriteString("\n")
		}
		fmt.Fprintf(&b, "[%d/%d] %s\n", g.Ordinal, len(r.Gates), g.Title)
		for _, f := range g.Findings {
			marker := map[model.Severity]string{
				model.SevOK: "ok", model.SevWarn: "warn", model.SevFail: "FAIL",
			}[f.Severity]
			line := f.Message
			if f.Subject != "" {
				line = f.Subject + ": " + f.Message
			}
			fmt.Fprintf(&b, "  [%s] %s\n", marker, line)
		}
	}
	if n := r.Problems(); n > 0 {
		fmt.Fprintf(&b, "\nFAIL — %d problem(s)\n", n)
	} else {
		b.WriteString("\nPASS\n")
	}
	return b.String()
}

func golden(t *testing.T, name string) string {
	t.Helper()
	p := filepath.Join("..", "..", "..", "examples", name)
	data, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("reading oracle %s: %v", p, err)
	}
	return string(data)
}

func ok(subject, code, msg string, f model.Fields) model.Finding {
	return model.Finding{Severity: model.SevOK, Code: code, Subject: subject, Message: msg, Fields: f}
}

func fail(subject, code, msg string, f model.Fields) model.Finding {
	return model.Finding{Severity: model.SevFail, Code: code, Subject: subject, Message: msg, Fields: f}
}

func warn(subject, code, msg string, f model.Fields) model.Finding {
	return model.Finding{Severity: model.SevWarn, Code: code, Subject: subject, Message: msg, Fields: f}
}

// coreOK is the gate-4 in-sync finding, which is by far the most repeated line
// in both passing goldens.
func coreOK(path string) model.Finding {
	return model.Finding{
		Severity: model.SevOK,
		Code:     model.CodeFrontmatterInSync,
		Subject:  path,
		Path:     path,
		Message:  "core fields + tags in sync",
	}
}

// TestPassingGoldenIsReproducibleFromRecords is the R-2.1 proof: gate 3 in
// examples/golden-validate.txt:11-12 is a header with zero findings under it. A
// flat []Finding cannot express it; a GateResult with an empty Findings slice
// can, and this test fails byte-for-byte if that ever regresses.
func TestPassingGoldenIsReproducibleFromRecords(t *testing.T) {
	r := model.Report{
		Root: "<WORKSPACE>",
		Gates: []model.GateResult{
			{Ordinal: 1, Slug: "ownership-reconciliation", Title: "ownership reconciliation", Findings: []model.Finding{
				ok("customer-notification-service", model.CodeOwnershipAgrees,
					"registry and descriptor agree (communications)",
					model.Fields{"component": "customer-notification-service", "platform": "communications"}),
			}},
			{Ordinal: 2, Slug: "governance-expiry", Title: "deviation and exception expiry", Findings: []model.Finding{
				ok("customer-engagement", model.CodeDeviationCurrent,
					"deviation platform-standard://communications/prd-structure current (review 2035-01-15)",
					model.Fields{"team": "customer-engagement", "rule": "platform-standard://communications/prd-structure", "reviewDate": "2035-01-15"}),
				ok("customer-engagement", model.CodeDeviationCurrent,
					"deviation company-standard://estimation/story-points current (review 2035-01-14)",
					model.Fields{"team": "customer-engagement", "rule": "company-standard://estimation/story-points", "reviewDate": "2035-01-14"}),
				ok("customer-engagement", model.CodeExceptionValid,
					"exception platform-standard://communications/message-schema valid until 2035-12-31",
					model.Fields{"team": "customer-engagement", "rule": "platform-standard://communications/message-schema", "expires": "2035-12-31"}),
			}},
			// The whole point of R-2.1: this gate ran and found nothing.
			{Ordinal: 3, Slug: "prd-contracts", Title: "active PRD contracts"},
			{Ordinal: 4, Slug: "frontmatter-tags", Title: "frontmatter core and tag derivation (interop contract)", Findings: []model.Finding{
				coreOK("company-os/onboarding/developer.md"),
				coreOK("company-os/skills/syncing-knowledge.SKILL.md"),
				coreOK("platforms/communications/archive/prds/2026-per-channel-quiet-hours/outcome.md"),
				coreOK("platforms/communications/archive/prds/2026-per-channel-quiet-hours/prd.md"),
				coreOK("platforms/communications/reality/components/customer-notification-service.md"),
				coreOK("platforms/communications/skills/creating-prd.SKILL.md"),
				coreOK("teams/customer-engagement/onboarding/developer.md"),
				coreOK("teams/customer-engagement/product/discovery/2026-per-channel-quiet-hours/brief.md"),
				coreOK("teams/customer-engagement/standards/definition-of-done.md"),
				coreOK("teams/customer-engagement/standards/definition-of-ready.md"),
				coreOK("company-ontology/concepts/capability--message-delivery.md"),
				coreOK("company-ontology/concepts/component.md"),
				coreOK("company-ontology/context-maps/crm-to-communications.md"),
				coreOK("company-ontology/contexts/communications.md"),
			}},
			{Ordinal: 5, Slug: "claude-node-drift", Title: "CLAUDE.md context node drift (fail-safe, absence-tolerant)", Findings: []model.Finding{
				ok("company-os/CLAUDE.md", model.CodeNodeInSync, "context node in sync", nil),
				ok("platforms/communications/CLAUDE.md", model.CodeNodeInSync, "context node in sync", nil),
				ok("teams/customer-engagement/CLAUDE.md", model.CodeNodeInSync, "context node in sync", nil),
				ok("company-ontology/CLAUDE.md", model.CodeNodeInSync, "context node in sync", nil),
			}},
			{Ordinal: 6, Slug: "feature-index-drift", Title: "feature-index drift (derived component->artifact map)", Findings: []model.Finding{
				ok("communications", model.CodeFeatureIndexInSync,
					"feature-index in sync (1 component(s))",
					model.Fields{"platform": "communications", "components": 1}),
			}},
			{Ordinal: 7, Slug: "skills-layering", Title: "custom skills layering (shadowing + extends resolution)", Findings: []model.Finding{
				ok("", model.CodeSkillsClean,
					"skills layered cleanly (2 canonical, 0 team; no shadowing or dangling extends)",
					model.Fields{"canonical": 2, "team": 0}),
			}},
		},
	}

	if got, want := renderText(r), golden(t, "golden-validate.txt"); got != want {
		t.Errorf("record set does not reproduce the passing golden\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
	if n := r.Problems(); n != 0 {
		t.Errorf("Problems() = %d, want 0", n)
	}
}

// TestFailingGoldenIsReproducibleFromRecords is the load-bearing one. It covers
// every fail() site in gates 1-7 plus the warn site, and with it the four shapes
// the sketch did not obviously handle:
//
//   - gate 1's three Subject shapes, including :946's single-quoted component id
//   - gate 3's ORDERED list field, which map[string]string cannot hold
//   - the :1013 warn loop, interleaved with ok and FAIL in document order
//   - gate 4's and gate 6's conditional [ok] — expressed by omitting a Finding
func TestFailingGoldenIsReproducibleFromRecords(t *testing.T) {
	const brief = "teams/ghost/product/discovery/2035-bad-pointers/brief.md"
	r := model.Report{
		Root: "<WORKSPACE>",
		Gates: []model.GateResult{
			{Ordinal: 1, Slug: "ownership-reconciliation", Title: "ownership reconciliation", Findings: []model.Finding{
				// :941 — Subject is the TEAM id, not the component id.
				fail("ghost", model.CodeOwnershipDescriptorMissing,
					"owns 'no-such-service' but no descriptor in any platform catalog",
					model.Fields{"team": "ghost", "component": "no-such-service"}),
				// :946 — Subject is the component id WRAPPED IN SINGLE QUOTES.
				// Fields carries the unquoted value for JSON consumers.
				fail("'svc-alpha'", model.CodeOwnershipAccountableMismatch,
					"team 'ghost' claims accountable but descriptor says 'team://other' "+
						"(single-source rule: descriptor ownership must match team registry)",
					model.Fields{"component": "svc-alpha", "team": "ghost", "accountableTeam": "team://other"}),
				ok("svc-beta", model.CodeOwnershipAgrees, "registry and descriptor agree (beta)",
					model.Fields{"component": "svc-beta", "platform": "beta"}),
			}},
			{Ordinal: 2, Slug: "governance-expiry", Title: "deviation and exception expiry", Findings: []model.Finding{
				fail("ghost", model.CodeDeviationExpired,
					"deviation for company-standard://estimation/story-points expired 2020-01-01 — re-review or remove",
					model.Fields{"team": "ghost", "rule": "company-standard://estimation/story-points", "reviewDate": "2020-01-01"}),
				ok("ghost", model.CodeDeviationCurrent,
					"deviation company-standard://retro-cadence current (review 2035-01-14)",
					model.Fields{"team": "ghost", "rule": "company-standard://retro-cadence", "reviewDate": "2035-01-14"}),
				fail("ghost", model.CodeExceptionNoExpiry,
					"exception for company-standard://security-service-baseline has NO expiry — invalid",
					model.Fields{"team": "ghost", "rule": "company-standard://security-service-baseline"}),
				fail("ghost", model.CodeExceptionExpired,
					"exception for company-standard://customer-data-privacy expired 2020-01-01",
					model.Fields{"team": "ghost", "rule": "company-standard://customer-data-privacy", "expires": "2020-01-01"}),
				ok("ghost", model.CodeExceptionValid,
					"exception company-standard://tier-1-observability valid until 2035-12-31",
					model.Fields{"team": "ghost", "rule": "company-standard://tier-1-observability", "expires": "2035-12-31"}),
			}},
			{Ordinal: 3, Slug: "prd-contracts", Title: "active PRD contracts", Findings: []model.Finding{
				// The ordered []string is the reason Fields cannot be map[string]string.
				fail("alpha/2035-broken-contract", model.CodePRDFrontmatterMissing,
					"missing frontmatter ['team', 'components', 'governanceSnapshot']",
					model.Fields{"platform": "alpha", "prd": "2035-broken-contract",
						"missing": []string{"team", "components", "governanceSnapshot"}}),
				ok("beta/2035-beta-change", model.CodePRDContractPresent, "contract fields present",
					model.Fields{"platform": "beta", "prd": "2035-beta-change"}),
			}},
			{Ordinal: 4, Slug: "frontmatter-tags", Title: "frontmatter core and tag derivation (interop contract)", Findings: []model.Finding{
				coreOK("company-os/skills/creating-prd.SKILL.md"),
				coreOK("platforms/alpha/change-records/active/2035-broken-contract/prd.md"),
				// Conditional [ok] (:1003-1008): this document has a core-field
				// error, so it emits its [FAIL] and NO [ok]. No model support is
				// needed — the producer simply does not append the ok Finding.
				fail("platforms/alpha/reality/components/svc-alpha.md", model.CodeFrontmatterCoreField,
					"frontmatter core: reality doc has no 'updated' date (the done-gate reads it)",
					model.Fields{"field": "updated"}),
				coreOK("platforms/beta/archive/prds/2035-old-dirname/outcome.md"),
				coreOK("platforms/beta/archive/prds/2035-old-dirname/prd.md"),
				coreOK("platforms/beta/change-records/active/2035-beta-change/prd.md"),
				// The :1013 warn loop: one document, four warn lines, each its own
				// Finding, emitted immediately after that document's [ok].
				coreOK(brief),
				warn(brief, model.CodePointerGuidance, "pointers[0]: missing 'label' — pointer guidance (not blocking)",
					model.Fields{"index": 0, "problem": "missing 'label'"}),
				warn(brief, model.CodePointerGuidance, "pointers[1]: missing 'system' — pointer guidance (not blocking)",
					model.Fields{"index": 1, "problem": "missing 'system'"}),
				warn(brief, model.CodePointerGuidance, "pointers[1]: needs 'url' or 'id' — pointer guidance (not blocking)",
					model.Fields{"index": 1, "problem": "needs 'url' or 'id'"}),
				warn(brief, model.CodePointerGuidance, "pointers[2]: must be a mapping — pointer guidance (not blocking)",
					model.Fields{"index": 2, "problem": "must be a mapping"}),
				fail("teams/ghost/product/discovery/2035-drifted-tags/brief.md", model.CodeTagsDrift,
					"committed tags drifted from frontmatter derivation — run: company-os graph build", nil),
				coreOK("teams/ghost/skills/creating-prd.SKILL.md"),
				coreOK("teams/ghost/skills/reviewing-prd.SKILL.md"),
			}},
			// Gate 5's three Subject shapes: <root>/CLAUDE.md, <root>/team.yaml,
			// and the bare <root>. One renderer rule, three producer values.
			{Ordinal: 5, Slug: "claude-node-drift", Title: "CLAUDE.md context node drift (fail-safe, absence-tolerant)", Findings: []model.Finding{
				ok("company-os/CLAUDE.md", model.CodeNodeHandOwned, "hand-owned, no generated markers (-> pass)", nil),
				fail("platforms/alpha/CLAUDE.md", model.CodeNodeDrift,
					"generated block drifted — run: company-os graph build", nil),
				ok("platforms/beta/CLAUDE.md", model.CodeNodeInSync, "context node in sync", nil),
				fail("teams/ghost/team.yaml", model.CodeNodeIdentity, "roster[1]: needs 'name' and 'role'",
					model.Fields{"index": 1}),
				ok("teams/ghost", model.CodeNodeAbsent, "no CLAUDE.md node (absent -> pass)", nil),
				ok("company-ontology/CLAUDE.md", model.CodeNodeInSync, "context node in sync", nil),
			}},
			{Ordinal: 6, Slug: "feature-index-drift", Title: "feature-index drift (derived component->artifact map)", Findings: []model.Finding{
				fail("alpha", model.CodeFeatureIndexDrift,
					"feature-index drifted from derivation — run: company-os graph build",
					model.Fields{"platform": "alpha"}),
				fail("beta", model.CodeFeatureIndexUnresolved,
					"feature-index component 'svc-beta' references discovery '2035-no-such-brief' which resolves to no document",
					model.Fields{"platform": "beta", "component": "svc-beta", "kind": "discovery", "ref": "2035-no-such-brief"}),
				fail("beta", model.CodeFeatureIndexUnresolved,
					"feature-index component 'svc-beta' references prd '2035-renamed-change' which resolves to no document",
					model.Fields{"platform": "beta", "component": "svc-beta", "kind": "prd", "ref": "2035-renamed-change"}),
			}},
			// Gates 7 and 8 use no prefix at all: Subject is empty.
			{Ordinal: 7, Slug: "skills-layering", Title: "custom skills layering (shadowing + extends resolution)", Findings: []model.Finding{
				fail("", model.CodeSkillShadowing,
					"skill shadowing: teams/ghost/skills/creating-prd.SKILL.md reuses the canonical "+
						"id 'skill://product/creating-prd' of company-os/skills/creating-prd.SKILL.md — "+
						"extend it with `extends: platform-skill://...` instead of replacing it",
					model.Fields{"skill": "teams/ghost/skills/creating-prd.SKILL.md",
						"shadows": "company-os/skills/creating-prd.SKILL.md",
						"reason":  "id 'skill://product/creating-prd'"}),
				fail("", model.CodeSkillDanglingExtends,
					"dangling extends: teams/ghost/skills/reviewing-prd.SKILL.md declares "+
						"extends: platform-skill://alpha/no-such-base but no such base skill exists",
					model.Fields{"skill": "teams/ghost/skills/reviewing-prd.SKILL.md",
						"extends": "platform-skill://alpha/no-such-base"}),
			}},
		},
	}

	if got, want := renderText(r), golden(t, "failing-workspace-golden-validate.txt"); got != want {
		t.Errorf("record set does not reproduce the failing golden\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
	// The trailer count is derived from the records, not tracked alongside them:
	// 15 [FAIL] lines, and the 4 [warn] lines deliberately do not count.
	if n := r.Problems(); n != 15 {
		t.Errorf("Problems() = %d, want 15 (warns must not count)", n)
	}
}

// TestFederatedGoldenAddsGateEight proves the dynamic denominator is derived
// from the gate list rather than stored: the same records render [N/7] or [N/8]
// purely by whether gate 8 is present.
func TestFederatedGoldenAddsGateEight(t *testing.T) {
	r := model.Report{
		Root: "<WORKSPACE>",
		Gates: []model.GateResult{
			{Ordinal: 1, Slug: "ownership-reconciliation", Title: "ownership reconciliation"},
			{Ordinal: 2, Slug: "governance-expiry", Title: "deviation and exception expiry"},
			{Ordinal: 3, Slug: "prd-contracts", Title: "active PRD contracts"},
			{Ordinal: 4, Slug: "frontmatter-tags", Title: "frontmatter core and tag derivation (interop contract)"},
			{Ordinal: 5, Slug: "claude-node-drift", Title: "CLAUDE.md context node drift (fail-safe, absence-tolerant)", Findings: []model.Finding{
				ok("platforms/sliced-alpha", model.CodeNodeAbsent, "no CLAUDE.md node (absent -> pass)", nil),
			}},
			{Ordinal: 6, Slug: "feature-index-drift", Title: "feature-index drift (derived component->artifact map)", Findings: []model.Finding{
				ok("sliced-alpha", model.CodeFeatureIndexAbsent,
					"no feature-index (absent -> pass; run graph build to enable)",
					model.Fields{"platform": "sliced-alpha"}),
			}},
			{Ordinal: 7, Slug: "skills-layering", Title: "custom skills layering (shadowing + extends resolution)", Findings: []model.Finding{
				ok("", model.CodeSkillsClean,
					"skills layered cleanly (0 canonical, 0 team; no shadowing or dangling extends)",
					model.Fields{"canonical": 0, "team": 0}),
			}},
			{Ordinal: 8, Slug: "federated-slice-integrity", Title: "federated slice integrity (read-only derived content)", Findings: []model.Finding{
				fail("", model.CodeSliceSetDrift,
					"repo 'sliced-alpha': slice set in workspace.yaml differs from workspace.lock.yaml "+
						"(a target or allowlist changed without a re-sync) — run: company-os workspace sync",
					model.Fields{"repo": "sliced-alpha"}),
				fail("", model.CodeSliceHandEdited,
					"federated slice hand-edited: platforms/sliced-alpha/governance/requirements.yaml — "+
						"content hash differs from workspace.lock.yaml; slices are read-only derived "+
						"content — re-sync: company-os workspace sync",
					model.Fields{"repo": "sliced-alpha"}),
				fail("", model.CodeSliceFileMissing,
					"federated slice file missing: platforms/sliced-alpha/components/svc-sliced.yaml "+
						"(recorded in workspace.lock.yaml) — re-sync: company-os workspace sync", nil),
				fail("", model.CodeRepoNotLocked,
					"repo 'never-synced' in workspace.yaml has no workspace.lock.yaml entry "+
						"(lock does not cover the manifest) — run: company-os workspace sync",
					model.Fields{"repo": "never-synced"}),
			}},
		},
	}
	if got, want := renderText(r), golden(t, "failing-federated-golden-validate.txt"); got != want {
		t.Errorf("record set does not reproduce the federated failing golden\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

// TestFieldsReachTheTextRenderer is R-2.3: Fields is not JSON-only. Every count
// that appears inside a human [ok] line must be recoverable from the record with
// its original type, and gate 3's ordered list must survive as a list.
func TestFieldsReachTheTextRenderer(t *testing.T) {
	f := model.Fields{
		"components": 1,
		"canonical":  2,
		"team":       0,
		"platform":   "communications",
		"missing":    []string{"team", "components", "governanceSnapshot"},
	}
	if got := f.Int("components"); got != 1 {
		t.Errorf("Int(components) = %d, want 1", got)
	}
	if got := f.Str("platform"); got != "communications" {
		t.Errorf("Str(platform) = %q, want communications", got)
	}
	// Order within a list value is load-bearing: it is rendered verbatim.
	got := f.Strs("missing")
	want := []string{"team", "components", "governanceSnapshot"}
	if len(got) != len(want) {
		t.Fatalf("Strs(missing) = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("Strs(missing) = %v, want %v (order is load-bearing)", got, want)
		}
	}
	// Missing and mistyped keys must be inert, never panic, so a renderer bug is
	// a wrong line rather than a crashed CLI.
	if f.Int("nope") != 0 || f.Str("nope") != "" || f.Strs("nope") != nil {
		t.Error("absent keys must return zero values")
	}
	if f.Int("platform") != 0 || f.Str("components") != "" {
		t.Error("mistyped reads must return zero values")
	}
	var nilFields model.Fields
	if nilFields.Int("x") != 0 || nilFields.Str("x") != "" || nilFields.Strs("x") != nil {
		t.Error("a nil Fields must be readable")
	}
}

// TestEveryRenderSiteHasADistinctCode is R-2.4: a stable machine code per site
// that does not move when the human message is reworded.
func TestEveryRenderSiteHasADistinctCode(t *testing.T) {
	all := []string{
		model.CodeOwnershipDescriptorMissing, model.CodeOwnershipAccountableMismatch, model.CodeOwnershipAgrees,
		model.CodeDeviationExpired, model.CodeDeviationCurrent,
		model.CodeExceptionNoExpiry, model.CodeExceptionExpired, model.CodeExceptionValid,
		model.CodePRDFrontmatterMissing, model.CodePRDContractPresent,
		model.CodeFrontmatterCoreField, model.CodeTagsDrift, model.CodeFrontmatterInSync, model.CodePointerGuidance,
		model.CodeNodeIdentity, model.CodeNodeAbsent, model.CodeNodeHandOwned, model.CodeNodeDrift, model.CodeNodeInSync,
		model.CodeFeatureIndexAbsent, model.CodeFeatureIndexDrift, model.CodeFeatureIndexUnresolved, model.CodeFeatureIndexInSync,
		model.CodeSkillShadowing, model.CodeSkillDanglingExtends, model.CodeSkillsClean,
		model.CodeSliceLockMissing, model.CodeRepoNotLocked, model.CodeSliceSetDrift,
		model.CodeSliceFileMissing, model.CodeSliceHandEdited, model.CodeFederationSlicesMatch,
	}
	seen := map[string]bool{}
	for _, c := range all {
		if c == "" {
			t.Error("empty code constant")
		}
		if seen[c] {
			t.Errorf("duplicate code %q", c)
		}
		seen[c] = true
	}
	if len(seen) != len(all) {
		t.Errorf("got %d distinct codes, want %d", len(seen), len(all))
	}
}

func TestSeverityString(t *testing.T) {
	cases := map[model.Severity]string{
		model.SevOK: "ok", model.SevWarn: "warn", model.SevFail: "fail", model.Severity(99): "unknown",
	}
	for sev, want := range cases {
		if got := sev.String(); got != want {
			t.Errorf("Severity(%d).String() = %q, want %q", sev, got, want)
		}
	}
}

func TestExitCodeContract(t *testing.T) {
	want := map[model.ExitCode]int{
		model.ExitOK: 0, model.ExitValidation: 1, model.ExitUsage: 2, model.ExitWorkspace: 3,
		model.ExitArtifact: 4, model.ExitPrecondition: 5, model.ExitExternalTool: 6,
		model.ExitInteractive: 7, model.ExitConflict: 8,
	}
	for code, n := range want {
		if int(code) != n {
			t.Errorf("exit code %v = %d, want %d", code, int(code), n)
		}
	}
	if got := model.CodeOf(nil); got != model.ExitOK {
		t.Errorf("CodeOf(nil) = %d, want 0", got)
	}
	// An error with no code resolves to 1, matching Python's uncaught-exception
	// exit rather than inventing a new status.
	if got := model.CodeOf(fmt.Errorf("plain")); got != model.ExitValidation {
		t.Errorf("CodeOf(plain) = %d, want 1", got)
	}
	err := model.Errorf(model.ExitWorkspace, "no such team %q", "ghost")
	if got := model.CodeOf(err); got != model.ExitWorkspace {
		t.Errorf("CodeOf(workspace err) = %d, want 3", got)
	}
	if err.Error() != `no such team "ghost"` {
		t.Errorf("Error() = %q", err.Error())
	}
	// Wrapped errors must keep their code, or the seam leaks status information.
	if got := model.CodeOf(fmt.Errorf("resolving governance: %w", err)); got != model.ExitWorkspace {
		t.Errorf("CodeOf(wrapped) = %d, want 3", got)
	}
}

func TestExitCodeForReport(t *testing.T) {
	pass := model.Report{Gates: []model.GateResult{{Ordinal: 1, Findings: []model.Finding{
		{Severity: model.SevOK}, {Severity: model.SevWarn},
	}}}}
	if pass.Problems() != 0 {
		t.Error("warns must not count as problems")
	}
	if got := pass.ExitCode(); got != model.ExitOK {
		t.Errorf("ExitCode() = %d, want 0", got)
	}
	bad := model.Report{Gates: []model.GateResult{{Ordinal: 1, Findings: []model.Finding{{Severity: model.SevFail}}}}}
	if got := bad.ExitCode(); got != model.ExitValidation {
		t.Errorf("ExitCode() = %d, want 1", got)
	}
}
