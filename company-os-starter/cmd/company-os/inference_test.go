package main

// ux-simplification 1.2: dispatch-layer tests for platform/team inference.
//
// These tests drive run() — the real dispatch path — so they prove the
// wiring from parse through resolvePlatform/resolveTeam to the product
// command, not just that workspace.FindPRD returns the right string. The
// workspace-level scan has its own tests in internal/workspace.
//
// The fixture trees are small: one or two platforms with active or archived
// PRD directories, and one or two teams with discovery directories. No PRD
// bodies are needed for the ambiguity and absence tests because the error
// fires before the file is read.

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/model"
)

// inferWorkspace builds a workspace with the given extra platform and team
// directories under a temp root. scratchWorkspace already creates "plat" and
// "core"; the slices name additional ids to scaffold.
func inferWorkspace(t *testing.T, platforms, teams []string) string {
	t.Helper()
	root := scratchWorkspace(t)
	for _, p := range platforms {
		if p != "plat" {
			mkAll(t, filepath.Join(root, "platforms", p))
		}
	}
	for _, tm := range teams {
		if tm != "core" {
			mkAll(t, filepath.Join(root, "teams", tm))
		}
	}
	return root
}

func mkAll(t *testing.T, p string) {
	t.Helper()
	if err := os.MkdirAll(p, 0o755); err != nil {
		t.Fatal(err)
	}
}

// ------------------------------------------------ unique inference (active)

// TestPRDValidateInfersPlatformFromActivePRD is the headline case: the user
// omits --platform, the PRD exists under exactly one platform's active
// records, and the command proceeds as if the flag had been passed.
func TestPRDValidateInfersPlatformFromActivePRD(t *testing.T) {
	root := inferWorkspace(t, []string{"comms"}, nil)
	mkAll(t, filepath.Join(root, "platforms", "comms",
		"change-records", "active", "2026-unique"))
	writeFile(t, filepath.Join(root, "platforms", "comms",
		"change-records", "active", "2026-unique", "prd.md"),
		"---\ntype: prd\nid: 2026-unique\ntitle: Unique\nstatus: proposed\n"+
			"team: core\nplatform: comms\ncomponents: []\n"+
			"governanceSnapshot: 2026-01-01\ndecisionOwner: someone\n"+
			"created: 2026-01-01\n---\n\n# PRD: Unique\n\n"+
			"## Problem signal\nP\n\n## Success metrics\nM\n\n"+
			"## Affected components\n\n## Applicable governance\n")

	var stdout, stderr bytes.Buffer
	code := run([]string{"--root", root, "prd", "validate", "2026-unique"},
		&stdout, &stderr)
	// Exit 0 means the PRD passed validation; exit 1 means it failed
	// validation with [FAIL] findings. Either is fine — the point is that
	// inference resolved the platform and the command ran, rather than
	// exiting 2 (usage) or 3 (not found).
	if code != 0 && code != 1 {
		t.Fatalf("run() = %d, want 0 or 1\nstdout: %s\nstderr: %s",
			code, stdout.String(), stderr.String())
	}
}

// --------------------------------------------- unique inference (archived)

// TestPRDValidateInfersPlatformFromArchivedPRD pins the archive path: the
// PRD is archived, inference finds it, and validate refuses at exit 3 because
// it only validates active records.
//
// The refusal says WHY. It used to report "no active PRD at <active-path>",
// which reads as "nothing found" for an id the tool just located, and names a
// path under active/ the user never typed — an artifact of inference resolving
// the platform from the archive. It now names the archive and the platform.
func TestPRDValidateInfersPlatformFromArchivedPRD(t *testing.T) {
	root := inferWorkspace(t, []string{"comms"}, nil)
	mkAll(t, filepath.Join(root, "platforms", "comms",
		"archive", "prds", "2026-archived"))

	var stdout, stderr bytes.Buffer
	code := run([]string{"--root", root, "prd", "validate", "2026-archived"},
		&stdout, &stderr)
	if code != int(model.ExitWorkspace) {
		t.Fatalf("run() = %d, want %d (not-found in active)\nstdout: %s\nstderr: %s",
			code, model.ExitWorkspace, stdout.String(), stderr.String())
	}
	errText := stderr.String()
	for _, want := range []string{"2026-archived", "archived", "comms"} {
		if !strings.Contains(errText, want) {
			t.Errorf("stderr = %q, missing %q", errText, want)
		}
	}
	// The misleading sentence must be gone from THIS path, not merely joined.
	if strings.Contains(errText, "no active PRD at") {
		t.Errorf("stderr = %q, still uses the not-found voice for an archived id", errText)
	}
}

// TestPRDValidateAbsentKeepsNotFoundVoice is the other half of the contract:
// when the id is nowhere — not active, not archived — the original sentence is
// unchanged, because there "not here" is the whole truth.
func TestPRDValidateAbsentKeepsNotFoundVoice(t *testing.T) {
	root := inferWorkspace(t, []string{"comms"}, nil)

	var stdout, stderr bytes.Buffer
	code := run([]string{"--root", root, "prd", "validate", "2026-ghost",
		"--platform", "comms"}, &stdout, &stderr)
	if code != int(model.ExitWorkspace) {
		t.Fatalf("run() = %d, want %d\nstdout: %s\nstderr: %s",
			code, model.ExitWorkspace, stdout.String(), stderr.String())
	}
	if !strings.Contains(stderr.String(), "no active PRD at") {
		t.Errorf("stderr = %q, want the unchanged not-found sentence", stderr.String())
	}
}

// --------------------------------------------------------- ambiguous (usage)

// TestPRDValidateAmbiguousListsPlatforms pins the multi-platform case: two
// platforms hold the same id, and the error names every candidate so the
// user can pick without another command.
func TestPRDValidateAmbiguousListsPlatforms(t *testing.T) {
	root := inferWorkspace(t, []string{"alpha", "beta"}, nil)
	mkAll(t, filepath.Join(root, "platforms", "alpha",
		"change-records", "active", "2026-dup"))
	mkAll(t, filepath.Join(root, "platforms", "beta",
		"change-records", "active", "2026-dup"))

	var stdout, stderr bytes.Buffer
	code := run([]string{"--root", root, "prd", "validate", "2026-dup"},
		&stdout, &stderr)
	if code != int(model.ExitUsage) {
		t.Fatalf("run() = %d, want %d\nstdout: %s\nstderr: %s",
			code, model.ExitUsage, stdout.String(), stderr.String())
	}
	errText := stderr.String()
	for _, want := range []string{"alpha", "beta", "--platform"} {
		if !strings.Contains(errText, want) {
			t.Errorf("stderr = %q, missing %q", errText, want)
		}
	}
}

// ------------------------------------------------------------ absent (error)

// TestPRDValidateAbsentID pins the zero-match case: the id is not under any
// platform, and the error says so in the existing not-found style.
func TestPRDValidateAbsentID(t *testing.T) {
	root := inferWorkspace(t, nil, nil)

	var stdout, stderr bytes.Buffer
	code := run([]string{"--root", root, "prd", "validate", "2026-ghost"},
		&stdout, &stderr)
	if code != int(model.ExitWorkspace) {
		t.Fatalf("run() = %d, want %d\nstdout: %s\nstderr: %s",
			code, model.ExitWorkspace, stdout.String(), stderr.String())
	}
	if !strings.Contains(stderr.String(), "2026-ghost") {
		t.Errorf("stderr = %q, should name the absent id", stderr.String())
	}
}

// ----------------------------------------- explicit flag overrides inference

// TestPRDValidateExplicitFlagOverridesInference pins the override contract:
// the user passes --platform A even though the id also exists under platform
// B. The flag wins; the command runs against A and reports A's result.
func TestPRDValidateExplicitFlagOverridesInference(t *testing.T) {
	root := inferWorkspace(t, []string{"alpha", "beta"}, nil)
	// alpha has the PRD as a real active record.
	mkAll(t, filepath.Join(root, "platforms", "alpha",
		"change-records", "active", "2026-override"))
	writeFile(t, filepath.Join(root, "platforms", "alpha",
		"change-records", "active", "2026-override", "prd.md"),
		"---\ntype: prd\nid: 2026-override\ntitle: Override\nstatus: proposed\n"+
			"team: core\nplatform: alpha\ncomponents: []\n"+
			"governanceSnapshot: 2026-01-01\ndecisionOwner: someone\n"+
			"created: 2026-01-01\n---\n\n# PRD: Override\n\n"+
			"## Problem signal\nP\n\n## Success metrics\nM\n\n"+
			"## Affected components\n\n## Applicable governance\n")
	// beta also has the id, but as an empty directory (so it would be
	// ambiguous if inference ran).
	mkAll(t, filepath.Join(root, "platforms", "beta",
		"change-records", "active", "2026-override"))

	var stdout, stderr bytes.Buffer
	code := run([]string{"--root", root, "prd", "validate",
		"--platform", "alpha", "2026-override"}, &stdout, &stderr)
	// The flag resolves to alpha; the PRD is valid -> exit 0 or 1.
	if code != 0 && code != 1 {
		t.Fatalf("run() = %d, want 0 or 1\nstdout: %s\nstderr: %s",
			code, stdout.String(), stderr.String())
	}
	// The error should NOT mention ambiguity — the flag bypassed inference.
	if strings.Contains(stderr.String(), "multiple platforms") {
		t.Errorf("stderr = %q, should not mention ambiguity when --platform is explicit",
			stderr.String())
	}
}

// ----------------------------------------- discover validate inference

// TestDiscoverValidateInfersTeam pins the team-side twin: the user omits
// --team, the brief exists under exactly one team, and validate proceeds.
func TestDiscoverValidateInfersTeam(t *testing.T) {
	root := inferWorkspace(t, nil, []string{"edge"})
	mkAll(t, filepath.Join(root, "teams", "edge",
		"product", "discovery", "2026-found"))
	writeFile(t, filepath.Join(root, "teams", "edge",
		"product", "discovery", "2026-found", "brief.md"),
		"---\ntype: discovery-brief\nid: 2026-found\ntitle: Found\n"+
			"status: draft\nteam: edge\ncreated: 2026-01-01\n"+
			"tags: [kind/discovery, team/edge, status/draft]\n---\n\n"+
			"# Discovery: Found\n\n## Problem signal\nP\n\n"+
			"## Success criteria\nS\n\n## Stakeholders\n\n## Risks\n")

	var stdout, stderr bytes.Buffer
	code := run([]string{"--root", root, "discover", "validate", "2026-found"},
		&stdout, &stderr)
	if code != 0 && code != 1 {
		t.Fatalf("run() = %d, want 0 or 1\nstdout: %s\nstderr: %s",
			code, stdout.String(), stderr.String())
	}
}

// TestDiscoverValidateAmbiguousListsTeams pins the multi-team case.
func TestDiscoverValidateAmbiguousListsTeams(t *testing.T) {
	root := inferWorkspace(t, nil, []string{"alpha", "beta"})
	mkAll(t, filepath.Join(root, "teams", "alpha",
		"product", "discovery", "2026-shared"))
	mkAll(t, filepath.Join(root, "teams", "beta",
		"product", "discovery", "2026-shared"))

	var stdout, stderr bytes.Buffer
	code := run([]string{"--root", root, "discover", "validate", "2026-shared"},
		&stdout, &stderr)
	if code != int(model.ExitUsage) {
		t.Fatalf("run() = %d, want %d\nstdout: %s\nstderr: %s",
			code, model.ExitUsage, stdout.String(), stderr.String())
	}
	errText := stderr.String()
	for _, want := range []string{"alpha", "beta", "--team"} {
		if !strings.Contains(errText, want) {
			t.Errorf("stderr = %q, missing %q", errText, want)
		}
	}
}

// TestDiscoverValidateAbsentID pins the zero-match case for discovery.
func TestDiscoverValidateAbsentID(t *testing.T) {
	root := inferWorkspace(t, nil, nil)

	var stdout, stderr bytes.Buffer
	code := run([]string{"--root", root, "discover", "validate", "2026-ghost"},
		&stdout, &stderr)
	if code != int(model.ExitWorkspace) {
		t.Fatalf("run() = %d, want %d\nstdout: %s\nstderr: %s",
			code, model.ExitWorkspace, stdout.String(), stderr.String())
	}
	if !strings.Contains(stderr.String(), "2026-ghost") {
		t.Errorf("stderr = %q, should name the absent id", stderr.String())
	}
}

// --------------------------------------- existing flag-carrying invocations

// TestExistingFlagCarryingInvocationsUnchanged pins byte-identical behavior
// for every existing flag-carrying invocation: passing --platform / --team
// explicitly reaches the same code path it always did.
func TestExistingFlagCarryingInvocationsUnchanged(t *testing.T) {
	root := inferWorkspace(t, nil, nil)
	// prd validate with --platform and a missing PRD -> exit 3, same as before.
	var stdout, stderr bytes.Buffer
	code := run([]string{"--root", root, "prd", "validate",
		"--platform", "plat", "2026-nothing"}, &stdout, &stderr)
	if code != int(model.ExitWorkspace) {
		t.Fatalf("prd validate --platform: run() = %d, want %d", code, model.ExitWorkspace)
	}
	if !strings.Contains(stderr.String(), "no active PRD") {
		t.Errorf("stderr = %q, want the existing not-found message", stderr.String())
	}

	// discover validate with --team and a missing brief -> exit 3, same as before.
	stdout.Reset()
	stderr.Reset()
	code = run([]string{"--root", root, "discover", "validate",
		"--team", "core", "2026-nothing"}, &stdout, &stderr)
	if code != int(model.ExitWorkspace) {
		t.Fatalf("discover validate --team: run() = %d, want %d", code, model.ExitWorkspace)
	}
	if !strings.Contains(stderr.String(), "no brief") {
		t.Errorf("stderr = %q, want the existing not-found message", stderr.String())
	}
}

// ============================================================================
// ux-simplification Phase 2 (Unit 2): `prd new` context inference.
//
// The Phase 1 tests above cover inference that LOCATES an existing artifact.
// These cover inference that decides where a NEW one belongs — a different
// claim, so they use a different resolver (InferPlatformForNew enumerates
// platforms/ rather than searching for a record by id) and get their own
// tests rather than extending the ones above.
// ============================================================================

// TestPRDNewInfersSolePlatform is the headline case (R-2.1): a workspace with
// exactly one platform needs no --platform, because there is only one answer.
func TestPRDNewInfersSolePlatform(t *testing.T) {
	root := inferWorkspace(t, nil, nil) // scratchWorkspace makes exactly "plat"

	var stdout, stderr bytes.Buffer
	code := run([]string{"--root", root, "prd", "new",
		"--team", "core", "--title", "Sole platform"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("run() = %d, want 0\nstdout: %s\nstderr: %s",
			code, stdout.String(), stderr.String())
	}
	created := filepath.Join(root, "platforms", "plat",
		"change-records", "active", "2026-sole-platform", "prd.md")
	if _, err := os.Stat(created); err != nil {
		t.Errorf("PRD not created under the inferred platform: %v", err)
	}
}

// TestPRDNewAmbiguousPlatformListsCandidates pins R-2.2: several platforms is
// a usage error naming every one of them, never a guess. `prd new` WRITES, so
// a wrong inference files the record in the wrong catalog.
func TestPRDNewAmbiguousPlatformListsCandidates(t *testing.T) {
	root := inferWorkspace(t, []string{"alpha", "beta"}, nil)

	var stdout, stderr bytes.Buffer
	code := run([]string{"--root", root, "prd", "new",
		"--team", "core", "--title", "Ambiguous"}, &stdout, &stderr)
	if code != int(model.ExitUsage) {
		t.Fatalf("run() = %d, want %d\nstderr: %s", code, model.ExitUsage, stderr.String())
	}
	for _, want := range []string{"alpha", "beta", "plat", "--platform"} {
		if !strings.Contains(stderr.String(), want) {
			t.Errorf("stderr = %q, missing %q", stderr.String(), want)
		}
	}
}

// TestPRDNewExplicitPlatformWins pins R-2.6 for the create path: the flag is
// used unread even where inference would have refused as ambiguous.
func TestPRDNewExplicitPlatformWins(t *testing.T) {
	root := inferWorkspace(t, []string{"alpha", "beta"}, nil)

	var stdout, stderr bytes.Buffer
	code := run([]string{"--root", root, "prd", "new", "--team", "core",
		"--platform", "beta", "--title", "Explicit"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("run() = %d, want 0\nstderr: %s", code, stderr.String())
	}
	created := filepath.Join(root, "platforms", "beta",
		"change-records", "active", "2026-explicit", "prd.md")
	if _, err := os.Stat(created); err != nil {
		t.Errorf("PRD not created under the explicit platform: %v", err)
	}
}

// TestPRDNewNoPlatformKeepsExistingVoice pins R-2.3: a workspace with no
// platforms at all is not an ambiguity the user can fix with a flag, so the
// existing PlatformDir diagnostic is left to speak rather than a usage error
// naming a remedy that does not exist.
func TestPRDNewNoPlatformKeepsExistingVoice(t *testing.T) {
	root := t.TempDir()
	mkAll(t, filepath.Join(root, "teams", "core"))

	var stdout, stderr bytes.Buffer
	code := run([]string{"--root", root, "prd", "new",
		"--team", "core", "--title", "Nowhere"}, &stdout, &stderr)
	if code == 0 {
		t.Fatalf("run() = 0, want a failure\nstdout: %s", stdout.String())
	}
	if code == int(model.ExitUsage) {
		t.Errorf("run() = %d (usage); a platform-less workspace is not a "+
			"usage error the user can fix by passing --platform", code)
	}
}

// ============================================================================
// ux-simplification Phase 2 (Unit 3): the guidance chain prints resolved values.
// ============================================================================

// writeDraftBrief creates a status:draft brief under the named team.
func writeDraftBrief(t *testing.T, root, team, id string) {
	t.Helper()
	dir := filepath.Join(root, "teams", team, "product", "discovery", id)
	mkAll(t, dir)
	writeFile(t, filepath.Join(dir, "brief.md"),
		"---\ntype: discovery-brief\nid: "+id+"\ntitle: Guidance\n"+
			"status: draft\nteam: "+team+"\ncreated: 2026-01-01\n"+
			"tags: [kind/discovery, team/"+team+", status/draft]\n---\n\n"+
			// All three DiscoverySections must be present or validation FAILS
			// before any guidance is emitted — which would make these tests
			// pass vacuously by finding no placeholder in an error message.
			"# Discovery: Guidance\n\n## Problem signal\nP\n\n"+
			"## Hypothesis\nH\n\n"+
			"## Success criteria\nS\n\n## Stakeholders\n\n## Risks\n")
}

// TestDiscoverValidateGuidanceResolvesPlaceholders is R-3.1/R-3.6: in a
// workspace with one platform the printed `prd new` carries real values, so the
// user can paste it instead of going to look two ids up.
func TestDiscoverValidateGuidanceResolvesPlaceholders(t *testing.T) {
	root := inferWorkspace(t, nil, nil) // exactly one platform: "plat"
	writeDraftBrief(t, root, "core", "2026-guidance")

	var stdout, stderr bytes.Buffer
	if code := run([]string{"--root", root, "discover", "validate", "2026-guidance"},
		&stdout, &stderr); code != 0 && code != 1 {
		t.Fatalf("run() = %d\nstdout: %s\nstderr: %s", code, stdout.String(), stderr.String())
	}
	out := stdout.String()
	if strings.Contains(out, "<platform-id>") {
		t.Errorf("guidance still carries the platform placeholder:\n%s", out)
	}
	if !strings.Contains(out, "--platform plat") {
		t.Errorf("guidance should name the sole platform:\n%s", out)
	}
}

// TestDiscoverValidateGuidanceKeepsPlaceholdersWhenAmbiguous is R-3.2: with
// several platforms there is nothing to substitute, and the placeholder is the
// honest answer. Ambiguity must NOT turn this read-mostly command into a usage
// error — the guidance is advice, not a gate.
func TestDiscoverValidateGuidanceKeepsPlaceholdersWhenAmbiguous(t *testing.T) {
	root := inferWorkspace(t, []string{"alpha", "beta"}, nil)
	writeDraftBrief(t, root, "core", "2026-ambiguous")

	var stdout, stderr bytes.Buffer
	code := run([]string{"--root", root, "discover", "validate", "2026-ambiguous"},
		&stdout, &stderr)
	if code == int(model.ExitUsage) {
		t.Fatalf("ambiguity turned discover validate into a usage error:\n%s",
			stderr.String())
	}
	if !strings.Contains(stdout.String(), "<platform-id>") {
		t.Errorf("guidance should keep the placeholder when ambiguous:\n%s",
			stdout.String())
	}
}
