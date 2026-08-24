package validate_test

// Draft-time validation posture (R-3.1 … R-3.5) and the green-workspace
// guarantee (R-3.4).
//
// Nothing here drove a code change. The posture these tests describe is a
// consequence of two decisions already in the tree, and the point of the file is
// to make either one impossible to undo silently:
//
//   - Gate 3 (`internal/product.Gate`) sweeps `platforms/*/change-records/active`
//     and nothing else, so the four-field process contract is scoped to records a
//     platform has accepted. A draft lives under `teams/<t>/product/change-
//     records/draft`, outside that sweep — R-3.3.
//   - `product.CoreFieldErrors` stops at type / identity / status for `type: prd`
//     and never reads the body, so gate 4 — the only gate that visits a draft at
//     all — asks a draft for the core contract and nothing more: R-3.1, R-3.2 and
//     R-3.5 together.
//
// The split exists because drafting is thinking. The process contract is what a
// platform requires before it will carry a change; the core contract is what the
// graph requires to address a document at all. A draft owes the second, not the
// first. A future refactor that "unifies" the two sweeps, or that grows
// CoreFieldErrors a `title`/`platform` requirement, would erase the drafting
// stage without touching anything named draft — which is exactly the failure
// these tests are here to name.
//
// The R-3.1/R-3.2/R-3.3/R-3.5 fixtures are deliberately hand-written rather than
// produced by DraftNew: the subject is what validate DEMANDS of a draft, and a
// draft the CLI itself wrote can only ever show what the CLI happens to supply.
// R-3.4 is the reverse question — whether the shipped tool leaves a workspace
// green — so it goes through DraftNew over a committed example workspace.

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/graph"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/model"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/product"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/validate"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/workspace"
)

// ------------------------------------------------------------------ fixtures

// draftPostureFixture is one team and one platform, the smallest workspace in
// which both a draft and an active change record can exist.
func draftPostureFixture(t *testing.T) *workspace.Workspace {
	t.Helper()
	root := t.TempDir()
	draftWrite(t, filepath.Join(root, "teams", "core", "team.yaml"),
		"schemaVersion: '1.0'\nid: team://core\n")
	draftWrite(t, filepath.Join(root, "platforms", "payments", "platform.yaml"),
		"schemaVersion: '1.0'\nid: platform://payments\n")
	return workspace.New(root)
}

func draftWrite(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o777); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o666); err != nil {
		t.Fatal(err)
	}
}

// writeMinimalDraft writes a draft carrying ONLY the core contract — `type`, an
// `id` identity, `status` — plus the tags that derivation produces for exactly
// those three keys, so gate 4's tag-drift check is not what the test measures.
//
// It writes no `title`, `platform`, `components`, `governanceSnapshot` or
// `decisionOwner`, and a body with no PRD section headings. If validate ever
// grows a demand for any of those, this fixture is what stops carrying it.
func writeMinimalDraft(t *testing.T, ws *workspace.Workspace, team, id string) string {
	t.Helper()
	dir, err := product.DraftDir(ws, team)
	if err != nil {
		t.Fatal(err)
	}
	path := product.DraftPath(dir, id)
	draftWrite(t, path, "---\ntype: prd\nid: "+id+"\nstatus: draft\n"+
		"tags: [kind/prd, status/draft, team/"+team+"]\n---\n\n"+
		"# Draft: "+id+"\n\nStill thinking.\n")
	return path
}

// draftFindings returns every finding in the run whose Path or Subject names
// rel — the run's whole opinion of one document, across all gates.
func draftFindings(sections []model.GateResult, rel string) []model.Finding {
	var out []model.Finding
	for _, s := range sections {
		for _, f := range s.Findings {
			if f.Path == rel || f.Subject == rel {
				out = append(out, f)
			}
		}
	}
	return out
}

func draftRun(t *testing.T, ws *workspace.Workspace) []model.GateResult {
	t.Helper()
	sections, err := validate.Run(ws)
	if err != nil {
		t.Fatalf("validate.Run: %v", err)
	}
	return sections
}

// ------------------------------------------------------------ R-3.1 and R-3.2

// TestDraftRequiresOnlyTheCoreContract is R-3.1 and R-3.2: a draft carrying
// nothing but type, identity and status — and a body with no PRD headings —
// draws no failure anywhere in the run.
//
// The assertion is over EVERY gate rather than gate 4 alone, because "which gate
// looks at drafts" is itself the thing under test. A future gate that starts
// demanding `governanceSnapshot` from a draft fails here without needing to be
// anticipated by name.
func TestDraftRequiresOnlyTheCoreContract(t *testing.T) {
	ws := draftPostureFixture(t)
	path := writeMinimalDraft(t, ws, "core", "2026-thinking-out-loud")
	rel, err := filepath.Rel(ws.Root, path)
	if err != nil {
		t.Fatal(err)
	}

	found := draftFindings(draftRun(t, ws), rel)
	// Non-vacuity: gate 4 must be VISITING the draft for its silence to mean
	// anything. A traversal that stopped including drafts would otherwise turn
	// this test green by looking away.
	if len(found) == 0 {
		t.Fatalf("no gate reported on %s at all, so this test proves nothing", rel)
	}
	for _, f := range found {
		if f.Severity == model.SevFail {
			t.Errorf("a core-contract-only draft drew a failure: [%s] %s", f.Code, f.Message)
		}
	}
}

// TestDraftBodyIsNeverCheckedForPRDHeadings is R-3.2 stated positively: no
// finding in the run names a PRD section heading for a draft.
//
// The heading sweep lives in `prd validate`, which addresses
// `platforms/<p>/change-records/active/<id>/prd.md` and cannot reach a draft at
// all; workspace `validate` never runs it. This test fixes that as a property of
// the OUTPUT rather than of the call graph, so a heading check reintroduced from
// any direction is caught.
func TestDraftBodyIsNeverCheckedForPRDHeadings(t *testing.T) {
	ws := draftPostureFixture(t)
	path := writeMinimalDraft(t, ws, "core", "2026-no-headings")
	rel, err := filepath.Rel(ws.Root, path)
	if err != nil {
		t.Fatal(err)
	}

	found := draftFindings(draftRun(t, ws), rel)
	if len(found) == 0 {
		t.Fatalf("no gate reported on %s at all, so this test proves nothing", rel)
	}
	for _, f := range found {
		for _, heading := range product.PRDSections {
			if strings.Contains(f.Message, heading) {
				t.Errorf("validate inspected a draft's body for %q: [%s] %s",
					heading, f.Code, f.Message)
			}
		}
	}
}

// ------------------------------------------------------------------ R-3.3

// TestDraftIsNotSweptByThePRDContractGate is R-3.3: gate 3 reports on the
// platform's active record and says nothing at all about the team's draft —
// neither a pass nor a failure, because the draft is not in its corpus.
//
// The absence of an [ok] line matters as much as the absence of a [FAIL] one. An
// ok for a draft would mean the gate had walked it and found the four process
// fields present, which is the opposite of the drafting posture.
func TestDraftIsNotSweptByThePRDContractGate(t *testing.T) {
	ws := draftPostureFixture(t)
	draftPath := writeMinimalDraft(t, ws, "core", "2026-private")
	draftRel, err := filepath.Rel(ws.Root, draftPath)
	if err != nil {
		t.Fatal(err)
	}
	// An active record carrying the same shortfall the draft has, so the gate is
	// known to be running and known to be strict at the same time.
	draftWrite(t, filepath.Join(ws.Root, "platforms", "payments",
		"change-records", "active", "2026-accepted", "prd.md"),
		"---\ntype: prd\nid: 2026-accepted\nstatus: proposed\n---\n\n# 2026-accepted\n")

	gate, err := product.Gate(ws, 3)
	if err != nil {
		t.Fatalf("product.Gate: %v", err)
	}

	sawRecord := false
	for _, f := range gate.Findings {
		if strings.Contains(f.Path, "change-records/draft") ||
			strings.Contains(filepath.ToSlash(f.Path), "change-records/draft") {
			t.Errorf("gate 3 swept a team draft: [%s] %s", f.Code, f.Message)
		}
		if f.Path == draftRel {
			t.Errorf("gate 3 reported on the draft: [%s] %s", f.Code, f.Message)
		}
		if strings.Contains(f.Subject, "2026-accepted") {
			sawRecord = true
		}
	}
	if !sawRecord {
		t.Fatalf("gate 3 did not report on the active record, so its silence about "+
			"the draft proves nothing; findings=%v", gate.Findings)
	}
}

// ------------------------------------------------------------------ R-3.5

// TestDraftOmittingStatusFailsCoreFields is R-3.5: relaxing the PROCESS contract
// for drafts does not relax the CORE one. `prd` is in LifecycleTypes, so a
// document of that type owes a `status:` whatever its residency — and without
// one there is no way to tell a draft from a record, which is the distinction
// every other rule in this unit rests on.
func TestDraftOmittingStatusFailsCoreFields(t *testing.T) {
	ws := draftPostureFixture(t)
	dir, err := product.DraftDir(ws, "core")
	if err != nil {
		t.Fatal(err)
	}
	path := product.DraftPath(dir, "2026-statusless")
	draftWrite(t, path, "---\ntype: prd\nid: 2026-statusless\n"+
		"tags: [kind/prd, team/core]\n---\n\n# Draft\n")
	rel, err := filepath.Rel(ws.Root, path)
	if err != nil {
		t.Fatal(err)
	}

	found := false
	for _, f := range draftFindings(draftRun(t, ws), rel) {
		if f.Severity == model.SevFail && strings.Contains(f.Message, "status") {
			found = true
		}
	}
	if !found {
		t.Fatalf("a draft omitting status passed core-field validation")
	}
}

// ------------------------------------------------------------------ R-3.4

// draftCopyFixture copies a committed example workspace into a temp tree so the
// test can add a draft to it without dirtying the checkout.
func draftCopyFixture(t *testing.T, name string) *workspace.Workspace {
	t.Helper()
	src, err := filepath.Abs(filepath.Join("..", "..", "..", "examples", name))
	if err != nil {
		t.Fatalf("resolving fixture: %v", err)
	}
	if _, err := os.Stat(src); err != nil {
		t.Skipf("fixture unavailable: %v", err)
	}
	dst := filepath.Join(t.TempDir(), "ws")
	err = filepath.WalkDir(src, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, p)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o777)
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return nil
		}
		data, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		// Slices ship 0444; the copy has to be writable or TempDir cleanup fails.
		return os.WriteFile(target, data, 0o666)
	})
	if err != nil {
		t.Fatalf("copying fixture: %v", err)
	}
	return workspace.New(dst)
}

// TestGreenWorkspaceStaysGreenWithDrafts is R-3.4: adding drafts to a workspace
// that validates clean leaves it validating clean.
//
// This is the guarantee the whole drafting stage rests on. If a half-formed
// draft could redden CI, teams would keep their thinking outside the workspace,
// and the promotion path would have nothing to promote FROM.
//
// The drafts are created through DraftNew rather than hand-written, because
// R-3.4 is conditioned on "derived artifacts are current" and DraftNew's own
// rebuild (R-2.8) is what makes that true without a separate `graph build`. Two
// drafts, not one, so a per-team collision or an id-uniqueness assumption would
// show.
func TestGreenWorkspaceStaysGreenWithDrafts(t *testing.T) {
	ws := draftCopyFixture(t, "workspace")

	before := draftRun(t, ws)
	if model.HasFailure(before) {
		t.Skipf("examples/workspace does not validate clean before the drafts; " +
			"R-3.4 is not measurable against it")
	}

	team := firstTeam(t, ws)
	for _, title := range []string{"Draft One", "Draft Two"} {
		if _, err := product.DraftNew(ws, team, "", title, "", graph.Rebuild); err != nil {
			t.Fatalf("prd new --draft %q: %v", title, err)
		}
	}

	after := draftRun(t, ws)
	if model.HasFailure(after) {
		var msgs []string
		for _, s := range after {
			for _, f := range s.Findings {
				if f.Severity == model.SevFail {
					msgs = append(msgs, "["+f.Code+"] "+f.Message)
				}
			}
		}
		t.Fatalf("adding well-formed drafts reddened a green workspace:\n%s",
			strings.Join(msgs, "\n"))
	}
}

// firstTeam names a team the fixture actually has, so the test does not encode
// the example workspace's roster.
func firstTeam(t *testing.T, ws *workspace.Workspace) string {
	t.Helper()
	teams := ws.AllTeams()
	if len(teams) == 0 {
		t.Fatalf("fixture has no teams")
	}
	return filepath.Base(teams[0])
}
