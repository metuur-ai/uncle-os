// Package workspace resolves the workspace root and the canonical paths under
// it. It owns MANIFEST_NAME and LOCK_NAME: IsRoot needs the manifest name while
// internal/federation needs this package for path resolution, so putting those
// constants in federation would produce an import cycle.
//
// It is the port of the Python Workspace class (bin/company-os:211-263). Unlike
// require_root/platform_dir/team_dir there (:230/:238/:244), nothing here exits
// or writes to stdout — every failure is a *NotFoundError wrapping a
// *model.Error, so cmd/company-os classifies it by type, never by message.
package workspace

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/model"
)

// CanonicalRoots are the peer roots whose presence marks a workspace root
// (bin/company-os:184).
var CanonicalRoots = []string{
	"company-os", "platforms", "teams", "company-ontology", "knowledge",
}

const (
	// ManifestName is the hand-owned federation manifest. Its presence at the
	// root is the only switch into federated mode, and it marks a directory as a
	// workspace root even before the first sync.
	ManifestName = "workspace.yaml"
	// LockName is the machine-owned lock CI reads.
	LockName = "workspace.lock.yaml"
	// KnowledgeRoot is a node root but not a graph-docs root.
	KnowledgeRoot = "knowledge"
	// FederationCache holds git clones/caches; it is git-ignored.
	FederationCache = ".company-os/federation-cache"
	// RootEnv is consulted when --root is not given.
	RootEnv = "COMPANY_OS_WORKSPACE_ROOT"
)

// NotFoundKind names which workspace object a *NotFoundError is about.
type NotFoundKind string

const (
	// KindRoot is require_root's failure (bin/company-os:230).
	KindRoot NotFoundKind = "root"
	// KindPlatform is platform_dir's failure (bin/company-os:238).
	KindPlatform NotFoundKind = "platform"
	// KindTeam is team_dir's failure (bin/company-os:244).
	KindTeam NotFoundKind = "team"
)

// NotFoundError reports a workspace object that does not exist. All three sites
// are exit code 3 per .devlocal/go-port/exit-code-map.md, and the wrapped
// *model.Error carries that code, so model.CodeOf resolves it through
// errors.As without inspecting the message. Kind, ID and Dir let a caller (or
// the TUI) branch on which lookup failed for the same reason.
type NotFoundError struct {
	Kind NotFoundKind
	// ID is the requested platform or team id; empty for KindRoot.
	ID string
	// Dir is the directory searched — the root itself for KindRoot, the
	// platforms/ or teams/ directory otherwise.
	Dir string

	coded *model.Error
}

// Error returns the message byte-for-byte as Python's die() renders it, minus
// die()'s own "error: " prefix and trailing newline, which cmd/company-os adds.
func (e *NotFoundError) Error() string { return e.coded.Error() }

// Unwrap exposes the *model.Error so model.CodeOf finds the exit code.
func (e *NotFoundError) Unwrap() error { return e.coded }

func notFound(kind NotFoundKind, id, dir, format string, a ...any) *NotFoundError {
	err, _ := model.Errorf(model.ExitWorkspace, format, a...).(*model.Error)
	return &NotFoundError{Kind: kind, ID: id, Dir: dir, coded: err}
}

// Workspace is one workspace root plus the canonical directories under it. It is
// the port of Workspace.__init__ (bin/company-os:212-216).
type Workspace struct {
	Root      string
	Company   string
	Platforms string
	Teams     string
}

// Resolve applies the documented root resolution order:
// --root -> $COMPANY_OS_WORKSPACE_ROOT -> current directory (R-1.2).
// flagRoot is the raw --root value, or "" when the flag was not supplied.
func Resolve(flagRoot string) string {
	if flagRoot != "" {
		return flagRoot
	}
	if env := os.Getenv(RootEnv); env != "" {
		return env
	}
	return "."
}

// New builds a Workspace rooted at the given path, reproducing
// Path(root).resolve() at bin/company-os:213.
func New(root string) *Workspace {
	abs := resolvePath(root)
	return &Workspace{
		Root:      abs,
		Company:   filepath.Join(abs, "company-os"),
		Platforms: filepath.Join(abs, "platforms"),
		Teams:     filepath.Join(abs, "teams"),
	}
}

// resolvePath reproduces pathlib's non-strict Path.resolve(): make absolute,
// collapse "." and "..", then follow symlinks. Measured against CPython 3.13:
// a path whose tail does not exist still gets its longest existing ancestor
// resolved (`/tmp/no-such-dir` -> `/private/tmp/no-such-dir` on macOS), which
// plain filepath.EvalSymlinks cannot do — it fails outright on a missing path.
// Getting this wrong makes every error message that names the root diverge from
// Python's on any macOS temp dir.
func resolvePath(root string) string {
	abs, err := filepath.Abs(root)
	if err != nil {
		abs = filepath.Clean(root)
	}
	cur, tail := abs, ""
	for {
		if real, err := filepath.EvalSymlinks(cur); err == nil {
			if tail == "" {
				return real
			}
			return filepath.Join(real, tail)
		}
		parent := filepath.Dir(cur)
		if parent == cur {
			// Reached the filesystem root without finding anything that exists;
			// the lexically cleaned absolute path is the best available answer.
			return abs
		}
		tail = filepath.Join(filepath.Base(cur), tail)
		cur = parent
	}
}

// IsRoot reports whether this directory is a workspace root
// (bin/company-os:218-223). A federation manifest marks a root even before any
// canonical dir or slice exists.
func (w *Workspace) IsRoot() bool {
	if _, err := os.Stat(filepath.Join(w.Root, ManifestName)); err == nil {
		return true
	}
	for _, name := range CanonicalRoots {
		if fi, err := os.Stat(filepath.Join(w.Root, name)); err == nil && fi.IsDir() {
			return true
		}
	}
	return false
}

// RequireRoot fails fast when a workspace-scoped command runs outside a
// workspace root (bin/company-os:225-233, R-1.3). The message names the
// resolution order. Python die()s here; the exemption for `init` and
// `scratchpad` is the dispatch layer's (bin/company-os:2774-2776), not this
// package's.
func (w *Workspace) RequireRoot() error {
	if w.IsRoot() {
		return nil
	}
	return notFound(KindRoot, "", w.Root,
		"'%s' is not a workspace root: none of %s/ found here.\n"+
			"  workspace root resolution order: "+
			"--root -> $COMPANY_OS_WORKSPACE_ROOT -> current directory",
		w.Root, strings.Join(CanonicalRoots, "/, "))
}

// PlatformDir returns the catalog directory for a platform
// (bin/company-os:235-239). Python tests Path.exists(), not is_dir(), so a
// non-directory of the right name counts as found; that is reproduced here.
func (w *Workspace) PlatformDir(pid string) (string, error) {
	d := filepath.Join(w.Platforms, pid)
	if _, err := os.Stat(d); err != nil {
		return "", notFound(KindPlatform, pid, w.Platforms,
			"platform '%s' not found under %s", pid, w.Platforms)
	}
	return d, nil
}

// TeamDir returns the directory for a team (bin/company-os:241-245).
func (w *Workspace) TeamDir(tid string) (string, error) {
	d := filepath.Join(w.Teams, tid)
	if _, err := os.Stat(d); err != nil {
		return "", notFound(KindTeam, tid, w.Teams,
			"team '%s' not found under %s", tid, w.Teams)
	}
	return d, nil
}

// AllPlatforms lists every platform directory, sorted by name
// (bin/company-os:247-250). An absent platforms/ yields no entries, matching
// Python's early return.
func (w *Workspace) AllPlatforms() []string {
	return subdirs(w.Platforms)
}

// AllTeams lists every team directory, sorted by name (bin/company-os:252-255).
func (w *Workspace) AllTeams() []string {
	return subdirs(w.Teams)
}

// subdirs is sorted(dir.iterdir()) filtered by is_dir(). os.ReadDir already
// sorts by filename, which for a single parent is the order Python's sorted()
// over Path objects produces. Entries are stat'ed rather than read from the
// DirEntry so that, like Path.is_dir(), a symlink to a directory counts.
//
// Divergence, deliberate: where dir exists but is not a directory, Python raises
// NotADirectoryError and exits 1 through a traceback. R-2.10 forbids exiting
// from here and these two functions have no error return in Python, so an
// unreadable dir yields no entries.
func subdirs(dir string) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var out []string
	for _, e := range entries {
		p := filepath.Join(dir, e.Name())
		if fi, err := os.Stat(p); err == nil && fi.IsDir() {
			out = append(out, p)
		}
	}
	return out
}

// FindComponent searches every platform catalog for a component descriptor
// (bin/company-os:257-263), returning the owning platform's name and the
// descriptor's path. Python returns (platform_name, load_yaml(f)) and
// (None, None) when nothing matches; the load half is internal/yamlio's, so the
// seam here is the path. A miss is not an error — Python does not die().
func (w *Workspace) FindComponent(cid string) (platform, descriptor string, found bool) {
	for _, pdir := range w.AllPlatforms() {
		f := filepath.Join(pdir, "components", cid+".yaml")
		if _, err := os.Stat(f); err == nil {
			return filepath.Base(pdir), f, true
		}
	}
	return "", "", false
}
