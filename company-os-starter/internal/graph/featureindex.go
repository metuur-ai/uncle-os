package graph

// The feature index: bin/company-os:1400-1537.
//
// A derived, always-current map from each catalog component to all of its
// artifacts — reality doc, active and archived PRDs, discovery ids, pending
// outcome reviews, external pointers — following the reference edges already
// present in frontmatter. Written under generated/, never hand-edited, and
// produced by ONE builder that feeds both the writer here and the drift gate in
// validate, so write-then-validate cannot drift.

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/model"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/workspace"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/yamlio"
)

// PointerErrors is pointer_errors (`:147-168`): a `pointers:` list stores
// external references that are never fetched, and each entry needs `label` +
// `system` + at least one of `url`/`id`. Shape errors only — the target is not
// resolved.
func PointerErrors(meta yamlio.PyMap) []string {
	ptrs := meta.Get("pointers")
	if ptrs == nil || yamlio.PyIsNone(ptrs) {
		return nil
	}
	list, ok := ptrs.(yamlio.PySeq)
	if !ok {
		return []string{"pointers: must be a list"}
	}
	var errs []string
	for i, p := range list {
		m, ok := p.(yamlio.PyMap)
		if !ok {
			errs = append(errs, fmt.Sprintf("pointers[%d]: must be a mapping", i))
			continue
		}
		if yamlio.PyFalsy(m.Get("label")) {
			errs = append(errs, fmt.Sprintf("pointers[%d]: missing 'label'", i))
		}
		if yamlio.PyFalsy(m.Get("system")) {
			errs = append(errs, fmt.Sprintf("pointers[%d]: missing 'system'", i))
		}
		if yamlio.PyFalsy(m.Get("url")) && yamlio.PyFalsy(m.Get("id")) {
			errs = append(errs, fmt.Sprintf("pointers[%d]: needs 'url' or 'id'", i))
		}
	}
	return errs
}

// CollectPointers is collect_pointers (`:1421-1431`): merge, dedupe and sort the
// well-formed `pointers:` of several documents (R-2.3). A malformed pointer is
// dropped silently here — the warning belongs to gate 4, not to the index.
func CollectPointers(metas ...yamlio.PyMap) yamlio.PySeq {
	var seen yamlio.PySeq
	for _, meta := range metas {
		if meta == nil {
			continue
		}
		list, ok := meta.Get("pointers").(yamlio.PySeq)
		if !ok {
			continue
		}
		for _, p := range list {
			pm, ok := p.(yamlio.PyMap)
			if !ok {
				continue
			}
			if containsValue(seen, pm) {
				continue
			}
			if len(PointerErrors(yamlio.PyMap{{K: "pointers", V: yamlio.PySeq{pm}}})) > 0 {
				continue
			}
			seen = append(seen, pm)
		}
	}
	// key=(str(label), str(system), str(url or id or "")) — a stable sort, so
	// two pointers identical on all three keep their discovery order.
	sort.SliceStable(seen, func(i, j int) bool {
		return pointerKey(seen[i]) < pointerKey(seen[j])
	})
	return seen
}

// pointerKey flattens the three-element sort key into one comparable string.
// The separator is NUL, which cannot appear in a YAML scalar, so the flattened
// compare orders identically to Python's tuple compare.
func pointerKey(v yamlio.PyValue) string {
	m, _ := v.(yamlio.PyMap)
	target := m.Get("url")
	if yamlio.PyFalsy(target) {
		target = m.Get("id")
	}
	if yamlio.PyFalsy(target) {
		target = yamlio.PyStr("")
	}
	return strOrEmpty(m.Get("label")) + "\x00" +
		strOrEmpty(m.Get("system")) + "\x00" + yamlio.PyString(target)
}

// strOrEmpty is `str(p.get(k, ""))`: an absent key stringifies to "", a present
// null to "None".
func strOrEmpty(v yamlio.PyValue) string {
	if v == nil {
		return ""
	}
	return yamlio.PyString(v)
}

func containsValue(seq yamlio.PySeq, want yamlio.PyValue) bool {
	for _, v := range seq {
		if yamlio.PyEqual(v, want) {
			return true
		}
	}
	return false
}

// prdMeta is one `(prd_id, meta)` pair from _prd_meta_list (`:1404-1413`).
type prdMeta struct {
	ID   string
	Meta yamlio.PyMap
}

func prdMetaList(changeDir string) ([]prdMeta, error) {
	if !exists(changeDir) {
		return nil, nil
	}
	dirs, err := sortedChildren(changeDir)
	if err != nil {
		return nil, err
	}
	var out []prdMeta
	for _, d := range dirs {
		prd := filepath.Join(d, "prd.md")
		if !exists(prd) {
			continue
		}
		meta, _, err := ReadFrontmatter(prd)
		if err != nil {
			return nil, err
		}
		// `meta.get("id") or d.name` — a PRD that declares no id is keyed by
		// its directory, which is what makes the archive layout self-naming.
		id := filepath.Base(d)
		if v := meta.Get("id"); !yamlio.PyFalsy(v) {
			id = yamlio.PyString(v)
		}
		out = append(out, prdMeta{ID: id, Meta: meta})
	}
	return out, nil
}

// BuildFeatureIndex is build_feature_index (`:1434-1497`): a pure derivation of
// one platform's component -> artifact map, with no volatile field anywhere in
// it, so a fresh render of unchanged inputs is identical to the committed one.
//
// The `sorted(cids)` on the first line is load-bearing far beyond this
// function: it is the only reason FeatureIndexUnresolved's findings come out in
// a stable order (R-0.11). Nothing downstream sorts them.
func BuildFeatureIndex(ws *workspace.Workspace, pdir string) (yamlio.PyMap, error) {
	pid := filepath.Base(pdir)
	compDir := filepath.Join(pdir, "components")

	// `sorted(f.stem for f in comp_dir.glob("*.yaml"))` sorts STRINGS, not
	// paths — the stems have no separator in them, so the two orders coincide
	// here, but the mechanism is a plain string sort and stays one.
	var cids []string
	if isDir(compDir) {
		entries, err := os.ReadDir(compDir)
		if err != nil {
			return nil, model.Errorf(model.ExitArtifact, "cannot read %s: %v", compDir, err)
		}
		for _, e := range entries {
			if strings.HasSuffix(e.Name(), ".yaml") {
				cids = append(cids, strings.TrimSuffix(e.Name(), ".yaml"))
			}
		}
		sort.Strings(cids)
	}

	active, err := prdMetaList(filepath.Join(pdir, "change-records", "active"))
	if err != nil {
		return nil, err
	}
	archRoot := filepath.Join(pdir, "archive", "prds")
	archived, err := prdMetaList(archRoot)
	if err != nil {
		return nil, err
	}

	// Pending outcome reviews, keyed by the `prd:` edge they declare (R-2.2).
	// A completed review is omitted entirely rather than listed as done — the
	// index answers "what is still owed", not "what happened".
	outcomesByPRD := map[string]yamlio.PyMap{}
	if exists(archRoot) {
		dirs, err := sortedChildren(archRoot)
		if err != nil {
			return nil, err
		}
		for _, d := range dirs {
			o := filepath.Join(d, "outcome.md")
			if !exists(o) {
				continue
			}
			om, _, err := ReadFrontmatter(o)
			if err != nil {
				return nil, err
			}
			status, ok := om.Get("status").(yamlio.PyStr)
			if !ok || string(status) != "pending" {
				continue
			}
			key := filepath.Base(d)
			if v := om.Get("prd"); !yamlio.PyFalsy(v) {
				key = yamlio.PyString(v)
			}
			// `str(om.get("due")) if om.get("due") is not None else None` — a
			// date becomes its ISO string, which is why the committed index
			// quotes it: it is a str that would otherwise reload as a date.
			var due yamlio.PyValue = yamlio.PyNull{}
			if v := om.Get("due"); !yamlio.PyIsNone(v) {
				due = yamlio.PyStr(yamlio.PyString(v))
			}
			outcomesByPRD[key] = yamlio.PyMap{
				{K: "prd", V: yamlio.PyStr(key)},
				{K: "due", V: due},
				{K: "status", V: om.Get("status")},
			}
		}
	}

	components := yamlio.PyMap{}
	for _, cid := range cids {
		desc, err := loadMapping(filepath.Join(compDir, cid+".yaml"))
		if err != nil {
			return nil, err
		}
		entry := yamlio.PyMap{}
		var realityMeta yamlio.PyMap
		reality := filepath.Join(pdir, "reality", "components", cid+".md")
		if exists(reality) {
			entry = entry.Set("reality", yamlio.PyStr("reality/components/"+cid+".md"))
			if realityMeta, _, err = ReadFrontmatter(reality); err != nil {
				return nil, err
			}
		}

		var activeIDs, archivedIDs, discoveryIDs []string
		var outcomes yamlio.PySeq
		ptrMetas := []yamlio.PyMap{desc, realityMeta}

		collect := func(list []prdMeta, ids *[]string, archived bool) error {
			for _, pm := range list {
				in, err := pyContains(pm.Meta.Get("components"), cid)
				if err != nil {
					return err
				}
				if !in {
					continue
				}
				*ids = append(*ids, pm.ID)
				ptrMetas = append(ptrMetas, pm.Meta)
				if fd := pm.Meta.Get("fromDiscovery"); !yamlio.PyFalsy(fd) &&
					yamlio.PyString(fd) != "none" {
					discoveryIDs = append(discoveryIDs, yamlio.PyString(fd))
				}
				if archived {
					if o, ok := outcomesByPRD[pm.ID]; ok {
						outcomes = append(outcomes, o)
					}
				}
			}
			return nil
		}
		if err := collect(active, &activeIDs, false); err != nil {
			return nil, err
		}
		if err := collect(archived, &archivedIDs, true); err != nil {
			return nil, err
		}

		if len(activeIDs) > 0 {
			entry = entry.Set("activePrds", sortedUnique(activeIDs))
		}
		if len(archivedIDs) > 0 {
			entry = entry.Set("archivedPrds", sortedUnique(archivedIDs))
		}
		if len(discoveryIDs) > 0 {
			entry = entry.Set("discovery", sortedUnique(discoveryIDs))
		}
		if len(outcomes) > 0 {
			sort.SliceStable(outcomes, func(i, j int) bool {
				return yamlio.PyString(outcomes[i].(yamlio.PyMap).Get("prd")) <
					yamlio.PyString(outcomes[j].(yamlio.PyMap).Get("prd"))
			})
			entry = entry.Set("outcomes", outcomes)
		}
		if ptrs := CollectPointers(ptrMetas...); len(ptrs) > 0 {
			entry = entry.Set("externalPointers", ptrs)
		}
		components = components.Set(cid, entry)
	}
	return yamlio.PyMap{
		{K: "platform", V: yamlio.PyStr(pid)},
		{K: "components", V: components},
	}, nil
}

// pyContains is `cid in value`. A list membership test, but Python's `in` is
// also SUBSTRING containment on a str and key lookup on a dict, so a
// `components: customer-notification-service` written as a bare scalar matches
// every component id that is a substring of it. Same reasoning as pyIter.
func pyContains(v yamlio.PyValue, cid string) (bool, error) {
	if yamlio.PyFalsy(v) {
		return false, nil
	}
	switch t := v.(type) {
	case yamlio.PySeq:
		return containsValue(t, yamlio.PyStr(cid)), nil
	case yamlio.PyStr:
		return strings.Contains(string(t), cid), nil
	case yamlio.PyMap:
		for _, p := range t {
			if p.K == cid {
				return true, nil
			}
		}
		return false, nil
	}
	return false, model.Errorf(model.ExitArtifact,
		"'components:' must be a list, a mapping or a string")
}

// sortedUnique is `sorted(set(ids))`.
func sortedUnique(ids []string) yamlio.PySeq {
	seen := make(map[string]bool, len(ids))
	uniq := make([]string, 0, len(ids))
	for _, id := range ids {
		if !seen[id] {
			seen[id] = true
			uniq = append(uniq, id)
		}
	}
	sort.Strings(uniq)
	out := make(yamlio.PySeq, len(uniq))
	for i, id := range uniq {
		out[i] = yamlio.PyStr(id)
	}
	return out
}

// Unresolved is one dangling reference found by FeatureIndexUnresolved.
type Unresolved struct {
	// Kind is "discovery" or "prd".
	Kind string
	// ID is the reference that resolved to nothing.
	ID string
	// Component is the index entry that carried it.
	Component string
}

// FeatureIndexUnresolved is feature_index_unresolved (`:1513-1523`): discovery
// and PRD ids an index names that resolve to no document (R-2.7).
//
// The result is sorted by component — but nothing here sorts it. It comes out
// ordered because BuildFeatureIndex inserted the components in `sorted(cids)`
// order and this walks that insertion order (R-0.11). Iterating a Go map here
// would randomize the findings and break a committed golden.
func FeatureIndexUnresolved(ws *workspace.Workspace, idx yamlio.PyMap) []Unresolved {
	components, _ := idx.Get("components").(yamlio.PyMap)
	var bad []Unresolved
	for _, pair := range components {
		entry, _ := pair.V.(yamlio.PyMap)
		if list, ok := entry.Get("discovery").(yamlio.PySeq); ok {
			for _, did := range list {
				if !discoveryExists(ws, yamlio.PyString(did)) {
					bad = append(bad, Unresolved{"discovery", yamlio.PyString(did), pair.K})
				}
			}
		}
		if list, ok := entry.Get("outcomes").(yamlio.PySeq); ok {
			for _, o := range list {
				om, _ := o.(yamlio.PyMap)
				id := strOrEmpty(om.Get("prd"))
				if !prdExists(ws, id) {
					bad = append(bad, Unresolved{"prd", id, pair.K})
				}
			}
		}
	}
	return bad
}

func discoveryExists(ws *workspace.Workspace, did string) bool {
	for _, tdir := range ws.AllTeams() {
		if exists(filepath.Join(tdir, "product", "discovery", did, "brief.md")) {
			return true
		}
	}
	return false
}

func prdExists(ws *workspace.Workspace, prdID string) bool {
	for _, pdir := range ws.AllPlatforms() {
		if exists(filepath.Join(pdir, "change-records", "active", prdID, "prd.md")) {
			return true
		}
		if exists(filepath.Join(pdir, "archive", "prds", prdID, "prd.md")) {
			return true
		}
	}
	return false
}

// WriteFeatureIndexes is write_feature_indexes (`:1526-1534`), with the one
// sanctioned change this port makes to it (R-0.7a(a), R-0.7c).
//
// Python's guard is `out.read_text() != new` — a BYTE compare against a fresh
// render. It is safe there only because the bytes on disk came from the same
// PyYAML emitter. This port's emitter agrees with PyYAML on the documents it
// writes but not on every document it might READ, so a byte compare would
// rewrite each committed index on the first build; `graph build; graph build`
// would then never be a no-op and acceptance.sh §4's s0 == s1 would fail
// against Python-emitted bytes. The compare is therefore semantic — canonical
// YAML of the parsed structure — which is what gate 6 (`:1053`) already does to
// the very same file. Two spellings of one document now agree, as they should.
func WriteFeatureIndexes(ws *workspace.Workspace) ([]string, error) {
	var written []string
	for _, pdir := range ws.AllPlatforms() {
		idx, err := BuildFeatureIndex(ws, pdir)
		if err != nil {
			return nil, err
		}
		out := filepath.Join(pdir, "generated", "feature-index.yaml")
		fresh, err := yamlio.PyDumpCanonical(idx)
		if err != nil {
			return nil, model.Errorf(model.ExitArtifact, "cannot serialize %s: %v", out, err)
		}
		committed, err := loadMapping(out)
		if err != nil {
			return nil, err
		}
		current, err := yamlio.PyDumpCanonical(committed)
		if err != nil {
			return nil, model.Errorf(model.ExitArtifact, "cannot serialize %s: %v", out, err)
		}
		if current == fresh {
			continue
		}
		if err := yamlio.PyWriteCanonical(out, idx); err != nil {
			return nil, err
		}
		written = append(written, relTo(ws.Root, out))
	}
	return written, nil
}

// loadMapping is `load_yaml(path, {}) or {}`: the parsed mapping, or an empty
// one when the file is absent, empty, or holds a falsy document. A document
// that parses to a non-mapping exits 4 rather than reaching an attribute access
// that would raise (R-0.7a(j)).
func loadMapping(path string) (yamlio.PyMap, error) {
	v, err := yamlio.PyLoadFile(path)
	if err != nil {
		return nil, err
	}
	if yamlio.PyFalsy(v) {
		return yamlio.PyMap{}, nil
	}
	m, ok := v.(yamlio.PyMap)
	if !ok {
		return nil, model.Errorf(model.ExitArtifact,
			"%s: expected a mapping at the document root", path)
	}
	return m, nil
}

// sortedChildren is `sorted(dir.iterdir())` — every entry, files included,
// under PurePath order.
func sortedChildren(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, model.Errorf(model.ExitArtifact, "cannot read %s: %v", dir, err)
	}
	out := make([]string, 0, len(entries))
	for _, e := range entries {
		out = append(out, filepath.Join(dir, e.Name()))
	}
	yamlio.SortPaths(out)
	return out, nil
}
