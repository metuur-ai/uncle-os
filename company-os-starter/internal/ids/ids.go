// Package ids owns company-ontology/ids/registry.yaml — the single canonical
// source of canonical IDs (GPF-R-2.1) — and the `ids list` command built on it.
//
// It is a leaf with fan-in rather than a command package: Suggest is reached
// from `governance explain` (bin/company-os:365), and register_id (`:1813`) is
// reached from `init` and `add` (`:1950`, `:1951`, `:2008`, `:2015`, `:2025`).
// That is why the registry lives here and not inside whichever command happens
// to read it first.
//
// Not ported here, deliberately: register_id. Its only callers are in
// internal/scaffold, which is landing in a sibling task, and R-0.7c now requires
// that write to be guarded on a semantic compare of the parsed structure rather
// than emitted unconditionally as `:1823` does. When it lands it belongs in this
// package, with a one-way scaffold -> ids dependency, for the same reason
// rebuild_generated belongs in internal/graph.
//
// Nothing here composes a sentence: List returns records whose Fields carry the
// data, and internal/render turns them into the line the Python CLI prints
// (R-2.8, R-2.12).
package ids

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/model"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/roles"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/workspace"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/yamlio"
	"gopkg.in/yaml.v3"
)

// RegistryPath is the registry's workspace-relative path. It is rendered
// verbatim inside the listing header, so it is data, not a derived string.
const RegistryPath = "company-ontology/ids/registry.yaml"

// Entry is one row of the registry's `ids:` sequence, already reduced to the
// two strings `cmd_ids` reads (`:1286-1287`). Both are the Python str() of the
// authored value, so an unquoted date or an int id renders as PyYAML renders it.
type Entry struct {
	ID        string
	DefinedIn string
}

// Filter is the --prefix/--team/--platform triple. An empty field is an
// inactive filter, exactly as `if args.prefix` reads in Python; all active
// filters must pass (`:1289-1296` uses three independent `continue`s).
type Filter struct {
	Prefix   string
	Team     string
	Platform string
}

// Match reports whether an entry survives every active filter.
func (f Filter) Match(e Entry) bool {
	if f.Prefix != "" && !strings.HasPrefix(e.ID, f.Prefix) {
		return false
	}
	if f.Team != "" && !strings.HasPrefix(e.DefinedIn, "teams/"+f.Team+"/") &&
		e.ID != "team://"+f.Team {
		return false
	}
	if f.Platform != "" && !strings.HasPrefix(e.DefinedIn, "platforms/"+f.Platform+"/") &&
		e.ID != "platform://"+f.Platform {
		return false
	}
	return true
}

// Load is load_registry (`bin/company-os:1210-1217`). A missing, empty, or
// non-mapping registry yields no entries rather than an error, so `ids list`
// prints its helpful line instead of crashing. Rows without a truthy `id` are
// dropped, as the Python comprehension drops them.
func Load(ws *workspace.Workspace) ([]Entry, error) {
	root, err := loadYAMLFile(filepath.Join(ws.Root, filepath.FromSlash(RegistryPath)))
	if err != nil {
		return nil, err
	}
	root = yamlio.Deref(root)
	// isinstance(data, dict) — anything else, including a falsy document that
	// `or None` collapsed, resolves to no entries.
	if root == nil || root.Kind != yaml.MappingNode {
		return nil, nil
	}
	seq := yamlio.Deref(yamlio.MapGet(root, "ids"))
	// `(data.get("ids") or [])` is truthiness: a null, empty, or falsy value is
	// an empty list, and a non-sequence is not iterable as rows here.
	if seq == nil || seq.Kind != yaml.SequenceNode {
		return nil, nil
	}

	var out []Entry
	for _, item := range seq.Content {
		item = yamlio.Deref(item)
		// isinstance(e, dict)
		if item == nil || item.Kind != yaml.MappingNode {
			continue
		}
		id, ok, err := truthyString(item, "id")
		if err != nil {
			return nil, err
		}
		if !ok {
			continue
		}
		defined, _, err := truthyString(item, "definedIn")
		if err != nil {
			return nil, err
		}
		out = append(out, Entry{ID: id, DefinedIn: defined})
	}
	return out, nil
}

// List is cmd_ids (`bin/company-os:1275-1302`), as records.
//
// The section order reproduces the print order: the --role legend comes first
// and is emitted even when the registry turns out to be empty, because Python
// prints it before the emptiness check.
func List(ws *workspace.Workspace, role string, filter Filter) ([]model.GateResult, error) {
	var sections []model.GateResult
	if g, ok := roles.GlossarySection(role, 1); ok {
		sections = append(sections, g)
	}

	entries, err := Load(ws)
	if err != nil {
		return nil, err
	}

	registry := model.GateResult{Ordinal: len(sections) + 1, Slug: model.SlugRegistry}
	if len(entries) == 0 {
		registry.Findings = append(registry.Findings, model.Finding{
			Severity: model.SevOK,
			Code:     model.CodeRegistryEmpty,
			Path:     RegistryPath,
			Fields:   model.Fields{"registry": RegistryPath},
		})
		return append(sections, registry), nil
	}

	registry.Findings = append(registry.Findings, model.Finding{
		Severity: model.SevOK,
		Code:     model.CodeListingHeader,
		Path:     RegistryPath,
		Fields:   model.Fields{"registry": RegistryPath},
	})

	matched := 0
	for _, e := range entries {
		if !filter.Match(e) {
			continue
		}
		matched++
		fields := model.Fields{"id": e.ID, "definedIn": e.DefinedIn}
		// scheme/localName are not printed. They are the two axes a listing UI
		// groups and filters on, and re-deriving them from a display string is
		// exactly the flattening this port exists to remove.
		if scheme, local, split := strings.Cut(e.ID, "://"); split {
			fields["scheme"] = scheme
			fields["localName"] = local
		}
		registry.Findings = append(registry.Findings, model.Finding{
			Severity: model.SevOK,
			Code:     model.CodeEntry,
			Subject:  e.ID,
			Path:     e.DefinedIn,
			Fields:   fields,
		})
	}

	count := model.Fields{"matched": matched, "total": len(entries)}
	if filter.Prefix != "" {
		count["prefix"] = filter.Prefix
	}
	if filter.Team != "" {
		count["team"] = filter.Team
	}
	if filter.Platform != "" {
		count["platform"] = filter.Platform
	}
	registry.Findings = append(registry.Findings, model.Finding{
		Severity: model.SevOK,
		Code:     model.CodeCount,
		Fields:   count,
	})
	return append(sections, registry), nil
}

// truthyString returns the Python str() of a mapping value and whether Python
// would consider it truthy. An absent key is ("", false), which is what
// `e.get("id")` guards on and what `str(e.get("definedIn", ""))` defaults to.
func truthyString(m *yaml.Node, key string) (string, bool, error) {
	v := yamlio.Deref(yamlio.MapGet(m, key))
	if v == nil {
		return "", false, nil
	}
	switch v.Kind {
	case yaml.MappingNode, yaml.SequenceNode:
		// A container is truthy when non-empty. Its text falls back to "" rather
		// than to Python's repr of a dict or list: an `id:` holding a container
		// is a shape no registry has, and reproducing repr() for it is not worth
		// the code.
		return "", len(v.Content) > 0, nil
	case yaml.ScalarNode:
		s, err := yamlio.Resolve(v)
		if err != nil {
			return "", false, err
		}
		return s.String(), scalarTruthy(s), nil
	}
	return "", false, nil
}

// scalarTruthy is Python's bool() of the value safe_load produced. It matches
// yamlio's own `or {}` rule: null, false, an empty string, and a zero number are
// falsy; a date never is.
func scalarTruthy(s yamlio.Scalar) bool {
	switch s.Kind {
	case yamlio.KindNull:
		return false
	case yamlio.KindBool:
		return s.Bool
	case yamlio.KindInt:
		return s.Int != nil && s.Int.Sign() != 0
	case yamlio.KindFloat:
		return s.Float != 0
	case yamlio.KindTimestamp:
		return true
	default:
		return s.Raw != ""
	}
}

// loadYAMLFile is load_yaml(path, None) (`bin/company-os:58-63`) reduced to the
// node it returns: an absent file is (nil, nil), never an error.
func loadYAMLFile(path string) (*yaml.Node, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, model.Errorf(model.ExitArtifact, "%s: %v", path, err)
	}
	doc, err := yamlio.Load(data)
	if err != nil {
		return nil, model.Errorf(model.ExitArtifact, "%s: %v", path, err)
	}
	return doc.Root(), nil
}
