package validate

// R-7.1 … R-7.8 and R-10.2, one test per branch of the gate plus the two
// placement facts.
//
// Every fixture here is built rather than checked in, because the interesting
// input is a DIGEST — a value that only exists once the draft's bytes exist.
// Hard-coding one would freeze the normalized form into the test suite and make
// R-6.4 (one implementation) unfalsifiable from here.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/federation"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/model"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/product"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/skills"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/workspace"
)

// ------------------------------------------------------------------ fixtures

// promotionFixture is one team and one platform — the smallest workspace in
// which a draft can name a record.
func promotionFixture(t *testing.T) *workspace.Workspace {
	t.Helper()
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "teams", "core", "team.yaml"),
		"schemaVersion: '1.0'\nid: team://core\n")
	writeFile(t, filepath.Join(root, "platforms", "payments", "platform.yaml"),
		"schemaVersion: '1.0'\nid: platform://payments\n")
	return workspace.New(root)
}

func writeFile(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o777); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o666); err != nil {
		t.Fatal(err)
	}
}

// writeDraft writes one draft and returns its path. The extra frontmatter is
// appended verbatim so each test can state only the keys it is about.
func writeDraft(t *testing.T, ws *workspace.Workspace, team, id, status, extra string) string {
	t.Helper()
	dir, err := product.DraftDir(ws, team)
	if err != nil {
		t.Fatal(err)
	}
	path := product.DraftPath(dir, id)
	writeFile(t, path, "---\ntype: prd\nid: "+id+"\nstatus: "+status+"\n"+extra+"---\n\n# "+id+"\n")
	return path
}

// writeRecord writes a change record at rel (a platform-relative directory) and
// returns its path.
func writeRecord(t *testing.T, ws *workspace.Workspace, platform, rel, id, extra string) string {
	t.Helper()
	path := filepath.Join(ws.Root, "platforms", platform, rel, id, "prd.md")
	writeFile(t, path, "---\ntype: prd\nid: "+id+"\nstatus: proposed\n"+extra+"---\n\n# "+id+"\n")
	return path
}

// promotedToBlock is the draft-side provenance block.
func promotedToBlock(platform, record string) string {
	return "promotedTo:\n  platform: " + platform + "\n  record: " + record + "\n"
}

// promotedFromBlock is the record-side provenance block.
func promotedFromBlock(team, draft, digest string) string {
	return "promotedFrom:\n  team: " + team + "\n  draft: " + draft +
		"\n  digest: " + digest + "\n"
}

// federatedFixture is examples/federated, the committed workspace that carries a
// workspace.yaml. It is the only fixture that can prove the federation gate is
// still last, because gate placement is only observable when the conditional
// gate exists. This file is an internal test (it calls promotionGate directly),
// so it cannot reuse golden_test.go's copy of this path.
func federatedFixture(t *testing.T) string {
	t.Helper()
	dir, err := filepath.Abs(filepath.Join("..", "..", "..", "examples", "federated"))
	if err != nil {
		t.Fatalf("resolving examples/federated: %v", err)
	}
	if _, err := os.Stat(dir); err != nil {
		t.Skipf("fixture unavailable: %v", err)
	}
	return dir
}

// runPromotionGate runs the gate alone, at an arbitrary ordinal.
func runPromotionGate(t *testing.T, ws *workspace.Workspace) model.GateResult {
	t.Helper()
	g, err := promotionGate(ws, 8)
	if err != nil {
		t.Fatalf("promotionGate: %v", err)
	}
	if g.Slug != PromotionGateSlug || g.Title != PromotionGateTitle {
		t.Fatalf("gate identity = %q/%q", g.Slug, g.Title)
	}
	return g
}

// onlyFinding asserts the gate produced exactly one record and returns it.
func onlyFinding(t *testing.T, g model.GateResult) model.Finding {
	t.Helper()
	if len(g.Findings) != 1 {
		t.Fatalf("got %d findings, want 1: %+v", len(g.Findings), g.Findings)
	}
	return g.Findings[0]
}

// ------------------------------------------------------------------ branches

// TestPromotionGateEmptyWorkspace is R-7.8: no promoted drafts means the header
// and nothing under it, which is a GateResult with zero findings.
func TestPromotionGateEmptyWorkspace(t *testing.T) {
	ws := promotionFixture(t)
	if g := runPromotionGate(t, ws); len(g.Findings) != 0 {
		t.Errorf("got %d findings on a draft-less workspace, want 0: %+v",
			len(g.Findings), g.Findings)
	}
}

// TestPromotionGateIgnoresOpenAndAbandonedDrafts is R-7.7. Both drafts here name
// a record that does not exist, so any statement at all about them would be a
// failure — which is what makes the silence testable.
func TestPromotionGateIgnoresOpenAndAbandonedDrafts(t *testing.T) {
	ws := promotionFixture(t)
	writeDraft(t, ws, "core", "2026-open", "draft", promotedToBlock("payments", "ghost"))
	writeDraft(t, ws, "core", "2026-dead", "abandoned", promotedToBlock("payments", "ghost"))
	if g := runPromotionGate(t, ws); len(g.Findings) != 0 {
		t.Errorf("gate spoke about non-promoted drafts: %+v", g.Findings)
	}
}

// TestPromotionGateMatchingDigestPasses is R-7.2: recompute, compare, agree.
func TestPromotionGateMatchingDigestPasses(t *testing.T) {
	ws := promotionFixture(t)
	path := writeDraft(t, ws, "core", "2026-thing", "promoted",
		promotedToBlock("payments", "2026-thing"))
	digest, err := product.DraftDigestFile(path)
	if err != nil {
		t.Fatal(err)
	}
	writeRecord(t, ws, "payments", filepath.Join("change-records", "active"), "2026-thing",
		promotedFromBlock("core", "2026-thing", digest))

	f := onlyFinding(t, runPromotionGate(t, ws))
	if f.Severity != model.SevOK || f.Code != CodePromotionInSync {
		t.Fatalf("finding = %v/%s, want ok/%s", f.Severity, f.Code, CodePromotionInSync)
	}
	if f.Subject != "core/2026-thing" {
		t.Errorf("subject = %q, want core/2026-thing", f.Subject)
	}
	if f.Fields.Str("location") != "active" {
		t.Errorf("location = %q, want active", f.Fields.Str("location"))
	}
}

// TestPromotionGateArchivedRecordPasses is R-7.6: `prd complete` moved the
// record, and a match under archive/prds/ passes exactly as the active one does.
func TestPromotionGateArchivedRecordPasses(t *testing.T) {
	ws := promotionFixture(t)
	path := writeDraft(t, ws, "core", "2026-thing", "promoted",
		promotedToBlock("payments", "2026-thing"))
	digest, err := product.DraftDigestFile(path)
	if err != nil {
		t.Fatal(err)
	}
	writeRecord(t, ws, "payments", filepath.Join("archive", "prds"), "2026-thing",
		promotedFromBlock("core", "2026-thing", digest))

	f := onlyFinding(t, runPromotionGate(t, ws))
	if f.Severity != model.SevOK || f.Code != CodePromotionInSync {
		t.Fatalf("finding = %v/%s, want ok/%s", f.Severity, f.Code, CodePromotionInSync)
	}
	if f.Fields.Str("location") != "archived" {
		t.Errorf("location = %q, want archived", f.Fields.Str("location"))
	}
}

// TestPromotionGateDigestDriftFails is R-7.3: the draft was edited after
// promotion, and the failure names both documents.
func TestPromotionGateDigestDriftFails(t *testing.T) {
	ws := promotionFixture(t)
	path := writeDraft(t, ws, "core", "2026-thing", "promoted",
		promotedToBlock("payments", "2026-thing"))
	digest, err := product.DraftDigestFile(path)
	if err != nil {
		t.Fatal(err)
	}
	writeRecord(t, ws, "payments", filepath.Join("change-records", "active"), "2026-thing",
		promotedFromBlock("core", "2026-thing", digest))

	// The hand-edit the gate exists to catch.
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, path, string(body)+"\nA paragraph nobody carried forward.\n")

	f := onlyFinding(t, runPromotionGate(t, ws))
	if f.Severity != model.SevFail || f.Code != CodePromotionDigestDrift {
		t.Fatalf("finding = %v/%s, want FAIL/%s", f.Severity, f.Code, CodePromotionDigestDrift)
	}
	for _, want := range []string{"2026-thing", "payments/2026-thing"} {
		if !strings.Contains(f.Subject+" "+f.Message, want) {
			t.Errorf("%q does not name %q", f.Message, want)
		}
	}
	if f.Fields.Str("recorded") == f.Fields.Str("digest") {
		t.Error("drift reported with equal digests")
	}
}

// TestPromotionGateUnresolvedTargetFails is R-7.4: neither location holds it.
func TestPromotionGateUnresolvedTargetFails(t *testing.T) {
	ws := promotionFixture(t)
	writeDraft(t, ws, "core", "2026-thing", "promoted",
		promotedToBlock("payments", "2026-ghost"))

	f := onlyFinding(t, runPromotionGate(t, ws))
	if f.Severity != model.SevFail || f.Code != CodePromotionTargetUnresolved {
		t.Fatalf("finding = %v/%s, want FAIL/%s", f.Severity, f.Code,
			CodePromotionTargetUnresolved)
	}
	if !strings.Contains(f.Message, "payments/2026-ghost") {
		t.Errorf("%q does not name the unresolved target", f.Message)
	}
}

// TestPromotionGateInterruptedPromotionIsNotDrift is R-7.5. The record is there
// and the draft is intact; what is missing is the provenance write, so the
// reader is sent back to the command rather than hunting for an edit.
func TestPromotionGateInterruptedPromotionIsNotDrift(t *testing.T) {
	ws := promotionFixture(t)
	writeDraft(t, ws, "core", "2026-thing", "promoted",
		promotedToBlock("payments", "2026-thing"))
	writeRecord(t, ws, "payments", filepath.Join("change-records", "active"), "2026-thing", "")

	f := onlyFinding(t, runPromotionGate(t, ws))
	if f.Code != CodePromotionInterrupted {
		t.Fatalf("code = %s, want %s", f.Code, CodePromotionInterrupted)
	}
	if strings.Contains(f.Message, "drift") {
		t.Errorf("interrupted promotion reported as drift: %q", f.Message)
	}
	if !strings.Contains(f.Message, "prd promote --team core 2026-thing") {
		t.Errorf("%q does not name the command to re-run", f.Message)
	}
}

// TestPromotionGateScopesEachTeamToItsOwnDrafts: two teams, one gate, and the
// subjects stay attributable.
func TestPromotionGateScopesEachTeamToItsOwnDrafts(t *testing.T) {
	ws := promotionFixture(t)
	writeFile(t, filepath.Join(ws.Root, "teams", "risk", "team.yaml"),
		"schemaVersion: '1.0'\nid: team://risk\n")
	writeDraft(t, ws, "core", "2026-a", "promoted", promotedToBlock("payments", "ghost"))
	writeDraft(t, ws, "risk", "2026-b", "promoted", promotedToBlock("payments", "ghost"))

	g := runPromotionGate(t, ws)
	if len(g.Findings) != 2 {
		t.Fatalf("got %d findings, want 2", len(g.Findings))
	}
	got := []string{g.Findings[0].Subject, g.Findings[1].Subject}
	want := []string{"core/2026-a", "risk/2026-b"}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("subject[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

// ----------------------------------------------------------------- placement

// TestGateOrderPlacesPromotionAfterSkills is R-7.1 in monorepo mode: the gate is
// last of eight, the seven before it keep their slugs and their ordinals, and
// the denominator carried on the banner agrees with the list.
func TestGateOrderPlacesPromotionAfterSkills(t *testing.T) {
	ws := promotionFixture(t)
	sections, err := Run(ws)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	gates := sections[1:]
	if len(gates) != 8 {
		t.Fatalf("got %d gates, want 8", len(gates))
	}
	if gates[6].Slug != skills.GateSlug {
		t.Errorf("gate 7 = %q, want %q", gates[6].Slug, skills.GateSlug)
	}
	if gates[7].Slug != PromotionGateSlug || gates[7].Ordinal != 8 {
		t.Errorf("gate 8 = %q/%d, want %q/8", gates[7].Slug, gates[7].Ordinal,
			PromotionGateSlug)
	}
	for i, g := range gates {
		if g.Ordinal != i+1 {
			t.Errorf("gate %d carries ordinal %d", i+1, g.Ordinal)
		}
	}
	if total := sections[0].Findings[0].Fields.Int("gates"); total != 8 {
		t.Errorf("banner denominator = %d, want 8", total)
	}
}

// TestGateOrderKeepsFederationLast is R-10.2: with a manifest present the
// promotion gate is 8 and the federation gate is 9, still last.
func TestGateOrderKeepsFederationLast(t *testing.T) {
	root := federatedFixture(t)
	sections, err := Run(workspace.New(root))
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	gates := sections[1:]
	if len(gates) != 9 {
		t.Fatalf("got %d gates, want 9", len(gates))
	}
	if gates[7].Slug != PromotionGateSlug || gates[7].Ordinal != 8 {
		t.Errorf("gate 8 = %q/%d, want %q/8", gates[7].Slug, gates[7].Ordinal,
			PromotionGateSlug)
	}
	last := gates[len(gates)-1]
	if last.Slug != federation.GateSlug || last.Ordinal != 9 {
		t.Errorf("last gate = %q/%d, want %q/9", last.Slug, last.Ordinal,
			federation.GateSlug)
	}
	if total := sections[0].Findings[0].Fields.Int("gates"); total != 9 {
		t.Errorf("banner denominator = %d, want 9", total)
	}
}
