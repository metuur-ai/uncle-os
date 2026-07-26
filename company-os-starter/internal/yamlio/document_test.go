package yamlio

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func mustLoad(t *testing.T, src string) *Document {
	t.Helper()
	d, err := Load([]byte(src))
	if err != nil {
		t.Fatalf("Load(%q): %v", src, err)
	}
	return d
}

func mustEmit(t *testing.T, d *Document) string {
	t.Helper()
	b, err := d.Bytes()
	if err != nil {
		t.Fatalf("Bytes: %v", err)
	}
	return string(b)
}

// ----------------------------------------------------------------- R-1.7 core

// authored is the block task 1.3 names in its acceptance line: unknown keys, a
// key order no sort would produce, and every quote style mixed together. It also
// carries comments, which R-0.7a(b) requires to survive.
//
// What it deliberately does NOT contain is the two layout constructs the
// emitter provably cannot reproduce — a multi-space gap before a line comment
// and a multi-line folded scalar. Both are pinned by
// TestMeasuredLayoutLossesOnReEmit instead of being hidden here.
const authored = `# governance deviations, hand-authored
zeta: last-alphabetically-but-authored-first
vendorField: keep-me
schemaVersion: '1.0'
team: customer-engagement # who owns this
"quoted key": plain value
double: "double quoted"
single: 'single quoted'
literal: |
  a literal block
  second line
folded: >-
  a folded block
flowSeq: [a, b, c]
flowMap: {x: 1, y: 2}
deviations:
  - rule: 'req://a'
    tier: default
tags: []
`

func TestAuthoredBlockRoundTripsByteIdentical(t *testing.T) {
	got := mustEmit(t, mustLoad(t, authored))
	if got != authored {
		t.Errorf("round-trip is not byte-identical:\n--- want ---\n%s\n--- got ---\n%s", authored, got)
	}
}

// TestMeasuredLayoutLossesOnReEmit pins the two layout constructs yaml.v3's
// emitter normalises away. Both are layout-only: the scalar VALUE and the
// comment TEXT are preserved, only their formatting is not. They are asserted
// so that the port's fidelity claim is exact, and so a future emitter change
// that fixes either one is noticed rather than absorbed.
func TestMeasuredLayoutLossesOnReEmit(t *testing.T) {
	t.Run("multi-space gap before a line comment collapses to one", func(t *testing.T) {
		const src = "team: customer-engagement    # who owns this\n"
		const want = "team: customer-engagement # who owns this\n"
		if got := mustEmit(t, mustLoad(t, src)); got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("a folded scalar is re-folded onto one line", func(t *testing.T) {
		// The break and its indent fold to a single space on read, so the value
		// is unchanged; only the authored wrapping is lost. This is the same
		// mechanism as the 80-column divergence on 41 committed documents.
		const src = "folded: >-\n  a folded\n  block\n"
		const want = "folded: >-\n  a folded block\n"
		d := mustLoad(t, src)
		if v := MapGet(d.Root(), "folded"); v == nil || v.Value != "a folded block" {
			t.Fatalf("value changed, not just layout: %#v", v)
		}
		if got := mustEmit(t, d); got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})
}

func TestPreserveUnknownKeysThroughATagRewrite(t *testing.T) {
	// examples/selftest.py:44-50 — rewrite tags, then assert the key the CLI
	// has never heard of is still there. A struct unmarshal drops it silently.
	const src = "type: prd\nid: x\nvendorField: keep-me\n"

	d := mustLoad(t, src)
	tags := NewSequence()
	tags.Style = 0
	if err := SeqAppend(tags, NewString("kind/prd"), NewString("component/x")); err != nil {
		t.Fatal(err)
	}
	if err := MapSet(d.Root(), "tags", tags); err != nil {
		t.Fatal(err)
	}

	got := mustEmit(t, d)
	if !strings.Contains(got, "vendorField: keep-me") {
		t.Errorf("unknown key did not survive the rewrite:\n%s", got)
	}
	// The three authored keys keep their authored order, and the new one lands
	// last rather than being sorted in.
	if want := []string{"type", "id", "vendorField", "tags"}; !equalStrings(MapKeys(d.Root()), want) {
		t.Errorf("key order = %v, want %v", MapKeys(d.Root()), want)
	}
	if !strings.HasPrefix(got, src) {
		t.Errorf("the authored bytes were rewritten, not appended to:\n%s", got)
	}
}

func TestCommentsSurviveAReadModifyWrite(t *testing.T) {
	// R-0.7a(b): PyYAML's safe_load destroys comments, so `deviation declare`
	// strips them today. Go keeps them. This is a SANCTIONED improvement, so it
	// is asserted rather than tolerated — if a future change starts dropping
	// comments, that is a regression against the carve-out, not a return to
	// parity.
	const src = `# file header comment
schemaVersion: '1.0'   # trailing comment
team: customer-engagement

# a comment introducing the list
deviations:
  - rule: 'req://a'   # why this one
    tier: default
`
	d := mustLoad(t, src)
	root := d.Root()
	if err := MapSet(root, "team", NewString("payments")); err != nil {
		t.Fatal(err)
	}
	got := mustEmit(t, d)

	for _, c := range []string{
		"# file header comment",
		"# trailing comment",
		"# a comment introducing the list",
		"# why this one",
	} {
		if !strings.Contains(got, c) {
			t.Errorf("comment %q did not survive:\n%s", c, got)
		}
	}
	if !strings.Contains(got, "team: payments") {
		t.Errorf("the edit did not land:\n%s", got)
	}
}

// ------------------------------------------------- modify fidelity, on the corpus

// TestModifyOneValueLeavesEveryOtherByteAlone drives the read-modify-write
// guarantee off the real corpus rather than a hand-picked sample: for every
// committed document that already round-trips byte-identically, it rewrites one
// top-level string value and asserts that exactly one line moved.
func TestModifyOneValueLeavesEveryOtherByteAlone(t *testing.T) {
	const sentinel = "MODIFIED-BY-TEST"
	checked := 0

	for _, doc := range loadCorpus(t) {
		d, err := Load(doc.src)
		if err != nil {
			t.Errorf("%s: %v", doc.name, err)
			continue
		}
		before, err := d.Bytes()
		if err != nil || !bytes.Equal(before, doc.src) {
			continue // only the byte-identical subset can prove "every other byte"
		}
		root := d.Root()
		key := lastPlainStringKey(root)
		if key == "" {
			continue
		}

		keysBefore := MapKeys(root)
		target := MapGet(root, key)
		replacement := NewString(sentinel)
		replacement.Style = target.Style
		if err := MapSet(root, key, replacement); err != nil {
			t.Fatal(err)
		}
		after := mustEmit(t, d)
		checked++

		if !equalStrings(MapKeys(root), keysBefore) {
			t.Errorf("%s: setting %q reordered keys: %v -> %v",
				doc.name, key, keysBefore, MapKeys(root))
		}
		changed := changedLines(string(doc.src), after)
		switch {
		case len(changed) != 1:
			t.Errorf("%s: setting %q changed %d lines, want 1: %v",
				doc.name, key, len(changed), changed)
		case !strings.Contains(changed[0], sentinel):
			t.Errorf("%s: the one changed line is not the edit: %q", doc.name, changed[0])
		}
	}

	if checked < 20 {
		t.Fatalf("only %d documents were exercised; the sample is too thin to prove anything", checked)
	}
	t.Logf("modify fidelity verified on %d byte-identical committed documents", checked)
}

// TestAppendToASequenceLeavesEveryOtherByteAlone is the `deviation declare`
// shape: load a list, push one entry, and assert nothing above it moved.
func TestAppendToASequenceLeavesEveryOtherByteAlone(t *testing.T) {
	const src = `schemaVersion: '1.0'
team: customer-engagement
deviations:
  - rule: 'req://a'
    tier: default
tags: [kind/deviations]
`
	d := mustLoad(t, src)
	entry := NewMapping()
	if err := MapSet(entry, "rule", NewString("req://b")); err != nil {
		t.Fatal(err)
	}
	if err := MapSet(entry, "tier", NewString("default")); err != nil {
		t.Fatal(err)
	}
	if err := SeqAppend(MapGet(d.Root(), "deviations"), entry); err != nil {
		t.Fatal(err)
	}

	got := mustEmit(t, d)
	want := `schemaVersion: '1.0'
team: customer-engagement
deviations:
  - rule: 'req://a'
    tier: default
  - rule: req://b
    tier: default
tags: [kind/deviations]
`
	if got != want {
		t.Errorf("append rewrote surrounding bytes:\n--- want ---\n%s\n--- got ---\n%s", want, got)
	}
}

// --------------------------------------------------------- the frontmatter seam

// TestFrontmatterFalsyCollapse pins `yaml.safe_load(m.group(1)) or {}`.
//
// Every row was measured against the repo's vendored PyYAML 6.0.2: the `or` is
// Python's truthiness operator, so 0, False, ” and [] collapse to an empty
// mapping exactly as None does. A None-only check would be wrong on six of
// these rows.
func TestFrontmatterFalsyCollapse(t *testing.T) {
	collapses := []struct{ name, src string }{
		{"empty", ""},
		{"comment only", "# just a comment"},
		{"whitespace only", "   "},
		{"document start marker", "---"},
		{"explicit null", "null"},
		{"tilde", "~"},
		{"empty single-quoted string", "''"},
		{"empty double-quoted string", `""`},
		{"zero", "0"},
		{"zero float", "0.0"},
		{"octal zero", "0b0"},
		{"false", "false"},
		{"no", "no"},
		{"off", "off"},
		{"empty flow sequence", "[]"},
		{"empty flow mapping", "{}"},
	}
	for _, tc := range collapses {
		d, err := LoadFrontmatter([]byte(tc.src))
		if err != nil {
			t.Errorf("%s: %v", tc.name, err)
			continue
		}
		root := d.Root()
		if root == nil || root.Kind != yaml.MappingNode || len(root.Content) != 0 {
			t.Errorf("%s: %q did not collapse to an empty mapping (kind %v, %d children)",
				tc.name, tc.src, kindOf(root), len(root.Content))
		}
	}

	// Truthy values are handed back as authored, including the non-mapping ones
	// Python also hands back — a caller doing meta.get() on those is Python's
	// AttributeError, not this package's problem.
	survives := []struct{ name, src, wantKey string }{
		{"mapping", "type: prd\n", "type"},
		{"quoted zero is a non-empty string", "'0'", ""},
		{"sequence", "- a\n- b\n", ""},
	}
	for _, tc := range survives {
		d, err := LoadFrontmatter([]byte(tc.src))
		if err != nil {
			t.Errorf("%s: %v", tc.name, err)
			continue
		}
		if tc.wantKey != "" && MapGet(d.Root(), tc.wantKey) == nil {
			t.Errorf("%s: key %q lost", tc.name, tc.wantKey)
		}
		if tc.wantKey == "" && d.IsFalsy() {
			t.Errorf("%s: %q was treated as falsy", tc.name, tc.src)
		}
	}
}

// TestSyntaxErrorPropagation covers what Python raises out of frontmatter()
// with no try around it. R-0.7a(e) turns the traceback into exit 4, so what
// this package owes is a typed error on exactly the inputs PyYAML refuses.
func TestSyntaxErrorPropagation(t *testing.T) {
	rejected := []struct{ name, src, why string }{
		{"dashes_in_yaml_string", `title: "a`,
			"PyYAML: ScannerError, unexpected end of stream in a quoted scalar"},
		{"unclosed flow mapping", "a: {b: 1",
			"PyYAML: ParserError"},
		{"multiple documents", "a: 1\n---\nb: 2\n",
			"PyYAML: ComposerError, expected a single document in the stream"},
		{"unconstructible date", "d: 2035-02-30\n",
			"PyYAML: ValueError, day is out of range for month"},
	}
	for _, tc := range rejected {
		_, err := LoadFrontmatter([]byte(tc.src))
		if err == nil {
			t.Errorf("%s: accepted %q, but %s", tc.name, tc.src, tc.why)
			continue
		}
		var se *SyntaxError
		if !errors.As(err, &se) {
			t.Errorf("%s: error is %T, want *SyntaxError so dispatch can map it to exit 4", tc.name, err)
		}
	}

	// MEASURED GAP, deliberately not closed. PyYAML's scanner raises
	// ScannerError("found character '\t' that cannot start any token") on this;
	// yaml.v3's scanner accepts it and yields {type: prd}. Closing it means
	// reimplementing PyYAML's scanner. The assertion states the divergence so it
	// is a known quantity rather than a surprise in the differential harness.
	d, err := LoadFrontmatter([]byte("type:\tprd"))
	if err != nil {
		t.Fatalf("expected the measured divergence, got an error: %v", err)
	}
	if v := MapGet(d.Root(), "type"); v == nil || v.Value != "prd" {
		t.Fatalf("tab-after-colon parsed unexpectedly: %#v", d.Root())
	}
}

// --------------------------------------------------------------- node helpers

func TestMapGetTakesTheLastDuplicate(t *testing.T) {
	// safe_load("a: 1\na: 2") is {'a': 2}; the node tree keeps both pairs, so
	// reading forwards would disagree with Python.
	d := mustLoad(t, "a: 1\nb: x\na: 2\n")
	if v := MapGet(d.Root(), "a"); v == nil || v.Value != "2" {
		t.Errorf("MapGet returned %v, want the last duplicate (2)", v)
	}
	if want := []string{"a", "b"}; !equalStrings(MapKeys(d.Root()), want) {
		t.Errorf("MapKeys = %v, want %v (a keeps its first position)", MapKeys(d.Root()), want)
	}
}

func TestMapSetRejectsTheWrongKind(t *testing.T) {
	d := mustLoad(t, "- a\n")
	if err := MapSet(d.Root(), "k", NewString("v")); err == nil {
		t.Error("MapSet accepted a sequence node")
	}
	if err := SeqAppend(mustLoad(t, "a: 1\n").Root(), NewString("v")); err == nil {
		t.Error("SeqAppend accepted a mapping node")
	}
}

func TestEmptyStreamEmitsNothing(t *testing.T) {
	d := mustLoad(t, "# only a comment\n")
	if d.Root() != nil {
		t.Errorf("empty stream has a root: %#v", d.Root())
	}
	if got := mustEmit(t, d); got != "" {
		t.Errorf("empty stream emitted %q", got)
	}
}

// ------------------------------------------------------------------- helpers

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

// lastPlainStringKey picks the last top-level key whose value is a single-line
// string scalar. It has to be a string: replacing a timestamp or an int with
// text would change the emitted quoting, and the test would then be measuring
// the emitter's quoting policy instead of its positional fidelity.
func lastPlainStringKey(root *yaml.Node) string {
	if root == nil || root.Kind != yaml.MappingNode {
		return ""
	}
	for i := len(root.Content) - 2; i >= 0; i -= 2 {
		k, v := root.Content[i], root.Content[i+1]
		if v.Kind != yaml.ScalarNode || strings.ContainsAny(v.Value, "\n") {
			continue
		}
		if v.Style&(yaml.LiteralStyle|yaml.FoldedStyle) != 0 {
			continue
		}
		if s, err := Resolve(v); err != nil || s.Kind != KindStr {
			continue
		}
		return k.Value
	}
	return ""
}

func changedLines(before, after string) []string {
	b, a := strings.Split(before, "\n"), strings.Split(after, "\n")
	if len(b) != len(a) {
		return []string{"line count changed"}
	}
	var out []string
	for i := range b {
		if b[i] != a[i] {
			out = append(out, a[i])
		}
	}
	return out
}
