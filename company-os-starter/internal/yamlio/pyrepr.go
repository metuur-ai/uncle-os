package yamlio

// str() and repr() for the PyValue object graph.
//
// pyobject.go already answers both for a *yaml.Node. This file answers them for
// the objects PyLoadFile hands back, for the same reason PyFalsy exists twice:
// a caller that loaded a document as PyValue (because it also has to re-emit it)
// cannot reach the node-level answer without re-parsing.
//
// The site that forces this is bin/company-os:2646 —
//
//	f"drifted (manifest pin {kind}:{ref} != lock {pin_lock})"
//
// where `pin_lock` is a dict straight out of workspace.lock.yaml, so the f-string
// interpolates a Python dict repr: `{'commit': 'abc'}`. Two more are lists:
// `pin key(s) {floating}` (:2211) and `(got {present or 'neither'})` (:2215).
// Getting the quoting or the ", " spacing wrong is a byte divergence on a
// stderr line the differential harness compares.
//
// TestPyReprAgreesWithNodeRepr pins this file to pyobject.go so the two answers
// cannot drift.

// PyRepr is repr(value). A container renders its ELEMENTS with repr(), which is
// why a list of strings shows quotes that str() of the same string would not.
func PyRepr(v PyValue) string {
	switch t := v.(type) {
	case nil:
		return "None"
	case PyStr:
		return pyStrRepr(string(t))
	case PySeq:
		parts := make([]string, 0, len(t))
		for _, e := range t {
			parts = append(parts, PyRepr(e))
		}
		return "[" + joinComma(parts) + "]"
	case PyMap:
		parts := make([]string, 0, len(t))
		for _, p := range t {
			parts = append(parts, pyStrRepr(p.K)+": "+PyRepr(p.V))
		}
		return "{" + joinComma(parts) + "}"
	case PyTime:
		if s, err := constructTimestamp(string(t)); err == nil {
			return s.pyTimestampRepr()
		}
		return pyStrRepr(string(t))
	}
	// None, True/False, ints and floats repr() exactly as they str().
	return PyString(v)
}

// PyString is str(value). It differs from PyRepr only for a str — and for a
// container, whose str() *is* its repr().
//
// Note PyBool renders "True"/"False" here, not the "true"/"false" that
// PyBool.pyRepr emits: that method answers "how does safe_dump serialize this",
// a different question asked on the way out.
func PyString(v PyValue) string {
	switch t := v.(type) {
	case nil:
		return "None"
	case PyStr:
		return string(t)
	case PyNull:
		return "None"
	case PyBool:
		if bool(t) {
			return "True"
		}
		return "False"
	case PyInt:
		if t.N == nil {
			return "0"
		}
		return t.N.String()
	case PyFloat:
		return pyFloat(float64(t))
	case PyTime:
		if s, err := constructTimestamp(string(t)); err == nil {
			return s.pyTimestamp()
		}
		return string(t)
	case PySeq, PyMap:
		return PyRepr(v)
	}
	return ""
}

// PyStrings is repr() of a list of Python strs, for the two sites that build
// one in Go rather than loading it (`floating` and `present` in repo_pin).
func PyStrings(items []string) string {
	seq := make(PySeq, 0, len(items))
	for _, s := range items {
		seq = append(seq, PyStr(s))
	}
	return PyRepr(seq)
}

func joinComma(parts []string) string {
	out := ""
	for i, p := range parts {
		if i > 0 {
			out += ", "
		}
		out += p
	}
	return out
}
