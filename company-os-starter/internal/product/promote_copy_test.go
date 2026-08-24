package product

// `prd promote` — the copy-forward half (R-4.6, R-5.1 … R-5.17, R-6.2, R-6.5).
//
// Two things here are not ordinary unit assertions and are load-bearing:
//
//   - every refusal case is run against a whole-tree snapshot, because R-4.6 is
//     a statement about the FILESYSTEM and not about the error value. A refusal
//     that returns the right error after writing half a record still fails.
//   - the ordering test (R-5.9) observes the workspace from INSIDE the rebuild
//     hook, which is the only point from which the sequence is visible at all.
//     Asserting the end state cannot distinguish "rebuilt then digested" from
//     "digested then rebuilt"; the digest is the thing the two disagree about.

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/graph"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/model"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/workspace"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/yamlio"
)

// ------------------------------------------------------------- helpers

// snapshot is every file under root, path -> contents. Comparing two of these is
// the R-4.6 assertion: a refusal changed nothing, including files this test does
// not know exist.
func snapshot(t *testing.T, root string) map[string]string {
	t.Helper()
	out := map[string]string{}
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(root, path)
		out[filepath.ToSlash(rel)] = string(b)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func sameTree(t *testing.T, before, after map[string]string, why string) {
	t.Helper()
	for p, b := range before {
		a, ok := after[p]
		if !ok {
			t.Errorf("%s: %s was deleted", why, p)
		} else if a != b {
			t.Errorf("%s: %s was modified", why, p)
		}
	}
	for p := range after {
		if _, ok := before[p]; !ok {
			t.Errorf("%s: %s was created", why, p)
		}
	}
}

// readyDraft scaffolds a draft and fills exactly what R-5.3 demands, so the
// tests that are not about readiness start from a promotable draft. The body is
// the built-in template's, untouched, so its three sections are the real ones.
func readyDraft(t *testing.T, ws *workspace.Workspace, title string) (id, path string) {
	t.Helper()
	if _, err := DraftNew(ws, "core", "payments", title, "", nil); err != nil {
		t.Fatal(err)
	}
	id = draftID(slugify(title))
	dir, err := DraftDir(ws, "core")
	if err != nil {
		t.Fatal(err)
	}
	path = DraftPath(dir, id)
	fillDraft(t, path, nil)
	return id, path
}

// fillDraft supplies the six process fields, then applies edit for the cases
// that want one of them back in its unfilled state.
func fillDraft(t *testing.T, path string, edit func(yamlio.PyMap) yamlio.PyMap) {
	t.Helper()
	meta, body, err := graph.ReadFrontmatter(path)
	if err != nil {
		t.Fatal(err)
	}
	meta = meta.Set("title", yamlio.PyStr("Faster refunds")).
		Set("platform", yamlio.PyStr("payments")).
		Set("components", yamlio.PySeq{yamlio.PyStr("checkout-service")}).
		Set("governanceSnapshot", yamlio.PyStr("2026-01-01")).
		Set("decisionOwner", yamlio.PyStr("ada"))
	if edit != nil {
		meta = edit(meta)
	}
	if err := writeArtifact(path, meta, body); err != nil {
		t.Fatal(err)
	}
}

// slugify mirrors the id derivation for the titles these tests use.
func slugify(title string) string {
	return strings.ToLower(strings.ReplaceAll(title, " ", "-"))
}

func recordPath(ws *workspace.Workspace, platform, id string) string {
	return filepath.Join(ws.Root, "platforms", platform, "change-records", "active", id, "prd.md")
}

// findingByCode returns every finding carrying code, across all sections.
func findingsByCode(sections []model.GateResult, code string) []model.Finding {
	var out []model.Finding
	for _, s := range sections {
		for _, f := range s.Findings {
			if f.Code == code {
				out = append(out, f)
			}
		}
	}
	return out
}

// ------------------------------------------------------------- 5.1 preflight

// TestPromotePreflightRefusals is R-4.6 with R-5.6 and R-5.7: each refusal
// returns the stated exit code AND leaves the tree byte-identical.
func TestPromotePreflightRefusals(t *testing.T) {
	for _, tc := range []struct {
		name string
		// setup runs after the ready draft exists and may break it.
		setup    func(t *testing.T, ws *workspace.Workspace, id, path string)
		wantCode model.ExitCode
		wantIn   []string
	}{
		{
			// R-5.6: the status found is named, so the reader knows which of
			// promoted/abandoned they are looking at.
			name: "status is not draft",
			setup: func(t *testing.T, ws *workspace.Workspace, id, path string) {
				fillDraft(t, path, func(m yamlio.PyMap) yamlio.PyMap {
					return m.Set("status", yamlio.PyStr("promoted"))
				})
			},
			wantCode: model.ExitConflict,
			wantIn:   []string{"status is 'promoted'", "not 'draft'"},
		},
		{
			name: "abandoned draft names its own status",
			setup: func(t *testing.T, ws *workspace.Workspace, id, path string) {
				fillDraft(t, path, func(m yamlio.PyMap) yamlio.PyMap {
					return m.Set("status", yamlio.PyStr("abandoned"))
				})
			},
			wantCode: model.ExitConflict,
			wantIn:   []string{"status is 'abandoned'"},
		},
		{
			// R-4.2, reached through ResolvePromoteTarget: no target at all.
			name: "no target platform",
			setup: func(t *testing.T, ws *workspace.Workspace, id, path string) {
				fillDraft(t, path, func(m yamlio.PyMap) yamlio.PyMap {
					return m.Set("platform", yamlio.PyStr("TODO")).
						Set(KeyPromoteTo, yamlio.PyMap{{K: "platform", V: yamlio.PyStr("TODO")}})
				})
			},
			wantCode: model.ExitArtifact,
		},
		{
			// R-4.3.
			name: "target platform does not exist",
			setup: func(t *testing.T, ws *workspace.Workspace, id, path string) {
				fillDraft(t, path, func(m yamlio.PyMap) yamlio.PyMap {
					return m.Set("platform", yamlio.PyStr("ledger")).
						Set(KeyPromoteTo, yamlio.PyMap{{K: "platform", V: yamlio.PyStr("ledger")}})
				})
			},
			wantCode: model.ExitWorkspace,
		},
		{
			// R-4.4: two fields disagreeing is a refusal, never a silent pick.
			// The target itself resolves — R-4.3 is checked first, deliberately
			// — so this is the disagreement and nothing else.
			name: "the two platform fields disagree",
			setup: func(t *testing.T, ws *workspace.Workspace, id, path string) {
				fillDraft(t, path, func(m yamlio.PyMap) yamlio.PyMap {
					return m.Set("platform", yamlio.PyStr("ledger"))
				})
			},
			wantCode: model.ExitArtifact,
			wantIn:   []string{"payments", "ledger", "must agree"},
		},
		{
			// R-5.3: one unfilled field is enough, and it is exit 5 (a
			// precondition), not exit 4 (a malformed artifact) — R-5.5.
			name: "a process field is still TODO",
			setup: func(t *testing.T, ws *workspace.Workspace, id, path string) {
				fillDraft(t, path, func(m yamlio.PyMap) yamlio.PyMap {
					return m.Set("decisionOwner", yamlio.PyStr("TODO"))
				})
			},
			wantCode: model.ExitPrecondition,
		},
		{
			// R-5.3: empty is as unfilled as TODO.
			name: "a process field is empty",
			setup: func(t *testing.T, ws *workspace.Workspace, id, path string) {
				fillDraft(t, path, func(m yamlio.PyMap) yamlio.PyMap {
					return m.Set("components", yamlio.PySeq{})
				})
			},
			wantCode: model.ExitPrecondition,
		},
		{
			// R-5.4.
			name: "a required section is missing from the body",
			setup: func(t *testing.T, ws *workspace.Workspace, id, path string) {
				meta, body, err := graph.ReadFrontmatter(path)
				if err != nil {
					t.Fatal(err)
				}
				cut := strings.SplitN(string(body), "## Success metrics", 2)[0]
				if err := writeArtifact(path, meta, []byte(cut)); err != nil {
					t.Fatal(err)
				}
			},
			wantCode: model.ExitPrecondition,
		},
		{
			// R-5.7: an existing target id is a conflict, and the record that is
			// already there is not touched.
			name: "the target record already exists",
			setup: func(t *testing.T, ws *workspace.Workspace, id, path string) {
				write(t, recordPath(ws, "payments", id),
					"---\ntype: prd\nstatus: proposed\npromotedFrom: {team: core, draft: "+
						id+", digest: sha256:deadbeef}\n---\n\n# somebody else's record\n")
			},
			wantCode: model.ExitConflict,
			wantIn:   []string{"already exists"},
		},
		{
			name: "no draft at that id",
			setup: func(t *testing.T, ws *workspace.Workspace, id, path string) {
				if err := os.Remove(path); err != nil {
					t.Fatal(err)
				}
			},
			wantCode: model.ExitWorkspace,
			wantIn:   []string{"no draft at"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ws := fixture(t)
			id, path := readyDraft(t, ws, "Faster refunds")
			tc.setup(t, ws, id, path)

			before := snapshot(t, ws.Root)
			_, err := PRDPromote(ws, "core", id, graph.Rebuild)
			if err == nil {
				t.Fatal("promote succeeded, want a refusal")
			}
			if got := model.CodeOf(err); got != tc.wantCode {
				t.Errorf("exit = %v, want %v: %v", got, tc.wantCode, err)
			}
			for _, want := range tc.wantIn {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("error = %q, want it to contain %q", err.Error(), want)
				}
			}
			// R-4.6: no partial state, no rebuild, no log line.
			sameTree(t, before, snapshot(t, ws.Root), "a refused promote")
		})
	}
}

// TestPromotePreflightReportsEveryGap is R-5.3 + R-5.4 + R-5.5: one run names
// ALL six fields and ALL three sections, and phrases it as work to do.
func TestPromotePreflightReportsEveryGap(t *testing.T) {
	ws := fixture(t)
	if _, err := DraftNew(ws, "core", "payments", "Faster refunds", "", nil); err != nil {
		t.Fatal(err)
	}
	id := draftID("faster-refunds")
	dir, _ := DraftDir(ws, "core")
	path := DraftPath(dir, id)
	// A draft as scaffolded: every field a placeholder. Strip the body too, so
	// the field gaps and the section gaps have to coexist in one report.
	meta, _, err := graph.ReadFrontmatter(path)
	if err != nil {
		t.Fatal(err)
	}
	// promoteTo stays `payments` so the target resolves and readiness — not the
	// target — is what refuses.
	if err := writeArtifact(path, meta, []byte("\n# Faster refunds\n")); err != nil {
		t.Fatal(err)
	}

	before := snapshot(t, ws.Root)
	out, err := PRDPromote(ws, "core", id, graph.Rebuild)
	if err == nil {
		t.Fatal("promote succeeded, want a readiness refusal")
	}
	if got := model.CodeOf(err); got != model.ExitPrecondition {
		t.Errorf("exit = %v, want %v (R-5.5: not-ready is a precondition, "+
			"not a malformed artifact)", got, model.ExitPrecondition)
	}
	sameTree(t, before, snapshot(t, ws.Root), "a not-ready promote")

	// R-5.3: all six named, in one pass. `platform` is filled by --platform, so
	// five of the six are still placeholders here.
	var fields []string
	for _, f := range findingsByCode(out, model.CodePromoteFieldMissing) {
		fields = append(fields, f.Fields.Str("field"))
	}
	for _, want := range processFields {
		// `team` and `platform` are supplied by `prd new --draft` itself; the
		// other four are the placeholders the author has to replace.
		if want == "platform" || want == "team" {
			continue
		}
		if !contains(fields, want) {
			t.Errorf("field %q is unfilled but unreported (R-5.3): %v", want, fields)
		}
	}
	// R-5.4: each missing section named, individually.
	var sections []string
	for _, f := range findingsByCode(out, model.CodePromoteSectionMissing) {
		sections = append(sections, f.Fields.Str("section"))
	}
	for _, want := range PRDSections {
		if !contains(sections, want) {
			t.Errorf("section %q is missing but unreported (R-5.4): %v", want, sections)
		}
	}
	// R-5.5 is a tone rule: a not-yet-ready report, not a malformed-artifact
	// error. The words that would break it are the done-gate's vocabulary.
	joined := strings.ToLower(strings.Join(messages(out, model.SevFail), "\n") +
		"\n" + strings.Join(messages(out, model.SevOK), "\n") + "\n" + err.Error())
	for _, forbidden := range []string{"malformed", "invalid", "corrupt", "broken"} {
		if strings.Contains(joined, forbidden) {
			t.Errorf("readiness report says %q (R-5.5):\n%s", forbidden, joined)
		}
	}
	if !strings.Contains(joined, "not ready") && !strings.Contains(joined, "not yet ready") {
		t.Errorf("readiness report never says the draft is not ready (R-5.5):\n%s", joined)
	}
}

// ------------------------------------------------------------- 5.2 copy-forward

// TestPromoteCopyForward is R-5.1, R-5.2, R-5.8 and R-6.2: the record appears
// with the draft's body and a change record's lifecycle fields, and the draft
// stays where it is.
func TestPromoteCopyForward(t *testing.T) {
	ws := fixture(t)
	id, path := readyDraft(t, ws, "Faster refunds")
	draftMetaBefore, draftBodyBefore, _ := readDraft(t, path)

	out, err := PRDPromote(ws, "core", id, graph.Rebuild)
	if err != nil {
		t.Fatal(err)
	}

	record := recordPath(ws, "payments", id)
	meta, body, _ := readDraft(t, record)

	// R-5.1: same id, seeded from the draft's body.
	if body != draftBodyBefore {
		t.Errorf("record body differs from the draft's (R-5.1):\n%q\n%q", body, draftBodyBefore)
	}
	// R-5.2: the three lifecycle values move.
	if got := strOf(meta, "status"); got != "proposed" {
		t.Errorf("record status = %q, want proposed (R-5.2)", got)
	}
	if got := strOf(meta, "created"); got != today().Format(isoDate) {
		t.Errorf("record created = %q, want today (R-5.2)", got)
	}
	if meta.Get(KeyPromoteTo) != nil {
		t.Error("record carries the draft's promoteTo routing field (R-5.2)")
	}
	// R-5.2: every other authored field is copied, not re-derived.
	for _, f := range processFields {
		if !yamlio.PyEqual(meta.Get(f), draftMetaBefore.Get(f)) {
			t.Errorf("record %s = %v, want the draft's %v (R-5.2)",
				f, meta.Get(f), draftMetaBefore.Get(f))
		}
	}

	// R-5.8: the origin draft is still there, in place, with its body intact.
	afterMeta, afterBody, _ := readDraft(t, path)
	if afterBody != draftBodyBefore {
		t.Error("the draft's body was truncated or rewritten (R-5.8)")
	}
	if got := strOf(afterMeta, "status"); got != "promoted" {
		t.Errorf("draft status = %q, want promoted (R-6.2)", got)
	}
	// R-6.2: the draft points forward, the record points back.
	to, ok := ReadPromotedTo(afterMeta)
	if !ok || to.Platform != "payments" || to.Record != id {
		t.Errorf("draft promotedTo = %+v, want payments/%s (R-6.2)", to, id)
	}
	from, ok := ReadPromotedFrom(meta)
	if !ok || from.Team != "core" || from.Draft != id {
		t.Errorf("record promotedFrom = %+v, want core/%s (R-6.2)", from, id)
	}
	// R-6.3: the digest is of the draft as it now stands — which is exactly what
	// the integrity gate recomputes.
	want, err := DraftDigestFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if from.Digest != want {
		t.Errorf("recorded digest = %q, want %q (R-6.3)", from.Digest, want)
	}

	// R-5.12: the closing line is the last thing said, and it points at
	// `prd validate` on the platform.
	next := model.NextCommands(out)
	if len(next) != 1 || !strings.Contains(next[0], "prd validate --platform payments") {
		t.Errorf("next commands = %v, want prd validate (R-5.12)", next)
	}
	last := out[len(out)-1]
	if last.Findings[len(last.Findings)-1].Code != model.CodePRDPromoteNext {
		t.Errorf("last finding = %q, want the promote pointer (R-5.12)",
			last.Findings[len(last.Findings)-1].Code)
	}
}

// TestPromoteCopyOrdering is R-5.9 and R-5.10: write record, rewrite draft,
// rebuild, digest, record it — observed from inside the rebuild, which is the
// only vantage point that can tell the order apart.
func TestPromoteCopyOrdering(t *testing.T) {
	ws := fixture(t)
	id, path := readyDraft(t, ws, "Faster refunds")
	record := recordPath(ws, "payments", id)

	called := 0
	hook := Rebuild(func(w *workspace.Workspace) ([]model.GateResult, error) {
		called++
		// Effects 1 and 2 have happened.
		rmeta, _, err := graph.ReadFrontmatter(record)
		if err != nil {
			t.Errorf("record does not exist at rebuild time (R-5.9): %v", err)
			return nil, nil
		}
		if got := strOf(rmeta, "status"); got != "proposed" {
			t.Errorf("record status at rebuild time = %q (R-5.9)", got)
		}
		dmeta, _, err := graph.ReadFrontmatter(path)
		if err != nil {
			t.Fatal(err)
		}
		if got := strOf(dmeta, "status"); got != "promoted" {
			t.Errorf("draft was not rewritten before the rebuild (R-5.9): %q", got)
		}
		// Effect 5 has NOT happened: the digest is recorded after the rebuild,
		// because the rebuild changes the bytes it is taken over (R-6.3).
		if from, ok := ReadPromotedFrom(rmeta); ok && from.Digest != "" {
			t.Errorf("digest was recorded before the rebuild (R-5.9): %q", from.Digest)
		}
		return graph.Rebuild(w)
	})

	if _, err := PRDPromote(ws, "core", id, hook); err != nil {
		t.Fatal(err)
	}
	if called != 1 {
		t.Errorf("rebuild called %d times, want 1 (R-5.10)", called)
	}
	// And after the run the digest agrees with the file the gate will hash.
	meta, _, _ := readDraft(t, record)
	from, _ := ReadPromotedFrom(meta)
	want, err := DraftDigestFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if from.Digest != want {
		t.Errorf("digest = %q, want %q — recorded before the rebuild settled (R-5.9)",
			from.Digest, want)
	}
}

// TestPromoteCopyDigestSurvivesRetagging is the reason R-5.9 pins the order:
// the rebuild rewrites the draft's `tags:`, so a digest taken one step earlier
// would be stale by the time the command returned. NormalizeForDigest exists so
// this holds either way (R-6.4) — this asserts it does.
func TestPromoteCopyDigestSurvivesRetagging(t *testing.T) {
	ws := fixture(t)
	id, path := readyDraft(t, ws, "Faster refunds")
	beforeRebuild, err := DraftDigestFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := PRDPromote(ws, "core", id, graph.Rebuild); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), "tags:") {
		t.Skip("the rebuild derived no tags for this fixture; nothing to prove")
	}
	after, err := DraftDigestFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if after == beforeRebuild {
		return // the normalized form already ignored the tag rewrite
	}
	// If it did move, the recorded digest must be the LATER one.
	meta, _, _ := readDraft(t, recordPath(ws, "payments", id))
	from, _ := ReadPromotedFrom(meta)
	if from.Digest != after {
		t.Errorf("recorded digest = %q, want the post-rebuild %q (R-5.9/R-6.4)",
			from.Digest, after)
	}
}

// ------------------------------------------------------------- 5.3 resume

// TestPromoteResume is R-5.11: an interrupted promotion — a record on disk with
// no digest recorded — is completed by re-running, not refused as a conflict.
func TestPromoteResume(t *testing.T) {
	ws := fixture(t)
	id, path := readyDraft(t, ws, "Faster refunds")

	// The state effect 1 leaves behind when effect 5 never runs: the record
	// exists, seeded, but nothing points back at the draft.
	meta, body, err := graph.ReadFrontmatter(path)
	if err != nil {
		t.Fatal(err)
	}
	record := recordPath(ws, "payments", id)
	if err := os.MkdirAll(filepath.Dir(record), 0o777); err != nil {
		t.Fatal(err)
	}
	if err := writeArtifact(record, promotedRecordMeta(meta), body); err != nil {
		t.Fatal(err)
	}

	if _, err := PRDPromote(ws, "core", id, graph.Rebuild); err != nil {
		t.Fatalf("re-running an interrupted promotion refused (R-5.11): %v", err)
	}
	rmeta, _, _ := readDraft(t, record)
	from, ok := ReadPromotedFrom(rmeta)
	if !ok || from.Digest == "" {
		t.Fatalf("resume left the record without a digest (R-5.11): %+v", from)
	}
	want, err := DraftDigestFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if from.Digest != want {
		t.Errorf("resumed digest = %q, want %q (R-5.11)", from.Digest, want)
	}

	// And a COMPLETE promotion is still a conflict — resume is the exception,
	// not the rule (R-5.7).
	before := snapshot(t, ws.Root)
	if _, err := PRDPromote(ws, "core", id, graph.Rebuild); err == nil {
		t.Error("promoting a finished promotion succeeded, want a conflict (R-5.7)")
	}
	sameTree(t, before, snapshot(t, ws.Root), "a second promote")
}

// TestPromoteResumeIsNotAnOverwrite is the other half of R-5.11: the resume
// path is reached only for a record that carries no digest. A record that
// someone else wrote by hand — with a digest — is never overwritten.
func TestPromoteResumeIsNotAnOverwrite(t *testing.T) {
	ws := fixture(t)
	id, _ := readyDraft(t, ws, "Faster refunds")
	record := recordPath(ws, "payments", id)
	write(t, record, "---\ntype: prd\nstatus: proposed\n"+
		"promotedFrom: {team: other, draft: somewhere-else, digest: sha256:beef}\n"+
		"---\n\n# not yours\n")
	kept, err := os.ReadFile(record)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := PRDPromote(ws, "core", id, graph.Rebuild); err == nil {
		t.Fatal("promote overwrote a foreign record, want a conflict (R-5.7)")
	}
	now, err := os.ReadFile(record)
	if err != nil {
		t.Fatal(err)
	}
	if string(now) != string(kept) {
		t.Error("the existing record was modified by a refused promote (R-4.6)")
	}
}

// ------------------------------------------------------------- 6.2 the log

// TestPromotionLogAppends is R-6.5: one dated line per promotion, appended, in
// the team's log.md — which is created if the team has never had one.
func TestPromotionLogAppends(t *testing.T) {
	ws := fixture(t)
	tdir, err := ws.TeamDir("core")
	if err != nil {
		t.Fatal(err)
	}
	log := filepath.Join(tdir, "log.md")
	if _, err := os.Stat(log); err == nil {
		t.Fatal("the fixture already has a log.md; this test proves it is created")
	}

	first, _ := readyDraft(t, ws, "Faster refunds")
	if _, err := PRDPromote(ws, "core", first, graph.Rebuild); err != nil {
		t.Fatal(err)
	}
	one, err := os.ReadFile(log)
	if err != nil {
		t.Fatalf("no log.md after a promotion (R-6.5): %v", err)
	}
	line := strings.TrimRight(string(one), "\n")
	for _, want := range []string{today().Format(isoDate), first, "payments"} {
		if !strings.Contains(line, want) {
			t.Errorf("log line %q does not name %q (R-6.5)", line, want)
		}
	}
	if strings.Count(string(one), "\n") != 1 {
		t.Errorf("one promotion wrote %d lines, want 1 (R-6.5):\n%s",
			strings.Count(string(one), "\n"), one)
	}

	second, _ := readyDraft(t, ws, "Slower refunds")
	if _, err := PRDPromote(ws, "core", second, graph.Rebuild); err != nil {
		t.Fatal(err)
	}
	two, err := os.ReadFile(log)
	if err != nil {
		t.Fatal(err)
	}
	// Append-only: the first line is byte-identical and still first.
	if !strings.HasPrefix(string(two), string(one)) {
		t.Errorf("the second promotion rewrote the log (R-6.5):\nwas:\n%s\nnow:\n%s", one, two)
	}
	lines := strings.Split(strings.TrimRight(string(two), "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("two promotions wrote %d lines, want 2 (R-6.5)", len(lines))
	}
	if !strings.Contains(lines[1], second) {
		t.Errorf("second line %q does not name the second draft (R-6.5)", lines[1])
	}
}

// TestPromotionLogUntouchedByRefusals is R-4.6 over the log specifically: the
// append is effect 6 and never runs early.
func TestPromotionLogUntouchedByRefusals(t *testing.T) {
	ws := fixture(t)
	if _, err := DraftNew(ws, "core", "payments", "Faster refunds", "", nil); err != nil {
		t.Fatal(err)
	}
	id := draftID("faster-refunds")
	if _, err := PRDPromote(ws, "core", id, graph.Rebuild); err == nil {
		t.Fatal("an unfilled draft promoted")
	}
	tdir, _ := ws.TeamDir("core")
	if _, err := os.Stat(filepath.Join(tdir, "log.md")); err == nil {
		t.Error("a refused promote wrote a log line (R-4.6/R-6.5)")
	}
}

// ------------------------------------------------------------- 5.4 parity

// TestPromoteDownstreamParity is R-5.13 … R-5.17: a promoted record and a
// hand-scaffolded one are the same KIND of thing, so nothing downstream can
// tell them apart.
//
// Each half runs in its OWN workspace, because the two records have to occupy
// the same position — same platform, same component, same active directory —
// for the comparison to mean anything. The outputs are then diffed modulo the
// ids and the temp root, which are the only two things that legitimately
// differ. Comparing codes alone would miss a message that leaked
// `promotedFrom` into a sentence; comparing raw text would fail on the ids.
func TestPromoteDownstreamParity(t *testing.T) {
	// The hand-scaffolded record: `prd new` on the same inputs, then filled in
	// the way an author would. Comparing a filled record against an unfilled one
	// would only prove that placeholders fail validation.
	wsA := fixture(t)
	if _, err := PRDNew(wsA, "core", "payments", "checkout-service",
		"Hand written", ""); err != nil {
		t.Fatal(err)
	}
	idA := derivePRDID("Hand written")
	fillDraft(t, recordPath(wsA, "payments", idA), nil)

	// The promoted one.
	wsB := fixture(t)
	idB, _ := readyDraft(t, wsB, "Faster refunds")
	if _, err := PRDPromote(wsB, "core", idB, graph.Rebuild); err != nil {
		t.Fatal(err)
	}

	// "Equivalent" is about the ARTIFACT, not about who typed the prose: the
	// body is the author's either way, and `prd new` fills its governance
	// checklist from the platform while a draft carries the template's. Give
	// both records the same body, so the only thing left that could make them
	// judged differently is the frontmatter promotion writes — which is the
	// thing R-5.13 … R-5.17 are actually about.
	_, sharedBody, err := graph.ReadFrontmatter(recordPath(wsB, "payments", idB))
	if err != nil {
		t.Fatal(err)
	}
	metaA, _, err := graph.ReadFrontmatter(recordPath(wsA, "payments", idA))
	if err != nil {
		t.Fatal(err)
	}
	if err := writeArtifact(recordPath(wsA, "payments", idA), metaA, sharedBody); err != nil {
		t.Fatal(err)
	}
	if _, err := graph.Rebuild(wsA); err != nil {
		t.Fatal(err)
	}

	for _, cmd := range []struct {
		rule string
		run  func(ws *workspace.Workspace, id string) ([]model.GateResult, error)
	}{
		{"R-5.13 prd validate", func(ws *workspace.Workspace, id string) ([]model.GateResult, error) {
			return PRDValidate(ws, "payments", id)
		}},
		{"R-5.14 prd complete", func(ws *workspace.Workspace, id string) ([]model.GateResult, error) {
			return PRDComplete(ws, "payments", id, false, graph.Rebuild)
		}},
		{"R-5.15 check ready", func(ws *workspace.Workspace, id string) ([]model.GateResult, error) {
			return Check(ws, "core", "checkout-service", "ready")
		}},
		{"R-5.16 check done", func(ws *workspace.Workspace, id string) ([]model.GateResult, error) {
			return Check(ws, "core", "checkout-service", "done")
		}},
		{"R-5.17 contract gate", func(ws *workspace.Workspace, id string) ([]model.GateResult, error) {
			g, err := Gate(ws, 3)
			return []model.GateResult{g}, err
		}},
	} {
		t.Run(cmd.rule, func(t *testing.T) {
			a, errA := cmd.run(wsA, idA)
			b, errB := cmd.run(wsB, idB)
			if (errA == nil) != (errB == nil) || model.CodeOf(errA) != model.CodeOf(errB) {
				t.Errorf("outcomes differ (%s):\nscaffolded %v\npromoted   %v",
					cmd.rule, errA, errB)
			}
			x := normalize(a, wsA.Root, idA)
			y := normalize(b, wsB.Root, idB)
			if !equalStrings(x, y) {
				t.Errorf("output differs modulo ids (%s):\nscaffolded %v\npromoted   %v",
					cmd.rule, x, y)
			}
		})
	}
}

// normalize is one finding per line, with the two things that legitimately
// differ between the two workspaces — the record id and the temp root — masked.
func normalize(sections []model.GateResult, root, id string) []string {
	var out []string
	mask := func(s string) string {
		s = strings.ReplaceAll(s, root, "<ROOT>")
		s = strings.ReplaceAll(s, id, "<ID>")
		return s
	}
	for _, s := range sections {
		for _, f := range s.Findings {
			out = append(out, mask(fmt.Sprintf("%s/%d/%s/%s",
				f.Code, f.Severity, f.Path, f.Message)))
		}
	}
	return out
}

// codes is a section's findings as `code/severity` pairs — the judgement,
// without the ids that necessarily differ between two records.
func codes(sections []model.GateResult) []string {
	var out []string
	for _, s := range sections {
		for _, f := range s.Findings {
			out = append(out, f.Code+"/"+string(rune('0'+int(f.Severity))))
		}
	}
	return out
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
