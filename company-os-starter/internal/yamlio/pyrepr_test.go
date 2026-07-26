package yamlio

import "testing"

// TestPyReprAgreesWithNodeRepr is the pin that keeps the object-level answers in
// pyrepr.go from drifting away from the node-level answers in pyobject.go. Both
// files answer str() and repr(); they must never disagree, or a message composed
// from a PyValue would differ from the same message composed from a node.
func TestPyReprAgreesWithNodeRepr(t *testing.T) {
	docs := []string{
		`{commit: abc123}`,
		`{commit: '0000000000000000000000000000000000000000'}`,
		`{commit: 1111111111111111111111111111111111111111}`,
		`{branch: main, ref: HEAD}`,
		`[commit, tag]`,
		`[]`,
		`{}`,
		`{a: null, b: true, c: 3, d: 3.5, e: "it's"}`,
		`{nested: {x: [1, 'two', null]}}`,
		`plain string`,
		`2026-07-26`,
		`2026-07-26 11:30:00`,
		`{d: 2026-07-26}`,
	}
	for _, src := range docs {
		doc, err := Load([]byte(src))
		if err != nil {
			t.Fatalf("Load(%q): %v", src, err)
		}
		node := doc.Root()
		obj, err := fromNode(node, "test")
		if err != nil {
			t.Fatalf("fromNode(%q): %v", src, err)
		}
		if got, want := PyRepr(obj), pyRepr(node); got != want {
			t.Errorf("PyRepr(%q) = %s, node repr = %s", src, got, want)
		}
		if got, want := PyString(obj), PyText(node); got != want {
			t.Errorf("PyString(%q) = %s, PyText = %s", src, got, want)
		}
	}
}

// The three sites this file exists for, spelled out. Each is a fragment of a
// message the differential harness compares byte-for-byte.
func TestPyReprRendersTheFederationMessageFragments(t *testing.T) {
	// bin/company-os:2646 — `!= lock {pin_lock}`.
	pin := PyMap{{K: "commit", V: PyStr("abc")}}
	if got := PyRepr(pin); got != `{'commit': 'abc'}` {
		t.Errorf("pin repr = %s", got)
	}
	if got := PyRepr(PyMap{}); got != "{}" {
		t.Errorf("empty pin repr = %s", got)
	}
	// :2211 — `pin key(s) {floating}`; :2215 — `(got {present})`.
	if got := PyStrings([]string{"branch"}); got != `['branch']` {
		t.Errorf("floating repr = %s", got)
	}
	if got := PyStrings([]string{"commit", "tag"}); got != `['commit', 'tag']` {
		t.Errorf("present repr = %s", got)
	}
	if got := PyStrings(nil); got != "[]" {
		t.Errorf("empty list repr = %s", got)
	}
}

// str(True) is "True", but PyBool.pyRepr — which answers "how does safe_dump
// serialize this" — is "true". The two must not be confused.
func TestPyStringOfBoolIsPythonNotYAML(t *testing.T) {
	if got := PyString(PyBool(true)); got != "True" {
		t.Errorf("PyString(true) = %s", got)
	}
	text, _, err := PyBool(true).pyRepr()
	if err != nil || text != "true" {
		t.Errorf("PyBool.pyRepr = %s, %v", text, err)
	}
}
