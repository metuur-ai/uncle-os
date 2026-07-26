// Package yamlio is the only place YAML is parsed or emitted.
//
// Responsibility: yaml.Node helpers, YAML-1.1 scalar resolution (PyYAML is
// YAML 1.1, gopkg.in/yaml.v3 is YAML 1.2 — an unquoted date must stay a string),
// canonical emit, and deterministic map ordering so byte-frozen output and
// workspace.lock.yaml reproduce exactly.
//
// Scalar resolution is scalar.go (task 1.2); node-level document access, the
// frontmatter seam and emit are document.go (task 1.3); deterministic ordering
// is order.go (task 1.4).
//
// Ordering is MEASURED against the committed fixtures, and the answer is not
// "sorted": workspace.lock.yaml's files: map and gate 8's [FAIL] lines both
// follow Python dict INSERTION order, which the fixtures record as the reverse
// of alphabetical. See order.go's header for the three sites and their
// evidence.
//
// Round-trip fidelity is measured, not assumed: roundtrip_test.go loads and
// re-emits every git-tracked YAML document under examples/ and freezes the
// exact set that does not come back byte-identical, with its cause, in
// testdata/roundtrip-divergent.txt.
package yamlio
