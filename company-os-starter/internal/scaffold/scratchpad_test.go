package scaffold

import "testing"

// pathJoin is `Path(a) / b`, which is not filepath.Join: pathlib drops "." and
// duplicate separators but KEEPS "..", where Clean resolves it. The whole
// difference is the one line `scratchpad init` prints, but R-0.7 lists no
// carve-out for it.
//
// Measured with CPython pathlib:
//
//	python3 -c 'from pathlib import Path; print(Path("a/..") / "scratchpad")'
func TestPathJoinMatchesPathlib(t *testing.T) {
	cases := map[string]string{
		".":       "scratchpad",
		"./":      "scratchpad",
		"a":       "a/scratchpad",
		"a/":      "a/scratchpad",
		"./a":     "a/scratchpad",
		"a/..":    "a/../scratchpad",
		"a/../..": "a/../../scratchpad",
		"..":      "../scratchpad",
		"a/./b":   "a/b/scratchpad",
		"a//b":    "a/b/scratchpad",
		"/abs":    "/abs/scratchpad",
		"/":       "/scratchpad",
	}
	for in, want := range cases {
		if got := pathJoin(in, "scratchpad"); got != want {
			t.Errorf("pathJoin(%q, \"scratchpad\") = %q, want %q", in, got, want)
		}
	}
}
