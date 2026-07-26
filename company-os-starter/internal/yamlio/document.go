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
//   - PyYAML's SCANNER is stricter than yaml.v3's. `type:\tprd` raises
//     ScannerError under PyYAML and parses cleanly under yaml.v3 — an
//     accept/reject divergence that only a scanner rewrite would close.
//   - `<<` merge keys are left as authored. PyYAML's SafeLoader flattens them
//     into the mapping; MapGet does not follow them. No committed artifact uses
//     a merge key.
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
)

// emitIndent is the block indent Bytes emits at. PyYAML's emitter uses 2, and
// 2 also reproduces more of the committed corpus byte-for-byte than yaml.v3's
// default of 4 (measured: 46 of 112 documents versus 43).
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
	// Timestamps, aliases and the merge/value tags are all truthy in Python.
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

// construct walks the tree resolving every scalar, which is what safe_load's
// construction pass does. Its only job is to surface the errors PyYAML raises
// there; the resolved values are discarded because callers read nodes.
func construct(n *yaml.Node) error {
	if n.Kind == yaml.ScalarNode {
		if _, err := Resolve(n); err != nil {
			return &SyntaxError{Err: err}
		}
		return nil
	}
	// Alias nodes carry no Content, so this cannot cycle on an anchor.
	for _, c := range n.Content {
		if err := construct(c); err != nil {
			return err
		}
	}
	return nil
}

// ---------------------------------------------------------------- node access

// MapGet returns the value node for key, or nil.
//
// It scans backwards because PyYAML accepts duplicate keys and the resulting
// dict keeps the LAST value — measured: safe_load("a: 1\na: 2") is {'a': 2},
// while the node tree keeps both pairs.
func MapGet(m *yaml.Node, key string) *yaml.Node {
	if m == nil || m.Kind != yaml.MappingNode {
		return nil
	}
	for i := len(m.Content) - 2; i >= 0; i -= 2 {
		if m.Content[i].Value == key {
			return m.Content[i+1]
		}
	}
	return nil
}

// MapSet replaces the value for key in place, preserving its position, or
// appends the pair when the key is absent. Position is what preserves authored
// key order across a read-modify-write.
func MapSet(m *yaml.Node, key string, value *yaml.Node) error {
	if m == nil || m.Kind != yaml.MappingNode {
		return fmt.Errorf("yamlio: MapSet wants a mapping node, got kind %d", kindOf(m))
	}
	for i := len(m.Content) - 2; i >= 0; i -= 2 {
		if m.Content[i].Value == key {
			m.Content[i+1] = value
			return nil
		}
	}
	m.Content = append(m.Content, NewString(key), value)
	return nil
}

// MapKeys returns the keys in authored order. A duplicate key appears once, at
// its first position, matching how a Python dict literal orders one.
func MapKeys(m *yaml.Node) []string {
	if m == nil || m.Kind != yaml.MappingNode {
		return nil
	}
	seen := make(map[string]bool, len(m.Content)/2)
	keys := make([]string, 0, len(m.Content)/2)
	for i := 0; i+1 < len(m.Content); i += 2 {
		k := m.Content[i].Value
		if !seen[k] {
			seen[k] = true
			keys = append(keys, k)
		}
	}
	return keys
}

// SeqAppend appends values to a sequence node.
func SeqAppend(s *yaml.Node, values ...*yaml.Node) error {
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
