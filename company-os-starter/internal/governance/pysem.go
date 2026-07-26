package governance

// The Python expressions this cluster leans on that internal/yamlio does not
// already answer at the PyValue level.
//
// Everything here is a two-line adapter over yamlio (PyFalsy, PyEqual,
// PyString) rather than a reimplementation — the scalar semantics, the
// truthiness rule and the repr all live there, and a second copy of any of them
// is how the two answers drift. What is local is only the SHAPE of the access:
// `m[key]`, `m.get(key, default)`, `x in y`.

import (
	"path/filepath"
	"strings"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/model"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/yamlio"
)

// loadOr is load_yaml(path, default) (`bin/company-os:56-62`), including the
// `or default` half: PyYAML returns the parsed document only when it is TRUTHY,
// so `[]`, `{}`, `0` and the empty string all collapse to the default (R-1.7a).
func loadOr(path string, def pyVal) (pyVal, error) {
	v, err := yamlio.PyLoadFile(path)
	if err != nil {
		return nil, err
	}
	if yamlio.PyFalsy(v) {
		return def, nil
	}
	return v, nil
}

// seqAt is `container.get(key, [])` where the result is then iterated. The two
// refusals are R-0.7a(j): `.get` on a non-mapping raises AttributeError and
// iterating a scalar raises TypeError, and Python writes nothing on either.
//
// A dict or a str IS iterable in Python, but iterating one yields keys or
// characters, which every caller here immediately subscripts as a mapping — so
// the refusal happens one step later there and at the same observable outcome.
func seqAt(container pyVal, key, path string) (pySeq, error) {
	m, ok := container.(pyMap)
	if !ok {
		return nil, model.Errorf(model.ExitArtifact,
			"%s: expected a mapping at the document root", path)
	}
	v := m.Get(key)
	if v == nil {
		return nil, nil
	}
	s, ok := v.(pySeq)
	if !ok {
		return nil, model.Errorf(model.ExitArtifact, "%s: '%s' must be a sequence", path, key)
	}
	return s, nil
}

// index is `m[key]`: a KeyError when the key is absent. Only absence returns
// nil from PyMap.Get — a present `key: null` comes back as PyNull.
func index(m pyMap, key, what string) (pyVal, error) {
	v := m.Get(key)
	if v == nil {
		return nil, model.Errorf(model.ExitArtifact, "%s: missing required key '%s'", what, key)
	}
	return v, nil
}

// getDefault is `m.get(key, default)`: the default applies to an ABSENT key
// only, so an explicit null comes back as None rather than as the default.
func getDefault(m pyMap, key string, def pyVal) pyVal {
	if v := m.Get(key); v != nil {
		return v
	}
	return def
}

// contains is `item in container`.
//
// The str case is not defensive padding: `appliesTo: {relationships: belongs-to}`
// is valid YAML, and Python's `in` on a str is a SUBSTRING test, so
// `"belongs-to" in "belongs-to-x"` is true and the requirement applies. Treating
// a str as a one-element collection would silently drop it.
func contains(container, item pyVal) (bool, error) {
	switch c := container.(type) {
	case pySeq:
		for _, e := range c {
			if yamlio.PyEqual(e, item) {
				return true, nil
			}
		}
		return false, nil
	case pyMap:
		for _, p := range c {
			if yamlio.PyEqual(pyStr(p.K), item) {
				return true, nil
			}
		}
		return false, nil
	case pyStr:
		s, ok := item.(pyStr)
		if !ok {
			// `in <str>` requires a str on the left; anything else is TypeError.
			return false, errNotContainer
		}
		return strings.Contains(string(c), string(s)), nil
	}
	return false, errNotContainer
}

// errNotContainer marks the `x in y` TypeError so the caller can name the field.
var errNotContainer = model.Errorf(model.ExitArtifact, "argument of type is not iterable")

// localName is `s.split("://")[-1]`.
func localName(s string) string {
	if i := strings.LastIndex(s, "://"); i >= 0 {
		return s[i+len("://"):]
	}
	return s
}

// relTo is Path.relative_to(ws.root) in POSIX form (R-1.12).
func relTo(root, path string) string {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return filepath.ToSlash(path)
	}
	return filepath.ToSlash(rel)
}
