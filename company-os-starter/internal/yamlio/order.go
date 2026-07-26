package yamlio

// Deterministic ordering (task 1.4, R-0.11).
//
// Python dicts are insertion-ordered and safe_dump(sort_keys=False) preserves
// that order; Go maps randomize iteration by design. Three sites in
// bin/company-os let a dict's order reach byte-frozen output, and the order
// they need is NOT the same at each one. It was measured, not assumed:
//
//	[A] workspace.lock.yaml `files:` (bin/company-os:2614). Built by
//	    _materialize_all (:2454) as a nested walk — for each slice in MANIFEST
//	    order, for each entry in that slice's `paths:` list order, for each file
//	    in sorted(src.rglob("*")) — then emitted with sort_keys=False. The
//	    committed examples/federated/workspace.lock.yaml records
//	    governance/… before components/…, which is the manifest's paths: order
//	    and the REVERSE of sorted order. A global sort here is wrong.
//
//	[B] gate 8's finding loop, `for rel, want in (lr.get("files") or {}).items()`
//	    (:2521). `lr` is safe_load of the lock file, so this is the lock's
//	    DOCUMENT order — which is [A]'s emission order arriving back through the
//	    parser. examples/failing-federated-golden-validate.txt freezes it:
//	    governance/requirements.yaml is reported before components/svc-sliced.yaml,
//	    again the reverse of sorted. Swapping the two lines in the fixture's lock
//	    swaps the two [FAIL] lines, which is how the dependence was confirmed.
//
//	[C] gate 6's feature_index_unresolved (:1519), `for cid, e in
//	    (idx.get("components") or {}).items()`. Here insertion order HAPPENS to
//	    be sorted, because build_feature_index (:1440) iterates
//	    `cids = sorted(...)`. The loop still does not sort; the builder does.
//
// The single rule that covers all three: never iterate a Go map where order is
// observable. Read order off the node tree (MapPairs), and build order into an
// OrderedMap. Both are provided here; neither exposes a map[string]T to range
// over. A fourth ordering lives in the same data — aggregate_hash (:2436) runs
// `for rel in sorted(files)`, a plain string sort over the very keys [A] emits
// in walk order — so OrderedMap offers Keys and SortedKeys separately and the
// caller must pick deliberately.

import (
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// Pair is one key/value of a mapping node.
type Pair struct {
	Key   string
	Value *yaml.Node
}

// MapPairs returns a mapping's pairs in authored document order — the Go
// equivalent of iterating a PyYAML-loaded dict, which is what sites [B] and [C]
// above do. Prefer it over MapKeys+MapGet: it is one pass rather than O(n²),
// and it cannot silently resolve to a different pair when duplicate keys
// survived (they do not — Load collapses them — but the property should not
// depend on that).
//
// Returns nil for a nil or non-mapping node, mirroring MapKeys.
func MapPairs(m *yaml.Node) []Pair {
	m = Deref(m)
	if m == nil || m.Kind != yaml.MappingNode {
		return nil
	}
	pairs := make([]Pair, 0, len(m.Content)/2)
	for i := 0; i+1 < len(m.Content); i += 2 {
		pairs = append(pairs, Pair{Key: keyText(m.Content[i]), Value: m.Content[i+1]})
	}
	return pairs
}

// OrderedMap is an insertion-ordered string-keyed mapping with Python dict
// assignment semantics, for the one case the node tree cannot supply an order:
// a mapping the port BUILDS rather than loads, which then reaches emission.
//
// Semantics deliberately copied from dict:
//
//   - Set on a new key appends it.
//   - Set on an existing key REPLACES the value and KEEPS the original
//     position. This is the subtle one: it is why re-materializing a slice
//     whose file also appeared in an earlier slice does not move that file to
//     the end of workspace.lock.yaml.
//
// The internal index is a Go map, and it is never ranged over — every ordered
// answer comes from the keys slice. That is the whole point of the type.
type OrderedMap struct {
	keys  []string
	vals  []*yaml.Node
	index map[string]int
}

// NewOrderedMap returns an empty OrderedMap.
func NewOrderedMap() *OrderedMap {
	return &OrderedMap{index: map[string]int{}}
}

// Set assigns key to value with dict `d[k] = v` semantics.
func (o *OrderedMap) Set(key string, value *yaml.Node) {
	if o.index == nil {
		o.index = map[string]int{}
	}
	if i, ok := o.index[key]; ok {
		o.vals[i] = value
		return
	}
	o.index[key] = len(o.keys)
	o.keys = append(o.keys, key)
	o.vals = append(o.vals, value)
}

// SetString assigns key to a plain string scalar. The path→sha256 map of site
// [A] is entirely string→string.
func (o *OrderedMap) SetString(key, value string) { o.Set(key, NewString(value)) }

// Update merges other into o with dict.update semantics: keys already present
// keep their position and take other's value; new keys append in other's order.
// This is `files.update(hash_tree(...))` at bin/company-os:2471.
func (o *OrderedMap) Update(other *OrderedMap) {
	if other == nil {
		return
	}
	for i, k := range other.keys {
		o.Set(k, other.vals[i])
	}
}

// Get returns the value for key.
func (o *OrderedMap) Get(key string) (*yaml.Node, bool) {
	i, ok := o.index[key]
	if !ok {
		return nil, false
	}
	return o.vals[i], true
}

// Len returns the number of pairs.
func (o *OrderedMap) Len() int { return len(o.keys) }

// Keys returns the keys in insertion order — the order safe_dump(sort_keys=False)
// would emit, and therefore the order site [A] must reproduce.
func (o *OrderedMap) Keys() []string {
	return append([]string(nil), o.keys...)
}

// SortedKeys returns the keys in plain byte-wise string order — Python's
// `sorted(files)`, which aggregate_hash (bin/company-os:2436) uses over the same
// key set that Keys returns in walk order. The two differ; picking the wrong one
// changes sliceHash or changes the lock's byte layout, so they are separate
// methods rather than an argument.
//
// Note this is NOT PathLess: aggregate_hash sorts the keys as strings, not as
// paths.
func (o *OrderedMap) SortedKeys() []string {
	keys := o.Keys()
	sort.Strings(keys)
	return keys
}

// Pairs returns the pairs in insertion order.
func (o *OrderedMap) Pairs() []Pair {
	pairs := make([]Pair, 0, len(o.keys))
	for i, k := range o.keys {
		pairs = append(pairs, Pair{Key: k, Value: o.vals[i]})
	}
	return pairs
}

// Node renders the map as a block mapping node in insertion order, ready for
// Bytes. A fresh node is built on every call so the result cannot be aliased
// into two documents.
func (o *OrderedMap) Node() *yaml.Node {
	n := NewMapping()
	n.Content = make([]*yaml.Node, 0, len(o.keys)*2)
	for i, k := range o.keys {
		n.Content = append(n.Content, NewString(k), o.vals[i])
	}
	return n
}

// PathLess reports whether posix path a sorts before b under CPython's
// PurePath ordering — what `sorted(src.rglob("*"))` at bin/company-os:2422
// produces, and therefore part of site [A]'s emission order.
//
// CPython compares paths COMPONENT-WISE (PurePath.__lt__ on _parts_normcase),
// not as raw strings, and the two disagree whenever a separator meets a
// character below '/' (0x2F). Measured under Python 3.12:
//
//	sorted(Path)   sdd/adr  sdd/adr/a.md  sdd/adr/z.md  sdd/adr-x.md  sdd/adr.md
//	sorted(str)    sdd/adr  sdd/adr-x.md  sdd/adr.md    sdd/adr/a.md  sdd/adr/z.md
//
// so sort.Strings is not a substitute. Comparison is case-sensitive: casefolding
// is a Windows-flavour behaviour in CPython and this port targets posix.
//
// Equal component sequences fall back to a raw string compare, which keeps the
// relation a total order for inputs that differ only in redundant separators.
func PathLess(a, b string) bool {
	ca, cb := pathParts(a), pathParts(b)
	n := len(ca)
	if len(cb) < n {
		n = len(cb)
	}
	for i := 0; i < n; i++ {
		if ca[i] != cb[i] {
			return ca[i] < cb[i]
		}
	}
	if len(ca) != len(cb) {
		return len(ca) < len(cb)
	}
	return a < b
}

// SortPaths sorts paths in place under PathLess.
func SortPaths(paths []string) {
	sort.SliceStable(paths, func(i, j int) bool { return PathLess(paths[i], paths[j]) })
}

// pathParts splits a posix path the way PurePosixPath.parts does: a leading
// separator becomes its own first component, and empty components from repeated
// or trailing separators are dropped.
func pathParts(p string) []string {
	var parts []string
	if strings.HasPrefix(p, "/") {
		parts = append(parts, "/")
	}
	for _, c := range strings.Split(p, "/") {
		if c != "" {
			parts = append(parts, c)
		}
	}
	return parts
}
