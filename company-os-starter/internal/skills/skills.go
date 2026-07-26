package skills

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/frontmatter"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/workspace"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/yamlio"
)

// Suffix marks a shared, versioned skill file (bin/company-os:771). Personal
// rules under scratchpad/personal-rules/ are plain `.md` and do not carry it.
const Suffix = ".SKILL.md"

// Layer is a skill's origin (GPF-R-5.1). The four values are also the order the
// merged view lists them in, which is why Layers exists as a slice rather than
// being recovered from a map.
type Layer string

const (
	LayerCompany  Layer = "company"
	LayerPlatform Layer = "platform"
	LayerTeam     Layer = "team"
	LayerPersonal Layer = "personal"
)

// Layers is the display order of the origin-labeled section
// (bin/company-os:875).
var Layers = []Layer{LayerCompany, LayerPlatform, LayerTeam, LayerPersonal}

// Value is one frontmatter field as Python sees it after `meta.get(key)`.
//
// Three distinct questions are asked of `id`, `authority` and `extends`, and a
// bare Go string answers none of them faithfully:
//
//   - truthiness — `if s["extends"]:` (`:862`, `:887`, `:900`) is Python's
//     truthiness, so an empty string is absent-equivalent but the string "0" is
//     not;
//   - equality — `s["id"] == k["id"]` (`:852`) compares VALUES, so an id
//     resolved to the integer 5 does not shadow one resolved to the string "5";
//   - rendering — `id={s['id']}` (`:886`) is str(), which prints "None" for a
//     key that is absent or explicitly null. An absent key is therefore not an
//     empty string on the output path.
type Value struct {
	// Present reports that the key existed in the frontmatter mapping. It does
	// NOT distinguish an absent key from an explicit null: Python's .get()
	// returns None for both, and every downstream use treats them alike.
	Present bool
	// Text is str(value) — "None" when the value is absent or null.
	Text string
	// Truthy is Python's `if value:`.
	Truthy bool
	// kind is the Python type name. It exists only so that Equal can refuse to
	// call two values of different types equal when their str() forms collide.
	kind string
}

// NewValue builds the Value a STRING-typed frontmatter field produces. It is
// for callers that already hold a URI or an authority as a Go string; a value
// read from a file goes through the resolver instead, so that `version: 1.0`
// stays a float rather than becoming the string "1.0".
func NewValue(text string) Value {
	return Value{Present: true, Text: text, Truthy: text != "", kind: "str"}
}

// Equal is Python's `==` between two frontmatter values.
func (v Value) Equal(o Value) bool { return v.kind == o.kind && v.Text == o.Text }

// String is str(value).
func (v Value) String() string { return v.Text }

// Is reports whether the value is the given string, which is what
// `s["authority"] == "canonical"` (`:843`) asks.
func (v Value) Is(s string) bool { return v.kind == "str" && v.Text == s }

// Skill is one discovered skill file, labeled with its origin layer.
type Skill struct {
	// Path is the absolute path on disk.
	Path string
	// Rel is the workspace-relative path, which is the form every message uses
	// (`relative_to(ws.root)`, `:858`).
	Rel string
	// Layer is the origin layer (GPF-R-5.1).
	Layer Layer
	// Platform is set on the platform layer only; Team on the team and personal
	// layers only. Both empty on the company layer.
	Platform string
	Team     string
	// Name is the shadowing identity: the file stem with `.SKILL` stripped.
	Name string
	// Body is everything after the frontmatter fence.
	Body string

	ID        Value
	Authority Value
	Extends   Value
}

// Scope is `s["platform"] or s["team"]` (`:881`, `:896`) — the platform when
// there is one, else the team, else empty.
func (s Skill) Scope() string {
	if s.Platform != "" {
		return s.Platform
	}
	return s.Team
}

// IsCanonical is `s["authority"] == "canonical"`.
func (s Skill) IsCanonical() bool { return s.Authority.Is("canonical") }

// Steps returns each numbered, tier-tagged step's head line in authored order
// (parse_skill_steps, `:820-823`). Only the head line is kept: it is the line
// that carries the `(mandatory|default|guidance)` marker.
func Steps(body string) []string {
	var out []string
	for _, line := range pySplitLines(body) {
		if isStep(line) {
			out = append(out, pyStrip(line))
		}
	}
	return out
}

// Name derives a skill's identity from its path (_skill_name, `:775-782`).
// The `.SKILL.md` marker is stripped for a shared skill; a personal rule keeps
// the plain stem, so `maria.md` is `maria` but `a.b.md` is `a.b`.
func Name(path string) string {
	n := filepath.Base(path)
	if strings.HasSuffix(n, Suffix) {
		return n[:len(n)-len(Suffix)]
	}
	return strings.TrimSuffix(n, filepath.Ext(n))
}

// Discover merges the four skill layers, each labeled with its origin
// (discover_skills, `:794-817`, GPF-R-5.1).
//
// Absence-tolerant: a missing layer directory is skipped, so a standalone team
// with no skills yields an empty but valid result rather than an error.
func Discover(ws *workspace.Workspace) ([]Skill, error) {
	var found []Skill

	if err := appendLayer(&found, ws, filepath.Join(ws.Company, "skills"),
		"*"+Suffix, LayerCompany, "", ""); err != nil {
		return nil, err
	}
	for _, pdir := range ws.AllPlatforms() {
		if err := appendLayer(&found, ws, filepath.Join(pdir, "skills"),
			"*"+Suffix, LayerPlatform, filepath.Base(pdir), ""); err != nil {
			return nil, err
		}
	}
	for _, tdir := range ws.AllTeams() {
		team := filepath.Base(tdir)
		if err := appendLayer(&found, ws, filepath.Join(tdir, "skills"),
			"*"+Suffix, LayerTeam, "", team); err != nil {
			return nil, err
		}
		// Personal rules are git-ignored, so this layer appears and disappears
		// with the local scratchpad. Only the merged view counts it; gate 7
		// deliberately leaves it out of its totals (`:1081-1084`).
		if err := appendLayer(&found, ws, filepath.Join(tdir, "scratchpad", "personal-rules"),
			"*.md", LayerPersonal, "", team); err != nil {
			return nil, err
		}
	}
	return found, nil
}

// appendLayer is one `if dir.is_dir(): for p in sorted(dir.glob(pat))` block.
func appendLayer(out *[]Skill, ws *workspace.Workspace, dir, pattern string,
	layer Layer, platform, team string) error {
	paths, err := globSorted(dir, pattern)
	if err != nil {
		return err
	}
	for _, p := range paths {
		s, err := read(ws, p, layer, platform, team)
		if err != nil {
			return err
		}
		*out = append(*out, s)
	}
	return nil
}

// globSorted is `sorted(dir.glob(pattern))`, and nothing when dir is not a
// directory. os.ReadDir already returns entries sorted by name, which for a
// single parent is the order sorted() over Path objects produces. Hidden files
// are included: pathlib's glob, unlike glob.glob, does not special-case a
// leading dot, and neither does filepath.Match.
func globSorted(dir, pattern string) ([]string, error) {
	fi, err := os.Stat(dir)
	if err != nil || !fi.IsDir() {
		return nil, nil
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", dir, err)
	}
	var out []string
	for _, e := range entries {
		ok, err := filepath.Match(pattern, e.Name())
		if err != nil {
			return nil, fmt.Errorf("matching %s: %w", pattern, err)
		}
		if ok {
			out = append(out, filepath.Join(dir, e.Name()))
		}
	}
	return out, nil
}

// read is _read_skill (`:785-791`).
func read(ws *workspace.Workspace, path string, layer Layer, platform, team string) (Skill, error) {
	doc, err := frontmatter.ParseFile(path)
	if err != nil {
		return Skill{}, err
	}
	s := Skill{
		Path:     path,
		Rel:      relative(ws, path),
		Layer:    layer,
		Platform: platform,
		Team:     team,
		Name:     Name(path),
		Body:     string(doc.Body),
	}
	meta, err := mapping(path, doc.YAML)
	if err != nil {
		return Skill{}, err
	}
	for _, f := range []struct {
		key string
		dst *Value
	}{{"id", &s.ID}, {"authority", &s.Authority}, {"extends", &s.Extends}} {
		v, err := value(path, f.key, meta)
		if err != nil {
			return Skill{}, err
		}
		*f.dst = v
	}
	return s, nil
}

// mapping completes `yaml.safe_load(block) or {}` and then the `meta or {}` at
// `:787`, returning the mapping node the three .get() calls read, or nil for an
// empty one.
//
// A frontmatter block that loads to a truthy NON-mapping (a list, a bare
// scalar) makes Python raise AttributeError on the first .get() and exit 1
// through a traceback; the error returned here carries no explicit code, which
// model.CodeOf also resolves to 1.
func mapping(path string, block []byte) (*yaml.Node, error) {
	if block == nil {
		return nil, nil
	}
	doc, err := yamlio.LoadFrontmatter(block)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	if doc.IsFalsy() {
		return nil, nil
	}
	root := yamlio.Deref(doc.Root())
	if root == nil || root.Kind != yaml.MappingNode {
		return nil, fmt.Errorf("%s: frontmatter is not a mapping", path)
	}
	return root, nil
}

// value resolves one `meta.get(key)`.
//
// Deliberate, recorded divergence: a key whose value is a list or a mapping is
// refused here where Python would render its repr into the merged view. Nothing
// in the ontology admits a collection-valued `id`, `authority` or `extends`, and
// a collection-valued `extends` fails validate under Python anyway (its str()
// cannot match the platform-skill:// pattern, so it reports as dangling). A
// visible refusal is preferred to a hand-rolled approximation of Python's repr.
func value(path, key string, meta *yaml.Node) (Value, error) {
	absent := Value{Text: "None", kind: "NoneType"}
	if meta == nil {
		return absent, nil
	}
	n := yamlio.Deref(yamlio.MapGet(meta, key))
	if n == nil {
		return absent, nil
	}
	if n.Kind != yaml.ScalarNode {
		return Value{}, fmt.Errorf("%s: frontmatter '%s' must be a scalar", path, key)
	}
	sc, err := yamlio.Resolve(n)
	if err != nil {
		return Value{}, fmt.Errorf("%s: frontmatter '%s': %w", path, key, err)
	}
	v := Value{Present: true, Text: sc.String()}
	switch sc.Kind {
	case yamlio.KindNull:
		v.kind, v.Truthy = "NoneType", false
	case yamlio.KindBool:
		v.kind, v.Truthy = "bool", sc.Bool
	case yamlio.KindInt:
		v.kind, v.Truthy = "int", sc.Int != nil && sc.Int.Sign() != 0
	case yamlio.KindFloat:
		v.kind, v.Truthy = "float", sc.Float != 0
	case yamlio.KindTimestamp:
		// datetime.date and datetime.datetime are always truthy.
		v.kind, v.Truthy = "datetime", true
	default:
		v.kind, v.Truthy = "str", sc.Raw != ""
	}
	return v, nil
}

// extendsRE is `^platform-skill://([^/]+)/(.+)$` (`:830`) transcribed for Go.
// `\A` spells out re.match's implicit start anchor, and `\n?\z` spells out
// Python's `$`, which also matches immediately before a single trailing
// newline. `.` excludes newlines in both engines; `[^/]` includes them in both.
var extendsRE = regexp.MustCompile(`\Aplatform-skill://([^/]+)/(.+)\n?\z`)

// ResolveExtends resolves `platform-skill://<p>/<name>` to the platform-LAYER
// skill file platforms/<p>/skills/<name>.SKILL.md (resolve_extends, `:826-834`,
// GPF-R-5.3), or reports found=false when the target does not exist.
//
// The URI addresses a FILE, not an id: an existing skill's own
// `id: skill://<scope>/<name>` names a semantic namespace unrelated to the
// platform directory (`:764-769`).
//
// A malformed URI and an absent file are the same answer — Python returns None
// for both — so a caller cannot and need not tell them apart.
func ResolveExtends(ws *workspace.Workspace, uri Value) (Skill, bool, error) {
	if !uri.Truthy {
		// `str(uri or "")`: a falsy URI is resolved against the empty string,
		// which never matches.
		return Skill{}, false, nil
	}
	m := extendsRE.FindStringSubmatch(uri.Text)
	if m == nil {
		return Skill{}, false, nil
	}
	// Joined without filepath.Clean so that a `..` segment keeps the same
	// existence semantics as Python's Path concatenation, which does not
	// normalize either.
	sep := string(os.PathSeparator)
	path := ws.Platforms + sep + m[1] + sep + "skills" + sep + m[2] + Suffix
	if _, err := os.Stat(path); err != nil {
		return Skill{}, false, nil
	}
	s, err := read(ws, path, LayerPlatform, m[1], "")
	if err != nil {
		return Skill{}, false, err
	}
	return s, true, nil
}

// relative is `path.relative_to(ws.root)` with forward slashes, which is what
// every rendered message shows.
func relative(ws *workspace.Workspace, path string) string {
	rel, err := filepath.Rel(ws.Root, path)
	if err != nil {
		return path
	}
	return filepath.ToSlash(rel)
}
