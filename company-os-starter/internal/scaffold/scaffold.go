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
		{"schemaVersion", pyStr("1.0")},
		{"kind", pyStr("Company")},
		{"metadata", pyMap{{"id", pyStr(slug)}, {"name", pyStr(company)}}},
		{"tags", strs("kind/company")},
	})
	if err != nil {
		return err
	}
	return writeNewYAML(
		filepath.Join(root, "company-os", "standards", "company-baseline.yaml"), pyMap{
			{"schemaVersion", pyStr("1.0")},
			{"kind", pyStr("CompanyBaseline")},
			{"controls", pySeq{pyMap{
				{"id", pyStr("security-baseline")},
				{"version", pyStr("1.0")},
				{"level", pyStr("default")},
				{"requirement", pyStr("Services authenticate inbound calls and " +
					"encrypt data in transit.")},
			}}},
			{"tags", strs("kind/baseline", "scope/company", "tier/mixed")},
		})
}

// scaffoldPlatform is bin/company-os:1846-1858.
func scaffoldPlatform(root, pid string) error {
	err := writeNewYAML(filepath.Join(root, "platforms", pid, "platform.yaml"), pyMap{
		{"schemaVersion", pyStr("1.0")},
		{"kind", pyStr("Platform")},
		{"metadata", pyMap{
			{"id", pyStr(pid)},
			{"name", pyStr(titleCase(strings.ReplaceAll(pid, "-", " ")))},
		}},
		{"conformance", pyMap{
			{"companyOsVersion", pyStr("2026.2")},
			{"profile", pyStr("standard")},
		}},
		{"tags", strs("kind/platform", "platform/"+pid)},
	})
	if err != nil {
		return err
	}
	return writeNewYAML(
		filepath.Join(root, "platforms", pid, "governance", "requirements.yaml"), pyMap{
			{"schemaVersion", pyStr("1.0")},
			{"kind", pyStr("PlatformRequirements")},
			{"platform", pyMap{{"id", pyStr(pid)}}},
			{"requirements", pySeq{}},
			{"tags", strs("kind/requirements", "platform/"+pid)},
		})
}

// scaffoldTeam is bin/company-os:1861-1876.
func scaffoldTeam(root, tid string) error {
	tname := titleCase(strings.ReplaceAll(tid, "-", " "))
	err := writeNewYAML(filepath.Join(root, "teams", tid, "team.yaml"), pyMap{
		{"schemaVersion", pyStr("1.0")},
		{"kind", pyStr("Team")},
		{"metadata", pyMap{{"id", pyStr(tid)}, {"name", pyStr(tname)}}},
		{"agentSkills", pyMap{
			{"canonicalPath", pyStr("skills/")},
			{"personalPath", pyStr("scratchpad/personal-rules/")},
			{"precedence", pyStr("canonical-mandatory > personal > canonical-default " +
				"> canonical-guidance")},
			{"onConflict", pyStr("prefer-canonical-and-inform-user")},
		}},
		{"tags", strs("kind/team", "team/"+tid)},
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
			{"schemaVersion", pyStr("1.0")},
			{"kind", pyStr("Component")},
			{"metadata", pyMap{
				{"id", pyStr(cid)},
				{"name", pyStr(titleCase(strings.ReplaceAll(cid, "-", " ")))},
				{"type", pyStr("service")},
				{"status", pyStr("active")},
			}},
			{"ownership", pyMap{
				{"accountableTeam", pyStr("team://TODO")},
				{"repository", pyStr("repo://" + cid)},
			}},
			{"platformRelationships", pySeq{pyMap{
				{"platform", pyStr("platform://" + pid)},
				{"relationship", pyStr("belongs-to")},
			}}},
			{"tags", strs("component/"+cid, "kind/component", "platform/"+pid)},
		})
}

// registerID is bin/company-os:1815-1825: append a canonical ID to
// company-ontology/ids/registry.yaml, idempotent per id. The whole file is
// re-dumped through safe_dump, which is why writePyYAML has to be
// byte-compatible with it rather than merely well-formed.
func registerID(root, canonicalID, definedIn string) error {
	path := filepath.Join(root, "company-ontology", "ids", "registry.yaml")
	loaded, err := loadPy(path)
	if err != nil {
		return err
	}
	data, ok := loaded.(pyMap)
	if !ok || len(data) == 0 {
		// `load_yaml(path, None) or {default}` — Python truthiness, so an empty
		// mapping falls back to the default too (R-1.7a).
		if loaded != nil && !ok {
			return model.Errorf(model.ExitArtifact,
				"%s: expected a mapping at the document root", path)
		}
		data = pyMap{
			{"schemaVersion", pyStr("1.0")},
			{"kind", pyStr("IdRegistry")},
			{"ids", pySeq{}},
			{"tags", strs("ontology/registry")},
		}
	}

	// `data.setdefault("ids", [])` — present-but-absent are the only two shapes
	// Python survives; anything else raises there, so it is an artifact fault
	// here rather than a silent divergence.
	var ids pySeq
	switch existing := data.get("ids").(type) {
	case nil:
		data = data.set("ids", pySeq{})
	case pySeq:
		ids = existing
	default:
		return model.Errorf(model.ExitArtifact, "%s: 'ids:' must be a list", path)
	}

	for _, e := range ids {
		if m, ok := e.(pyMap); ok {
			if id, ok := m.get("id").(pyStr); ok && string(id) == canonicalID {
				return writePyYAML(path, data)
			}
		}
	}
	data = data.set("ids", append(ids, pyMap{
		{"id", pyStr(canonicalID)},
		{"definedIn", pyStr(definedIn)},
	}))
	return writePyYAML(path, data)
}
