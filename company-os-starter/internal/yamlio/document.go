package yamlio

// Node-level document access. There is no struct-unmarshal path anywhere in
// this codebase, and there must never be one — three separate invariants depend
// on the node tree surviving a read-modify-write:
//
//  1. Unknown frontmatter keys must survive a tag rewrite (examples/selftest.py
//     :44-50, R-1.7). Struct unmarshal drops every field the struct does not
//     name, silently.
//  2. Authored key order must survive. Python dicts are insertion-ordered and
//     safe_dump(sort_keys=False) preserves that order; a Go map does not.
//  3. Quote and flow style must survive on hand-authored files, which
//     `deviation declare` (bin/company-os:1121) and `exception request` (:1136)
//     read-modify-write.
//
// What this file does NOT reproduce, measured against the vendored PyYAML
// 6.0.2 and recorded so the gaps are known rather than discovered:
//
//   - PyYAML's SCANNER is stricter than yaml.v3's about tabs, but NARROWLY so.
//     Measured across ten positions (TestTabDivergence): both reject a tab used
//     structurally — as indentation, before a key, after a block-sequence dash —
//     and both accept one inside a double-quoted, single-quoted or block-literal
//     scalar. They diverge on four intra-line positions only, where PyYAML
//     raises ScannerError and yaml.v3 accepts: after a `:`, inside a plain
//     scalar's content, trailing, and inside a flow collection. Because every
//     structural position agrees, the residual divergence cannot change a
//     document's shape — see R-0.7a(h). Closing it would take a scanner rewrite.
//   - `<<` merge keys are left as authored. PyYAML's SafeLoader flattens them
//     into the mapping; MapGet does not follow them, and a merge key whose value
//     is not a mapping is accepted here where PyYAML raises. A `<<` outside key
//     position IS rejected — see construct. No committed artifact uses a merge
//     key.
//   - The EMITTER's byte signature differs from safe_dump's. See
//     roundtrip_test.go, which measures exactly which committed files survive a
//     no-op round-trip and which do not. This is the divergence task 2.4
//     anticipated when it replaced the feature-index byte compare with a
//     semantic one.

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"gopkg.in/yaml.v3"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/model"
)

// emitIndent is the block indent Bytes emits at. PyYAML's emitter indents block
// collections at 2 and yaml.v3 defaults to 4, so matching PyYAML costs one
// setter call and removes one whole class of difference from every artifact the
// port re-emits. That reason holds whatever the corpus looks like; a count of
// byte-identical documents would not, because R-0.7c stops derived artifacts
// being re-emitted at all.
const emitIndent = 2

// SyntaxError reports YAML that PyYAML's safe_load refuses to load or to
// construct — its ScannerError, ParserError, ComposerError and ConstructorError
// collapsed into one Go type.
//
// Python has no try around load_yaml (bin/company-os:58) or frontmatter (:76),
// so today these surface as an uncaught YAMLError and a traceback. R-0.7a(e)
// sanctions replacing that with a diagnostic and exit code 4; carrying a
// distinct type is what lets dispatch make that mapping.
type SyntaxError struct{ Err error }

func (e *SyntaxError) Error() string { return "yamlio: " + e.Err.Error() }
func (e *SyntaxError) Unwrap() error { return e.Err }

// ExitCode is R-4.5 and R-0.7a(e): malformed YAML exits 4 with a diagnostic
// where Python exits 1 through a traceback. Implemented rather than wrapped in a
// model.Error so callers keep matching *SyntaxError with errors.As.
func (e *SyntaxError) ExitCode() model.ExitCode { return model.ExitArtifact }

func syntaxf(format string, a ...any) error {
	return &SyntaxError{Err: fmt.Errorf(format, a...)}
}

// Document is one YAML document held as a node tree, never as a struct.
//
// The zero value and a Document loaded from an empty stream both have a nil
// Root, which is Python's safe_load("") == None.
type Document struct {
	node *yaml.Node // a DocumentNode, or nil for an empty stream
}

// Load parses one YAML document the way yaml.safe_load parses it.
//
// It is deliberately stricter than a bare yaml.Unmarshal in two measured ways:
// a stream holding more than one document is rejected, because safe_load raises
// ComposerError("expected a single document in the stream") where yaml.v3
// silently keeps the first; and every scalar is resolved through Resolve, so a
// value PyYAML resolves but cannot construct — `2035-02-30` is the realistic
// one — fails here at load time, as it does in Python, rather than at first use.
func Load(data []byte) (*Document, error) {
	dec := yaml.NewDecoder(bytes.NewReader(data))

	var doc yaml.Node
	switch err := dec.Decode(&doc); {
	case errors.Is(err, io.EOF):
		return &Document{}, nil // safe_load of an empty stream is None
	case err != nil:
		return nil, &SyntaxError{Err: err}
	}

	var extra yaml.Node
	switch err := dec.Decode(&extra); {
	case errors.Is(err, io.EOF): // the only accepting case
	case err != nil:
		return nil, &SyntaxError{Err: err}
	default:
		return nil, syntaxf("expected a single document in the stream")
	}

	if doc.Kind != yaml.DocumentNode || len(doc.Content) == 0 {
		return &Document{}, nil
	}
	if err := construct(doc.Content[0]); err != nil {
		return nil, err
	}
	return &Document{node: &doc}, nil
}

// LoadFrontmatter is Load over the raw block internal/frontmatter hands back,
// completing `yaml.safe_load(m.group(1)) or {}` (bin/company-os:80).
//
// The `or {}` is Python's truthiness operator, not a None check: safe_load("0"),
// ("false"), ("”") and ("[]") all collapse to an empty mapping too, because 0,
// False, ” and [] are all falsy. Measured against the vendored PyYAML; see
// TestFrontmatterFalsyCollapse.
func LoadFrontmatter(raw []byte) (*Document, error) {
	d, err := Load(raw)
	if err != nil {
		return nil, err
	}
	if d.IsFalsy() {
		return NewDocument(NewMapping()), nil
	}
	return d, nil
}

// NewDocument wraps a root node so a freshly built tree can be emitted.
func NewDocument(root *yaml.Node) *Document {
	return &Document{node: &yaml.Node{Kind: yaml.DocumentNode, Content: []*yaml.Node{root}}}
}

// Root returns the document's top-level node, or nil for an empty stream.
func (d *Document) Root() *yaml.Node {
	if d == nil || d.node == nil || len(d.node.Content) == 0 {
		return nil
	}
	return d.node.Content[0]
}

// IsFalsy reports Python's truthiness of the loaded value, which is what both
// `or {}` (frontmatter, :80) and `or default` (load_yaml, :63) branch on.
func (d *Document) IsFalsy() bool { return isFalsy(d.Root()) }

func isFalsy(n *yaml.Node) bool {
	// An alias is its anchor's value, so it is falsy exactly when that value is
	// — measured: safe_load("a: &x []\nb: *x") is {'a': [], 'b': []}.
	n = Deref(n)
	if n == nil {
		return true
	}
	switch n.Kind {
	case yaml.MappingNode, yaml.SequenceNode:
		return len(n.Content) == 0
	case yaml.ScalarNode:
		s, err := Resolve(n)
		if err != nil {
			return false
		}
		switch s.Kind {
		case KindNull:
			return true
		case KindBool:
			return !s.Bool
		case KindInt:
			return s.Int == nil || s.Int.Sign() == 0
		case KindFloat:
			return s.Float == 0
		case KindStr:
			return s.Raw == ""
		}
	}
	// Timestamps and the merge/value tags are all truthy in Python.
	return false
}

// Bytes emits the document. An empty stream emits nothing, so a no-op
// round-trip of an empty file stays empty.
func (d *Document) Bytes() ([]byte, error) {
	if d.Root() == nil {
		return nil, nil
	}
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(emitIndent)
	if err := enc.Encode(d.node); err != nil {
		return nil, fmt.Errorf("yamlio: emit: %w", err)
	}
	if err := enc.Close(); err != nil {
		return nil, fmt.Errorf("yamlio: emit: %w", err)
	}
	return buf.Bytes(), nil
}

// construct walks the tree the way safe_load's construction pass walks it: it
// surfaces the errors PyYAML raises there, and it applies the two ways that
// pass CHANGES the value — duplicate keys collapse, and `<<`/`=` outside key
// position are refused. The resolved scalars themselves are discarded because
// callers read nodes.
func construct(n *yaml.Node) error { return constructNode(n, false) }

// constructNode carries whether n sits in a mapping's key position, which is
// the only place PyYAML tolerates the merge and value tags.
func constructNode(n *yaml.Node, isKey bool) error {
	switch n.Kind {
	case yaml.ScalarNode:
		s, err := Resolve(n)
		if err != nil {
			return &SyntaxError{Err: err}
		}
		// PyYAML's SafeConstructor registers no constructor for either tag, so
		// anywhere but a key it raises ConstructorError. Measured: `k: <<`,
		// `- <<`, a bare `<<` document and all three `=` forms all raise, while
		// `<<:` and `=:` as KEYS load — the first as a merge, the second as the
		// plain string "=". yaml.v3 accepts every one of them, and re-emits
		// `k: <<` as `k: !!merge <<`, materializing a tag into the author's file.
		if !isKey && (s.Kind == KindMerge || s.Kind == KindValue) {
			return syntaxf("could not determine a constructor for the tag %q", s.Kind.Tag())
		}
		return nil

	case yaml.MappingNode:
		// Every pair is constructed, including one a later duplicate will drop:
		// measured, safe_load("a: 2035-02-30\na: 1") still raises. So the walk
		// runs first and the collapse second.
		for i := 0; i+1 < len(n.Content); i += 2 {
			if err := constructNode(n.Content[i], true); err != nil {
				return err
			}
			if err := constructNode(n.Content[i+1], false); err != nil {
				return err
			}
		}
		return collapseDuplicateKeys(n)

	default:
		// Alias nodes carry no Content, so this cannot cycle on an anchor, and
		// nothing is skipped by not descending into one: the anchored node is
		// constructed at the position that defines it, and every reader reaches
		// the same node through Deref.
		for _, c := range n.Content {
			if err := constructNode(c, false); err != nil {
				return err
			}
		}
		return nil
	}
}

// collapseDuplicateKeys reduces a mapping to one pair per key, keeping the
// FIRST key's position and the LAST value.
//
// That is what building a Python dict does, and measured against the vendored
// PyYAML 6.0.2 it is visible on both sides of a read-modify-write:
// safe_load("a: 1\nb: x\na: 2\na: 3") is {'a': 3, 'b': 'x'} — 'a' first, valued
// 3 — and safe_dump writes back the one pair `a: 3`. Leaving both pairs in the
// tree made `deviation declare` and `exception request` re-emit a duplicate the
// Python CLI would have removed. R-0.7d.
//
// Keys are identified by their resolved kind and Python rendering rather than
// by raw text, because that is what Python hashes: `1` and `'1'` are distinct
// keys, while `~` and `null`, `true` and `yes`, and a plain and a quoted `a`
// each name one. The known gap is Python's cross-type numeric equality — it
// folds `1` and `1.0` into one key and this does not. A non-scalar key is left
// alone entirely; PyYAML raises "found unhashable key" for those, which is a
// separate divergence this does not widen.
func collapseDuplicateKeys(m *yaml.Node) error {
	type ident struct {
		kind Kind
		text string
	}
	first := make(map[ident]int, len(m.Content)/2)
	out := m.Content[:0:0]

	for i := 0; i+1 < len(m.Content); i += 2 {
		k, v := m.Content[i], m.Content[i+1]
		key := Deref(k)
		if key == nil || key.Kind != yaml.ScalarNode {
			out = append(out, k, v)
			continue
		}
		s, err := Resolve(key)
		if err != nil {
			return &SyntaxError{Err: err}
		}
		id := ident{s.Kind, s.String()}
		if at, seen := first[id]; seen {
			out[at+1] = v
			continue
		}
		first[id] = len(out)
		out = append(out, k, v)
	}
	m.Content = out
	return nil
}

// ---------------------------------------------------------------- node access

// The three Map* helpers scan FORWARD and stop at the first match, because a
// loaded mapping holds at most one pair per key: construct's collapse pass
// enforces that, and MapSet — the only way to add a pair — replaces rather than
// appends. Before the collapse landed, MapGet had to scan backwards to agree
// with Python on safe_load("a: 1\na: 2") == {'a': 2}; now the tree itself says
// so, on the emit side too.
//
// Each dereferences its container so `b: *x` behaves as the mapping or sequence
// the anchor names, which is what PyYAML's composer already produced.

// MapGet returns the value node for key, or nil.
func MapGet(m *yaml.Node, key string) *yaml.Node {
	m = Deref(m)
	if m == nil || m.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i+1 < len(m.Content); i += 2 {
		if keyText(m.Content[i]) == key {
			return m.Content[i+1]
		}
	}
	return nil
}

// MapSet replaces the value for key in place, preserving its position, or
// appends the pair when the key is absent. Position is what preserves authored
// key order across a read-modify-write.
func MapSet(m *yaml.Node, key string, value *yaml.Node) error {
	m = Deref(m)
	if m == nil || m.Kind != yaml.MappingNode {
		return fmt.Errorf("yamlio: MapSet wants a mapping node, got kind %d", kindOf(m))
	}
	for i := 0; i+1 < len(m.Content); i += 2 {
		if keyText(m.Content[i]) == key {
			m.Content[i+1] = value
			return nil
		}
	}
	m.Content = append(m.Content, NewString(key), value)
	return nil
}

// MapKeys returns the keys in authored order.
func MapKeys(m *yaml.Node) []string {
	m = Deref(m)
	if m == nil || m.Kind != yaml.MappingNode {
		return nil
	}
	keys := make([]string, 0, len(m.Content)/2)
	for i := 0; i+1 < len(m.Content); i += 2 {
		keys = append(keys, keyText(m.Content[i]))
	}
	return keys
}

// keyText is a key node's text, following an alias so an anchored key reads as
// the key it names rather than as the empty value an alias node carries.
func keyText(n *yaml.Node) string {
	if n = Deref(n); n == nil {
		return ""
	}
	return n.Value
}

// SeqAppend appends values to a sequence node.
func SeqAppend(s *yaml.Node, values ...*yaml.Node) error {
	s = Deref(s)
	if s == nil || s.Kind != yaml.SequenceNode {
		return fmt.Errorf("yamlio: SeqAppend wants a sequence node, got kind %d", kindOf(s))
	}
	s.Content = append(s.Content, values...)
	return nil
}

// NewString builds a plain string scalar. Style is left at zero so the emitter
// quotes it only when it must, which is safe_dump's policy too.
func NewString(s string) *yaml.Node {
	return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: s}
}

// NewMapping builds an empty block mapping.
func NewMapping() *yaml.Node {
	return &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
}

// NewSequence builds an empty block sequence.
func NewSequence() *yaml.Node {
	return &yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq"}
}
