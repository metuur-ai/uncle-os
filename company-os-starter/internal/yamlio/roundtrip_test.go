package yamlio

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// The corpus is every YAML document git tracks under examples/: each .yaml/.yml
// file whole, and the frontmatter block of each .md file that has one. It is
// read from the repository rather than from a hand-picked fixture set so the
// fidelity claim covers what the CLI will actually be pointed at.
//
// Three properties are asserted over all of it:
//
//	[1] every document loads, and the tree survives a re-emit STRUCTURALLY
//	    unchanged — same kinds, tags, values, anchors and comments. This is
//	    R-1.7, and it holds for 100% of the corpus. Scalar STYLE is checked
//	    separately, because it is the one thing that does drift: see
//	    testdata/roundtrip-restyled.txt.
//	[2] emitting is idempotent: emit(load(emit(load(x)))) == emit(load(x)). One
//	    pass reaches a fixed point, which is what lets `graph build; graph build`
//	    be a no-op diff (R-0.6) even where pass one rewrote bytes.
//	[3] byte-identity against the authored file is MEASURED, not assumed. It
//	    does not hold everywhere, and the exact set that fails is frozen in
//	    testdata/roundtrip-divergent.txt so a new kind of infidelity fails the
//	    test rather than being absorbed.
//
// Property [3] fails on files PyYAML's emitter wrote in a shape yaml.v3's
// emitter cannot produce. Two causes, both emitter policy and neither
// reachable through yaml.v3's public API — see classifyDivergence.

var fmFence = regexp.MustCompile(`(?s)\A---\n(.*?)\n---\n(.*)\z`)

type corpusDoc struct {
	name string // repo-relative path, plus " (frontmatter)" for a block
	src  []byte
}

// loadCorpus enumerates the git-tracked YAML under examples/.
func loadCorpus(t *testing.T) []corpusDoc {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatalf("resolve repo root: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "examples")); err != nil {
		t.Skipf("no examples/ corpus at %s: %v", root, err)
	}
	out, err := exec.Command("git", "-C", root, "ls-files", "-z", "examples").Output()
	if err != nil {
		t.Skipf("git ls-files unavailable: %v", err)
	}

	var docs []corpusDoc
	for _, rel := range strings.Split(strings.TrimRight(string(out), "\x00"), "\x00") {
		data, err := os.ReadFile(filepath.Join(root, rel))
		if err != nil {
			t.Fatalf("read %s: %v", rel, err)
		}
		switch {
		case strings.HasSuffix(rel, ".yaml"), strings.HasSuffix(rel, ".yml"):
			docs = append(docs, corpusDoc{rel, data})
		case strings.HasSuffix(rel, ".md"):
			m := fmFence.FindSubmatch(data)
			if m == nil {
				continue
			}
			// The block comes back without its trailing newline; a YAML
			// document has one, and the emitter will produce one.
			docs = append(docs, corpusDoc{rel + " (frontmatter)", append(m[1], '\n')})
		}
	}
	// A corpus that silently shrank to nothing must not read as a pass.
	if len(docs) < 100 {
		t.Fatalf("corpus is implausibly small (%d documents); enumeration is broken", len(docs))
	}
	sort.Slice(docs, func(i, j int) bool { return docs[i].name < docs[j].name })
	return docs
}

func TestRoundTripCommittedCorpus(t *testing.T) {
	docs := loadCorpus(t)

	var identical, divergent, restyled []string
	causes := map[string][]string{}

	for _, d := range docs {
		doc, err := Load(d.src)
		if err != nil {
			t.Errorf("%s: Load failed on a committed artifact: %v", d.name, err)
			continue
		}
		got, err := doc.Bytes()
		if err != nil {
			t.Errorf("%s: Bytes failed: %v", d.name, err)
			continue
		}

		// [1] the tree survives the re-emit.
		back, err := Load(got)
		if err != nil {
			t.Errorf("%s: re-emitted bytes do not parse: %v", d.name, err)
			continue
		}
		structural, style := nodeDiff(doc.Root(), back.Root(), "$")
		if structural != "" {
			t.Errorf("%s: node tree changed across a round-trip: %s", d.name, structural)
		}
		if style != "" {
			restyled = append(restyled, d.name+"\t"+style)
		}

		// [2] one pass reaches a fixed point.
		again, err := back.Bytes()
		if err != nil {
			t.Errorf("%s: second Bytes failed: %v", d.name, err)
			continue
		}
		if !bytes.Equal(got, again) {
			t.Errorf("%s: emit is not idempotent; a second pass changed bytes", d.name)
		}

		// [3] byte-identity against the authored file, measured.
		if bytes.Equal(got, d.src) {
			identical = append(identical, d.name)
		} else {
			c := classifyDivergence(d.src, got)
			divergent = append(divergent, d.name+"\t"+c)
			causes[c] = append(causes[c], d.name)
		}
	}

	t.Logf("round-trip over %d committed documents: %d byte-identical, %d divergent, %d restyled",
		len(docs), len(identical), len(divergent), len(restyled))
	for _, c := range sortedKeys(causes) {
		t.Logf("  %3d  %s  (e.g. %s)", len(causes[c]), c, causes[c][0])
	}

	compareGolden(t, "testdata/roundtrip-divergent.txt", divergent)
	compareGolden(t, "testdata/roundtrip-restyled.txt", restyled)
}

// classifyDivergence names why an authored file did not come back byte-identical.
//
// Both known causes are PyYAML emitter policies that yaml.v3 neither implements
// nor exposes a knob for, so both are layout-only — the characters that carry
// meaning are identical and only whitespace moved:
//
//   - INDENTLESS BLOCK SEQUENCE. safe_dump writes a block sequence at its
//     parent's indent ("key:\n- item"); yaml.v3 always indents it
//     ("key:\n  - item"). yaml_emitter_increase_indent has no indentless mode
//     reachable from the Encoder API.
//   - LINE FOLDING. safe_dump wraps at best_width=80, breaking long flow
//     collections and quoted scalars across continuation lines; yaml.v3 emits
//     them on one line. Folding a quoted scalar replaces the break and its
//     following indent with a single space on re-read, which is why dropping
//     ALL whitespace is the right equivalence test here.
//
// Anything that survives neither reduction is reported as uncharacterized and
// is a genuine defect rather than a layout difference.
// The third cause is REQUOTING: yaml.v3's emitter single-quotes a plain scalar
// inside a FLOW mapping when it contains "://", where PyYAML leaves it plain.
// Dropping quote characters is a safe reduction here only because nodeDiff has
// already proved, independently, that every value, tag and comment is identical.
func classifyDivergence(in, out []byte) string {
	si, so := squeeze(in), squeeze(out)
	if si != so {
		if unquote(si) != unquote(so) {
			return "UNCHARACTERIZED — differs beyond whitespace and quoting"
		}
		return "flow-context requoting (yaml.v3 quotes a plain `://` scalar inside `{...}`)"
	}
	if dedent(in) == dedent(out) {
		return "block-sequence indent (safe_dump writes `key:\\n- item`)"
	}
	return "80-column line folding (usually together with the indent difference)"
}

func unquote(s string) string {
	return strings.NewReplacer("'", "", `"`, "").Replace(s)
}

// squeeze drops every space, tab and newline, leaving only the characters that
// carry meaning across both emitter layouts.
func squeeze(b []byte) string {
	return strings.Map(func(r rune) rune {
		if r == ' ' || r == '\t' || r == '\n' {
			return -1
		}
		return r
	}, string(b))
}

// dedent strips leading indentation from every line, which normalises away the
// block-sequence indent but leaves a folded document with extra lines.
func dedent(b []byte) string {
	lines := strings.Split(string(b), "\n")
	for i, l := range lines {
		lines[i] = strings.TrimLeft(l, " \t")
	}
	return strings.Join(lines, "\n")
}

// compareGolden asserts got matches the recorded list, and rewrites it when
// YAMLIO_RECORD=1 so the measurement can be refreshed deliberately.
func compareGolden(t *testing.T, path string, got []string) {
	t.Helper()
	body := strings.Join(got, "\n") + "\n"
	if os.Getenv("YAMLIO_RECORD") == "1" {
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatalf("record %s: %v", path, err)
		}
		t.Logf("recorded %d entries into %s", len(got), path)
		return
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s (re-record with YAMLIO_RECORD=1): %v", path, err)
	}
	if string(want) == body {
		return
	}
	wantSet := map[string]bool{}
	for _, l := range strings.Split(strings.TrimRight(string(want), "\n"), "\n") {
		wantSet[l] = true
	}
	gotSet := map[string]bool{}
	for _, l := range got {
		gotSet[l] = true
	}
	for _, l := range got {
		if !wantSet[l] {
			t.Errorf("NEW round-trip divergence, not previously measured: %s", l)
		}
	}
	for l := range wantSet {
		if !gotSet[l] {
			t.Errorf("%s now round-trips; drop it from %s", l, path)
		}
	}
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// nodeDiff compares two node trees and reports the first structural difference
// and the first style difference separately.
//
// Structural covers everything a read-modify-write must never alter: kinds,
// tags, values, anchors, comments and shape. Style — the quoting and flow/block
// choice — is reported apart from it because it is the one property that does
// drift across a re-emit, and the drift needs to be visible and frozen rather
// than lumped in with a value corruption.
func nodeDiff(a, b *yaml.Node, path string) (structural, style string) {
	switch {
	case a == nil && b == nil:
		return "", ""
	case a == nil || b == nil:
		return fmt.Sprintf("%s: one side is nil (%v vs %v)", path, a != nil, b != nil), ""
	}
	for _, f := range []struct {
		name     string
		got, exp any
	}{
		{"kind", a.Kind, b.Kind},
		{"tag", a.Tag, b.Tag},
		{"value", a.Value, b.Value},
		{"anchor", a.Anchor, b.Anchor},
		{"headComment", a.HeadComment, b.HeadComment},
		{"lineComment", a.LineComment, b.LineComment},
		{"footComment", a.FootComment, b.FootComment},
		{"children", len(a.Content), len(b.Content)},
	} {
		if f.got != f.exp {
			return fmt.Sprintf("%s: %s %#v -> %#v", path, f.name, f.got, f.exp), ""
		}
	}
	if a.Style != b.Style {
		style = fmt.Sprintf("%s (%q): style %s -> %s",
			path, a.Value, styleName(a.Style), styleName(b.Style))
	}
	for i := range a.Content {
		child := fmt.Sprintf("%s[%d]", path, i)
		if a.Kind == yaml.MappingNode && i%2 == 0 {
			child = fmt.Sprintf("%s.%s", path, a.Content[i].Value)
		}
		s, st := nodeDiff(a.Content[i], b.Content[i], child)
		if s != "" {
			return s, style
		}
		if style == "" {
			style = st
		}
	}
	return "", style
}

func styleName(s yaml.Style) string {
	names := []struct {
		bit  yaml.Style
		name string
	}{
		{yaml.TaggedStyle, "tagged"},
		{yaml.DoubleQuotedStyle, "double-quoted"},
		{yaml.SingleQuotedStyle, "single-quoted"},
		{yaml.LiteralStyle, "literal"},
		{yaml.FoldedStyle, "folded"},
		{yaml.FlowStyle, "flow"},
	}
	var out []string
	for _, n := range names {
		if s&n.bit != 0 {
			out = append(out, n.name)
		}
	}
	if len(out) == 0 {
		return "plain"
	}
	return strings.Join(out, "+")
}
