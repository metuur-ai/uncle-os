// Package yamlio is the only place YAML is parsed or emitted.
//
// Responsibility: yaml.Node helpers, YAML-1.1 scalar resolution (PyYAML is
// YAML 1.1, gopkg.in/yaml.v3 is YAML 1.2 — an unquoted date must stay a string),
// canonical emit, and deterministic map ordering so byte-frozen output and
// workspace.lock.yaml reproduce exactly.
//
// Scalar resolution is scalar.go (task 1.2); node-level document access, the
// frontmatter seam and emit are document.go (task 1.3). Deterministic map
// ordering is still to come — task 1.4.
//
// Round-trip fidelity is measured, not assumed: roundtrip_test.go loads and
// re-emits every git-tracked YAML document under examples/ and freezes the
// exact set that does not come back byte-identical, with its cause, in
// testdata/roundtrip-divergent.txt.
package yamlio
