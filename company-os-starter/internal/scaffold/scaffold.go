package scaffold

// The scaffolding engine — the port of bin/company-os:1783-2060.
//
// One engine writes the SOURCE artifacts for every unit (company, ontology,
// platform, team, component); the GENERATED artifacts (CLAUDE.md nodes,
// feature-index, derived tags) are then produced by the same code path
// `graph build` uses, so a fresh scaffold validates green (GPF-R-1.7). That
// second half is internal/graph's; see the Rebuild seam below.
//
// Every writer refuses to clobber an existing file (GPF-R-1.9), and every
// refusal is exit code 8 per .devlocal/go-port/exit-code-map.md § A.

import (
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/model"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/workspace"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/yamlio"
)

// The PyYAML-safe_dump-compatible emitter lives in internal/yamlio, next to the
// loader, because internal/governance and internal/federation need it too and
// must not reach it through a command package. These aliases keep the scaffold
// dicts below reading as the Python literals they transcribe.
type (
	pyStr = yamlio.PyStr
	pySeq = yamlio.PySeq
	pyMap = yamlio.PyMap
)

var (
	pyDump      = yamlio.PyDump
	loadPy      = yamlio.PyLoadFile
	writePyYAML = yamlio.PyWriteFile
)

// Rebuild re-derives tags and generated aggregates for a workspace and reports
// the lines it printed — the port of rebuild_generated (bin/company-os:1803).
//
// It is a seam, not an implementation. The LLD puts rebuild_generated in
// internal/graph with a one-way scaffold -> graph dependency, because it also
// backs `graph build` and reaches iter_graph_docs / rewrite_frontmatter_tags /
// write_feature_indexes / write_claude_nodes; importing graph from here would
// be that dependency, but taking it as a parameter keeps this package testable
// without one and lets task 2.3 land independently.
//
// It returns lines rather than writing them because the Python order is
// load-bearing: rebuild_generated's output ("  wrote index …", "  node …")
// precedes the command's own "added platform 'x'" line, and only cmd/company-os
// may write to stdout.
//
// internal/graph must provide a function of exactly this shape, emitting one
// line per changed artifact in the same order Python does. A nil Rebuild is a
// no-op, which is what every caller passes until task 2.3 lands.
type Rebuild func(ws *workspace.Workspace) (lines []string, err error)

// Slugify is bin/company-os:72-73: lowercase, then collapse every run of
// characters outside [a-z0-9] to a single "-", then strip leading and trailing
// "-".
//
// R-1.13: Python's str.lower() and Go's strings.ToLower disagree on a handful of
// code points (U+0130 LATIN CAPITAL LETTER I WITH DOT ABOVE lowercases to two
// code points in Python and one in Go). The disagreement cannot reach the
// result: every code point either survives as ASCII [a-z0-9] — where the two
// agree — or is replaced by "-" regardless of how many of them there were.
func Slugify(title string) string {
	var b strings.Builder
	dash := false
	for _, r := range strings.ToLower(title) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			dash = false
			continue
		}
		if !dash {
			b.WriteByte('-')
			dash = true
		}
	}
	return strings.Trim(b.String(), "-")
}

// titleCase is Python's str.title(): a character is uppercased when the
// preceding character is not cased, and lowercased otherwise. It is applied to
// slug-derived ids, so "my-platform" -> "my platform" -> "My Platform".
func titleCase(s string) string {
	var b strings.Builder
	prevCased := false
	for _, r := range s {
		if prevCased {
			b.WriteRune(unicode.ToLower(r))
		} else {
			b.WriteRune(unicode.ToUpper(r))
		}
		prevCased = unicode.IsUpper(r) || unicode.IsLower(r) || unicode.IsTitle(r)
	}
	return b.String()
}

// ConflictError reports a destination that is already occupied — _write_new's
// refusal (bin/company-os:1793-1797) and the two command-level refusals at
// :1971 and :2037. All three are exit code 8: the invocation was well-formed
// and nothing was mutated.
type ConflictError struct {
	// Path is the artifact that already exists, rendered exactly as the oracle
	// renders it (absolute for _write_new, workspace-relative for `reality new`).
	Path  string
	coded *model.Error
}

func (e *ConflictError) Error() string { return e.coded.Error() }
func (e *ConflictError) Unwrap() error { return e.coded }

func conflict(path, format string, a ...any) *ConflictError {
	err, _ := model.Errorf(model.ExitConflict, format, a...).(*model.Error)
	return &ConflictError{Path: path, coded: err}
}

// writeNew is _write_new (bin/company-os:1793-1801): write a source file,
// refusing to overwrite anything existing.
func writeNew(path, text string) error {
	// Path.exists() follows symlinks, so a symlink to an existing file is a
	// conflict and a DANGLING one is not — os.Stat reproduces both, os.Lstat
	// would get the dangling case backwards.
	if _, err := os.Stat(path); err == nil {
		return conflict(path, "refusing to overwrite existing file: %s", path)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o777); err != nil {
		return model.Errorf(model.ExitArtifact, "cannot create %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(text), 0o666); err != nil {
		return model.Errorf(model.ExitArtifact, "cannot write %s: %v", path, err)
	}
	return nil
}

// writeNewYAML is _write_new(path, canonical_or_block(data)) — the stable block
// form every YAML source artifact is scaffolded in (bin/company-os:1886-1890).
func writeNewYAML(path string, data pyMap) error {
	text, err := pyDump(data)
	if err != nil {
		return model.Errorf(model.ExitArtifact, "cannot serialize %s: %v", path, err)
	}
	return writeNew(path, text)
}

func strs(items ...string) pySeq {
	out := make(pySeq, 0, len(items))
	for _, s := range items {
		out = append(out, pyStr(s))
	}
	return out
}

// scaffoldCompany is bin/company-os:1828-1843.
func scaffoldCompany(root, company string) error {
	slug := Slugify(company)
	err := writeNewYAML(filepath.Join(root, "company-os", "company.yaml"), pyMap{
		{K: "schemaVersion", V: pyStr("1.0")},
		{K: "kind", V: pyStr("Company")},
		{K: "metadata", V: pyMap{{K: "id", V: pyStr(slug)}, {K: "name", V: pyStr(company)}}},
		{K: "tags", V: strs("kind/company")},
	})
	if err != nil {
		return err
	}
	return writeNewYAML(
		filepath.Join(root, "company-os", "standards", "company-baseline.yaml"), pyMap{
			{K: "schemaVersion", V: pyStr("1.0")},
			{K: "kind", V: pyStr("CompanyBaseline")},
			{K: "controls", V: pySeq{pyMap{
				{K: "id", V: pyStr("security-baseline")},
				{K: "version", V: pyStr("1.0")},
				{K: "level", V: pyStr("default")},
				{K: "requirement", V: pyStr("Services authenticate inbound calls and " +
					"encrypt data in transit.")},
			}}},
			{K: "tags", V: strs("kind/baseline", "scope/company", "tier/mixed")},
		})
}

// scaffoldPlatform is bin/company-os:1846-1858.
func scaffoldPlatform(root, pid string) error {
	err := writeNewYAML(filepath.Join(root, "platforms", pid, "platform.yaml"), pyMap{
		{K: "schemaVersion", V: pyStr("1.0")},
		{K: "kind", V: pyStr("Platform")},
		{K: "metadata", V: pyMap{
			{K: "id", V: pyStr(pid)},
			{K: "name", V: pyStr(titleCase(strings.ReplaceAll(pid, "-", " ")))},
		}},
		{K: "conformance", V: pyMap{
			{K: "companyOsVersion", V: pyStr("2026.2")},
			{K: "profile", V: pyStr("standard")},
		}},
		{K: "tags", V: strs("kind/platform", "platform/"+pid)},
	})
	if err != nil {
		return err
	}
	return writeNewYAML(
		filepath.Join(root, "platforms", pid, "governance", "requirements.yaml"), pyMap{
			{K: "schemaVersion", V: pyStr("1.0")},
			{K: "kind", V: pyStr("PlatformRequirements")},
			{K: "platform", V: pyMap{{K: "id", V: pyStr(pid)}}},
			{K: "requirements", V: pySeq{}},
			{K: "tags", V: strs("kind/requirements", "platform/"+pid)},
		})
}

// scaffoldTeam is bin/company-os:1861-1876.
func scaffoldTeam(root, tid string) error {
	tname := titleCase(strings.ReplaceAll(tid, "-", " "))
	err := writeNewYAML(filepath.Join(root, "teams", tid, "team.yaml"), pyMap{
		{K: "schemaVersion", V: pyStr("1.0")},
		{K: "kind", V: pyStr("Team")},
		{K: "metadata", V: pyMap{{K: "id", V: pyStr(tid)}, {K: "name", V: pyStr(tname)}}},
		{K: "agentSkills", V: pyMap{
			{K: "canonicalPath", V: pyStr("skills/")},
			{K: "personalPath", V: pyStr("scratchpad/personal-rules/")},
			{K: "precedence", V: pyStr("canonical-mandatory > personal > canonical-default " +
				"> canonical-guidance")},
			{K: "onConflict", V: pyStr("prefer-canonical-and-inform-user")},
		}},
		{K: "tags", V: strs("kind/team", "team/"+tid)},
	})
	if err != nil {
		return err
	}
	if err := writeNew(
		filepath.Join(root, "teams", tid, "standards", "definition-of-ready.md"),
		strings.ReplaceAll(dorTemplate, "{tid}", tid)); err != nil {
		return err
	}
	if err := writeNew(
		filepath.Join(root, "teams", tid, "standards", "definition-of-done.md"),
		strings.ReplaceAll(dodTemplate, "{tid}", tid)); err != nil {
		return err
	}
	onboarding := strings.ReplaceAll(onboardingTemplate, "{tid}", tid)
	onboarding = strings.ReplaceAll(onboarding, "{tname}", tname)
	return writeNew(filepath.Join(root, "teams", tid, "onboarding", "developer.md"), onboarding)
}

// scaffoldComponent is bin/company-os:1879-1891.
func scaffoldComponent(root, pid, cid string) error {
	return writeNewYAML(
		filepath.Join(root, "platforms", pid, "components", cid+".yaml"), pyMap{
			{K: "schemaVersion", V: pyStr("1.0")},
			{K: "kind", V: pyStr("Component")},
			{K: "metadata", V: pyMap{
				{K: "id", V: pyStr(cid)},
				{K: "name", V: pyStr(titleCase(strings.ReplaceAll(cid, "-", " ")))},
				{K: "type", V: pyStr("service")},
				{K: "status", V: pyStr("active")},
			}},
			{K: "ownership", V: pyMap{
				{K: "accountableTeam", V: pyStr("team://TODO")},
				{K: "repository", V: pyStr("repo://" + cid)},
			}},
			{K: "platformRelationships", V: pySeq{pyMap{
				{K: "platform", V: pyStr("platform://" + pid)},
				{K: "relationship", V: pyStr("belongs-to")},
			}}},
			{K: "tags", V: strs("component/"+cid, "kind/component", "platform/"+pid)},
		})
}

// registerID is bin/company-os:1815-1825: append a canonical ID to
// company-ontology/ids/registry.yaml, idempotent per id. The whole file is
// re-dumped through safe_dump, which is why writePyYAML has to be
// byte-compatible with it rather than merely well-formed.
//
// "Idempotent per id" is stronger here than in the oracle: Python re-dumps the
// file even when the id was already registered, and this returns without
// touching it (R-0.7c). See the branch below for why that is a semantic compare
// rather than a shortcut.
func registerID(root, canonicalID, definedIn string) error {
	path := filepath.Join(root, "company-ontology", "ids", "registry.yaml")
	loaded, err := loadPy(path)
	if err != nil {
		return err
	}
	// `load_yaml(path, None) or {default}` — Python TRUTHINESS, not a nil check.
	// A registry holding `[]`, `{}`, `0` or `''` falls back to the default and
	// Python exits 0 having written a valid registry over it (R-1.7a); guarding
	// only on nil made `printf '[]' > registry.yaml; add platform zzz` fail.
	var data pyMap
	if yamlio.PyFalsy(loaded) {
		data = pyMap{
			{K: "schemaVersion", V: pyStr("1.0")},
			{K: "kind", V: pyStr("IdRegistry")},
			{K: "ids", V: pySeq{}},
			{K: "tags", V: strs("ontology/registry")},
		}
	} else {
		m, ok := loaded.(pyMap)
		if !ok {
			// `data.setdefault(...)` on a non-dict raises AttributeError and the
			// file is never rewritten. R-0.7a(j): same outcome, exit 4 with a
			// diagnostic instead of a traceback, and nothing written.
			return model.Errorf(model.ExitArtifact,
				"%s: expected a mapping at the document root", path)
		}
		data = m
	}

	// `data.setdefault("ids", [])` — present-but-absent are the only two shapes
	// Python survives; anything else raises there.
	var ids pySeq
	switch existing := data.Get("ids").(type) {
	case nil:
		data = data.Set("ids", pySeq{})
	case pySeq:
		ids = existing
	default:
		return model.Errorf(model.ExitArtifact, "%s: 'ids:' must be a list", path)
	}

	// `any(e.get("id") == canonical_id for e in ids)` — a generator, so it stops
	// at the first match and only reaches a malformed entry that precedes none.
	// `.get` on a non-dict raises AttributeError; skipping the entry instead let
	// a registry whose `ids:` held a bare string be silently rewritten and
	// reformatted where Python writes nothing (R-0.7a(j)).
	for _, e := range ids {
		m, ok := e.(pyMap)
		if !ok {
			return model.Errorf(model.ExitArtifact,
				"%s: every 'ids:' entry must be a mapping", path)
		}
		if id, ok := m.Get("id").(pyStr); ok && string(id) == canonicalID {
			// R-0.7c: skip the write, because the parsed structure is
			// unchanged. This is the ONE branch where that is true — `data` is
			// `loaded` itself here (no default was substituted and no
			// setdefault fired, or the loop could not be running), so a
			// semantic compare against the file on disk can only say "equal".
			//
			// Python rewrites anyway, and that rewrite is not a no-op even
			// under Python: safe_dump reflows the committed registry's
			// flow-style entries into block style. Under yaml.v3 it would
			// reflow them differently again, dirtying a tree that nothing
			// asked to change and breaking R-0.10.
			return nil
		}
	}
	data = data.Set("ids", append(ids, pyMap{
		{K: "id", V: pyStr(canonicalID)},
		{K: "definedIn", V: pyStr(definedIn)},
	}))
	return writePyYAML(path, data)
}
