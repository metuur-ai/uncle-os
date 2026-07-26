package graph

// Inherited from examples/selftest.py — the Unit 0 characterization harness that
// R-9.3 deletes. Task 6.1 ports every assertion it carried to a named Go test so
// the oracle survives the file.
//
// Covered here: ST-002, ST-003 (blocks_equal, `selftest.py:39,41`), ST-004
// (preserve-unknown through a tag rewrite, `:47`), and ST-005..ST-010 — the six
// generated-block marker states (`:54,59,65,72,77,82`).
//
// This is an in-package test file because rewriteGeneratedBlock is unexported.
// TestRewriteGeneratedBlockIsFailSafe in graph_test.go reaches the same code
// through graph.Build, but only for the unbalanced case and only via a fixture;
// the six states below are asserted directly, one per marker shape, which is
// what selftest.py did.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/yamlio"
)

// --------------------------------------------------------- blocks_equal

// TestBlocksEqualIgnoresWhitespaceAndLineEndings is selftest.py:38-40 (ST-002).
// Gate 5 compares a committed generated block against a freshly rendered one; if
// this were byte equality, every CRLF checkout would report drift.
func TestBlocksEqualIgnoresWhitespaceAndLineEndings(t *testing.T) {
	if !BlocksEqual("x \ny\n\n", "x\r\ny") {
		t.Error("BlocksEqual is sensitive to trailing whitespace, CRLF, or trailing blank lines")
	}
	// Each of the three normalizations, alone, so a regression names itself.
	for _, tc := range []struct{ name, a, b string }{
		{"trailing space", "x \ny", "x\ny"},
		{"CRLF", "x\r\ny", "x\ny"},
		{"trailing blank lines", "x\ny\n\n\n", "x\ny"},
	} {
		if !BlocksEqual(tc.a, tc.b) {
			t.Errorf("%s: BlocksEqual(%q, %q) = false, want true", tc.name, tc.a, tc.b)
		}
	}
}

// TestBlocksEqualDetectsRealDifference is selftest.py:41-42 (ST-003) — the
// not-vacuously-true half. Without it, `func BlocksEqual(a, b string) bool
// { return true }` passes the test above.
func TestBlocksEqualDetectsRealDifference(t *testing.T) {
	if BlocksEqual("x\ny", "x\nz") {
		t.Error("BlocksEqual reports two different blocks as equal")
	}
	if BlocksEqual("a\nb", "a") {
		t.Error("BlocksEqual reports a truncated block as equal")
	}
}

// ------------------------------------------- preserve-unknown (R-1.5)

// TestRewriteFrontmatterTagsPreservesUnknownKeys is selftest.py:44-50 (ST-004).
//
// yamlio's TestPreserveUnknownKeysThroughATagRewrite proves the primitive
// (MapSet on a loaded document keeps the key). This proves the WIRING: the
// derive -> rewrite -> re-read path graph build actually runs, across the
// frontmatter delimiters, where a struct-shaped round trip would drop the key
// silently.
func TestRewriteFrontmatterTagsPreservesUnknownKeys(t *testing.T) {
	path := filepath.Join(t.TempDir(), "doc.md")
	const src = "---\ntype: prd\nid: x\nvendorField: keep-me\n---\n\nbody\n"
	if err := os.WriteFile(path, []byte(src), 0o666); err != nil {
		t.Fatal(err)
	}

	meta, _, err := ReadFrontmatter(path)
	if err != nil {
		t.Fatalf("ReadFrontmatter: %v", err)
	}
	tags, err := DeriveTags(meta, nil)
	if err != nil {
		t.Fatalf("DeriveTags: %v", err)
	}
	if len(tags) == 0 {
		t.Fatal("DeriveTags returned nothing; the rewrite below would be a no-op")
	}
	changed, err := RewriteFrontmatterTags(path, tags)
	if err != nil {
		t.Fatalf("RewriteFrontmatterTags: %v", err)
	}
	if !changed {
		t.Fatal("the tag rewrite reported no change; nothing was exercised")
	}

	after, body, err := ReadFrontmatter(path)
	if err != nil {
		t.Fatalf("ReadFrontmatter after rewrite: %v", err)
	}
	if got := after.Get("vendorField"); !yamlio.PyEqual(got, yamlio.PyStr("keep-me")) {
		t.Errorf("vendorField = %v, want \"keep-me\" — the rewrite dropped an unknown key", got)
	}
	if got := after.Get("type"); !yamlio.PyEqual(got, yamlio.PyStr("prd")) {
		t.Errorf("type = %v, want \"prd\"", got)
	}
	if string(body) != "\nbody\n" {
		t.Errorf("body = %q, want %q", string(body), "\nbody\n")
	}
}

// ---------------------------------------------- generated-block states

// rewriteBlock is selftest.py's rewrite() helper (`:24-31`): write initial (or
// leave the file absent), rewrite, return whether it changed plus the resulting
// text.
func rewriteBlock(t *testing.T, initial *string, block string) (bool, *markerImbalance, string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "CLAUDE.md")
	if initial != nil {
		if err := os.WriteFile(path, []byte(*initial), 0o666); err != nil {
			t.Fatal(err)
		}
	}
	changed, bad, err := rewriteGeneratedBlock(path, block)
	if err != nil {
		t.Fatalf("rewriteGeneratedBlock: %v", err)
	}
	raw, readErr := os.ReadFile(path)
	if os.IsNotExist(readErr) {
		return changed, bad, ""
	}
	if readErr != nil {
		t.Fatal(readErr)
	}
	return changed, bad, string(raw)
}

func ptr(s string) *string { return &s }

// TestRewriteGeneratedBlockCreatesWhenAbsent is selftest.py:53-55 (ST-005),
// R-3.4.
func TestRewriteGeneratedBlockCreatesWhenAbsent(t *testing.T) {
	changed, bad, txt := rewriteBlock(t, nil, "B1")
	if !changed {
		t.Error("creating an absent node did not report a change")
	}
	if bad != nil {
		t.Errorf("marker imbalance reported on a fresh file: %+v", bad)
	}
	if txt == "" {
		t.Fatal("no file was created")
	}
	for _, want := range []string{genStart, genEnd, "B1"} {
		if !strings.Contains(txt, want) {
			t.Errorf("created node is missing %q:\n%s", want, txt)
		}
	}
	// The header is what makes the file self-describing to a human who opens it
	// before ever running graph build.
	if !strings.Contains(txt, "Hand-owned context") {
		t.Errorf("created node carries no hand-owned header:\n%s", txt)
	}
}

// TestRewriteGeneratedBlockAppendsPreservingProse is selftest.py:57-60 (ST-006),
// R-3.5: zero markers plus hand-written prose appends rather than overwrites.
func TestRewriteGeneratedBlockAppendsPreservingProse(t *testing.T) {
	const prose = "# Hand-written\n\nkeep this prose\n"
	changed, bad, txt := rewriteBlock(t, ptr(prose), "B1")
	if !changed {
		t.Error("appending a region to a marker-less node did not report a change")
	}
	if bad != nil {
		t.Errorf("marker imbalance reported for a marker-less node: %+v", bad)
	}
	if !strings.Contains(txt, "keep this prose") {
		t.Errorf("hand-written prose was destroyed:\n%s", txt)
	}
	if !strings.Contains(txt, genStart) || !strings.Contains(txt, "B1") {
		t.Errorf("the generated region was not appended:\n%s", txt)
	}
	// Appended, not prepended: the prose keeps the top of the file.
	if !strings.HasPrefix(txt, prose) {
		t.Errorf("the region landed before the prose:\n%s", txt)
	}
}

// TestRewriteGeneratedBlockReplacesInteriorOnly is selftest.py:63-67 (ST-007),
// R-3.3. Four separate assertions, not one boolean: the selftest form could not
// say which half broke.
func TestRewriteGeneratedBlockReplacesInteriorOnly(t *testing.T) {
	before := "PRE\n" + RenderGeneratedRegion("OLD") + "\nPOST\n"
	changed, bad, txt := rewriteBlock(t, ptr(before), "NEW")
	if !changed {
		t.Error("replacing the interior did not report a change")
	}
	if bad != nil {
		t.Errorf("marker imbalance reported for one balanced pair: %+v", bad)
	}
	if !strings.Contains(txt, "NEW") {
		t.Errorf("the new block was not written:\n%s", txt)
	}
	if strings.Contains(txt, "OLD") {
		t.Errorf("the old block survived the rewrite:\n%s", txt)
	}
	if !strings.HasPrefix(txt, "PRE\n") {
		t.Errorf("text before the start marker was not byte-verbatim:\n%s", txt)
	}
	if !strings.HasSuffix(strings.TrimRight(txt, " \t\n\r"), "POST") {
		t.Errorf("text after the end marker was not byte-verbatim:\n%s", txt)
	}
}

// TestRewriteGeneratedBlockIdenticalIsNoop is selftest.py:70-72 (ST-008): a
// rebuild that produces the same block must not touch the file, or `graph build;
// graph build` stops being a no-op diff and CI churns on every run.
func TestRewriteGeneratedBlockIdenticalIsNoop(t *testing.T) {
	same := "PRE\n" + RenderGeneratedRegion("NEW") + "\nPOST\n"
	changed, bad, txt := rewriteBlock(t, ptr(same), "NEW")
	if changed {
		t.Error("rewriting an identical block reported a change")
	}
	if bad != nil {
		t.Errorf("marker imbalance reported for one balanced pair: %+v", bad)
	}
	if txt != same {
		t.Errorf("the file was rewritten\n got: %q\nwant: %q", txt, same)
	}
}

// TestRewriteGeneratedBlockUnbalancedMarkersNoMutation is selftest.py:75-77
// (ST-009), R-3.6: a start with no end is a node someone is mid-editing. Warn,
// change nothing.
func TestRewriteGeneratedBlockUnbalancedMarkersNoMutation(t *testing.T) {
	bad := "x\n" + genStart + "\nonly start\n"
	changed, imbalance, txt := rewriteBlock(t, ptr(bad), "NEW")
	if changed {
		t.Error("an unbalanced node reported a change")
	}
	if txt != bad {
		t.Errorf("an unbalanced node was mutated\n got: %q\nwant: %q", txt, bad)
	}
	if imbalance == nil {
		t.Fatal("no imbalance was reported for a start without an end")
	}
	if imbalance.starts != 1 || imbalance.ends != 0 {
		t.Errorf("counts = %d start / %d end, want 1 / 0", imbalance.starts, imbalance.ends)
	}
}

// TestRewriteGeneratedBlockDuplicateMarkersNoMutation is selftest.py:80-82
// (ST-010), R-3.6. Two balanced pairs is ambiguous — there is no "the" block —
// so the fail-safe answer is the same as for an imbalance.
func TestRewriteGeneratedBlockDuplicateMarkersNoMutation(t *testing.T) {
	dup := RenderGeneratedRegion("A") + "\n" + RenderGeneratedRegion("B") + "\n"
	changed, imbalance, txt := rewriteBlock(t, ptr(dup), "NEW")
	if changed {
		t.Error("a node with two marker pairs reported a change")
	}
	if txt != dup {
		t.Errorf("a node with two marker pairs was mutated\n got: %q\nwant: %q", txt, dup)
	}
	if imbalance == nil {
		t.Fatal("no imbalance was reported for two marker pairs")
	}
	if imbalance.starts != 2 || imbalance.ends != 2 {
		t.Errorf("counts = %d start / %d end, want 2 / 2", imbalance.starts, imbalance.ends)
	}
}
