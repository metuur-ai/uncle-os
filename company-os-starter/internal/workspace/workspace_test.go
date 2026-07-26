package workspace

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/model"
)

// The three message templates below were captured by running the Python CLI,
// not read off its source. Transcript (PYTHONPATH=company-os-starter/vendor):
//
//	$ cd <empty dir>; company-os validate
//	error: '<root>' is not a workspace root: none of company-os/, platforms/,
//	teams/, company-ontology/, knowledge/ found here.
//	  workspace root resolution order: --root -> $COMPANY_OS_WORKSPACE_ROOT -> current directory
//
//	$ cd examples/workspace; company-os prd validate --platform ghost some-prd
//	error: platform 'ghost' not found under <root>/platforms
//
//	$ cd examples/workspace; company-os governance resolve --team ghost
//	error: team 'ghost' not found under <root>/teams
//
// die() (bin/company-os:41-43) adds the "error: " prefix and the trailing
// newline; everything after it is what these errors must carry byte-for-byte.
const (
	wantRootMsg = "'%s' is not a workspace root: none of company-os/, platforms/, " +
		"teams/, company-ontology/, knowledge/ found here.\n" +
		"  workspace root resolution order: " +
		"--root -> $COMPANY_OS_WORKSPACE_ROOT -> current directory"
	wantPlatformMsg = "platform 'ghost' not found under %s"
	wantTeamMsg     = "team 'ghost' not found under %s"
)

// makeRoot builds a workspace root under t.TempDir() by creating company-os/.
func makeRoot(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	for _, sub := range []string{"company-os", "platforms", "teams"} {
		if err := os.MkdirAll(filepath.Join(dir, sub), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

// --------------------------------------------------------------- R-1.2: root

func TestResolvePrecedence(t *testing.T) {
	t.Setenv(RootEnv, "/from/env")
	if got := Resolve("/from/flag"); got != "/from/flag" {
		t.Fatalf("--root should win over $%s: got %q", RootEnv, got)
	}
	if got := Resolve(""); got != "/from/env" {
		t.Fatalf("$%s should win over cwd: got %q", RootEnv, got)
	}
	t.Setenv(RootEnv, "")
	if got := Resolve(""); got != "." {
		t.Fatalf("cwd is the last resort: got %q", got)
	}
}

// TestNewResolvesSymlinks is the macOS temp-dir case: Path(root).resolve()
// follows symlinks, so on darwin every /var/folders path Python reports is
// really /private/var/folders. A root reached through a symlink must resolve to
// its target or every error message naming the root diverges.
func TestNewResolvesSymlinks(t *testing.T) {
	real := makeRoot(t)
	link := filepath.Join(t.TempDir(), "link")
	if err := os.Symlink(real, link); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	ws := New(link)
	wantRoot, err := filepath.EvalSymlinks(real)
	if err != nil {
		t.Fatal(err)
	}
	if ws.Root != wantRoot {
		t.Fatalf("Root = %q, want the symlink target %q", ws.Root, wantRoot)
	}
	if !ws.IsRoot() {
		t.Fatal("IsRoot() = false through a symlinked root")
	}
}

// TestNewResolvesMissingTailUnderSymlink pins the half of Path.resolve() that
// filepath.EvalSymlinks alone cannot do. Measured against CPython 3.13 on
// darwin: Path('/tmp/no-such-dir-xyz').resolve() == '/private/tmp/no-such-dir-xyz'.
// EvalSymlinks fails on any missing path, so a naive Abs fallback would leave
// the symlinked prefix unresolved.
func TestNewResolvesMissingTailUnderSymlink(t *testing.T) {
	real := t.TempDir()
	linkDir := t.TempDir()
	link := filepath.Join(linkDir, "link")
	if err := os.Symlink(real, link); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	resolvedReal, err := filepath.EvalSymlinks(real)
	if err != nil {
		t.Fatal(err)
	}
	ws := New(filepath.Join(link, "nope", "deeper"))
	want := filepath.Join(resolvedReal, "nope", "deeper")
	if ws.Root != want {
		t.Fatalf("Root = %q, want %q", ws.Root, want)
	}
}

// TestNewCollapsesDotDotLexically matches the measured
// Path('/tmp/nope/../other').resolve() == '/private/tmp/other'.
func TestNewCollapsesDotDotLexically(t *testing.T) {
	real := t.TempDir()
	resolved, err := filepath.EvalSymlinks(real)
	if err != nil {
		t.Fatal(err)
	}
	ws := New(filepath.Join(real, "nope", "..", "other"))
	if want := filepath.Join(resolved, "other"); ws.Root != want {
		t.Fatalf("Root = %q, want %q", ws.Root, want)
	}
}

// ------------------------------------------------------- R-1.3: RequireRoot

func TestIsRoot(t *testing.T) {
	cases := []struct {
		name  string
		setup func(t *testing.T, dir string)
		want  bool
	}{
		{"empty dir", func(*testing.T, string) {}, false},
		{"manifest only, pre-first-sync", func(t *testing.T, dir string) {
			writeFile(t, filepath.Join(dir, ManifestName), "repos: []\n")
		}, true},
		{"lock alone is not a root", func(t *testing.T, dir string) {
			writeFile(t, filepath.Join(dir, LockName), "repos: {}\n")
		}, false},
		{"canonical dir", func(t *testing.T, dir string) {
			mkdir(t, filepath.Join(dir, "teams"))
		}, true},
		{"knowledge counts", func(t *testing.T, dir string) {
			mkdir(t, filepath.Join(dir, KnowledgeRoot))
		}, true},
		{"canonical name as a file does not count", func(t *testing.T, dir string) {
			writeFile(t, filepath.Join(dir, "platforms"), "")
		}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			tc.setup(t, dir)
			if got := New(dir).IsRoot(); got != tc.want {
				t.Fatalf("IsRoot() = %v, want %v", got, tc.want)
			}
		})
	}
}

// TestRequireRootMessage freezes bin/company-os:230 byte-for-byte across the
// three shapes of non-root: a plain directory, a path that is a file, and a
// path that does not exist. Python resolves all three the same way and renders
// one message.
func TestRequireRootMessage(t *testing.T) {
	base := t.TempDir()
	file := filepath.Join(base, "afile")
	writeFile(t, file, "")

	cases := map[string]string{
		"directory that is not a workspace": filepath.Join(base, "plain"),
		"root is a file":                    file,
		"root does not exist":               filepath.Join(base, "nonexistent", "deep"),
	}
	mkdir(t, filepath.Join(base, "plain"))

	for name, in := range cases {
		t.Run(name, func(t *testing.T) {
			ws := New(in)
			err := ws.RequireRoot()
			if err == nil {
				t.Fatal("RequireRoot() = nil, want an error")
			}
			if got, want := err.Error(), fmt.Sprintf(wantRootMsg, ws.Root); got != want {
				t.Fatalf("message mismatch\n got: %q\nwant: %q", got, want)
			}
			assertWorkspaceError(t, err, KindRoot, "", ws.Root)
		})
	}
}

func TestRequireRootPassesAtARoot(t *testing.T) {
	if err := New(makeRoot(t)).RequireRoot(); err != nil {
		t.Fatalf("RequireRoot() at a root = %v, want nil", err)
	}
}

// TestExemptionIsDispatchLevel documents the seam for the `init`/`scratchpad`
// exemption (bin/company-os:2774-2776, mirrored at cmd/company-os/main.go:49).
// This package always enforces the requirement; only the dispatch layer decides
// whether to ask. The assertion here is the half this package owns: RequireRoot
// does not know the command and never exempts anything.
func TestExemptionIsDispatchLevel(t *testing.T) {
	ws := New(filepath.Join(t.TempDir(), "not-a-workspace"))
	if err := ws.RequireRoot(); err == nil {
		t.Fatal("RequireRoot() exempted something; the exemption is main's, not this package's")
	}
	if got := model.CodeOf(ws.RequireRoot()); got != model.ExitWorkspace {
		t.Fatalf("exit code = %d, want %d", got, model.ExitWorkspace)
	}
}

// -------------------------------------------- R-2.10: error-returning lookups

func TestPlatformDir(t *testing.T) {
	root := makeRoot(t)
	mkdir(t, filepath.Join(root, "platforms", "alpha"))
	ws := New(root)

	got, err := ws.PlatformDir("alpha")
	if err != nil {
		t.Fatalf("PlatformDir(alpha) = %v", err)
	}
	if want := filepath.Join(ws.Platforms, "alpha"); got != want {
		t.Fatalf("PlatformDir(alpha) = %q, want %q", got, want)
	}

	_, err = ws.PlatformDir("ghost")
	if err == nil {
		t.Fatal("PlatformDir(ghost) = nil, want an error")
	}
	if got, want := err.Error(), fmt.Sprintf(wantPlatformMsg, ws.Platforms); got != want {
		t.Fatalf("message mismatch\n got: %q\nwant: %q", got, want)
	}
	assertWorkspaceError(t, err, KindPlatform, "ghost", ws.Platforms)
}

func TestTeamDir(t *testing.T) {
	root := makeRoot(t)
	mkdir(t, filepath.Join(root, "teams", "core"))
	ws := New(root)

	got, err := ws.TeamDir("core")
	if err != nil {
		t.Fatalf("TeamDir(core) = %v", err)
	}
	if want := filepath.Join(ws.Teams, "core"); got != want {
		t.Fatalf("TeamDir(core) = %q, want %q", got, want)
	}

	_, err = ws.TeamDir("ghost")
	if err == nil {
		t.Fatal("TeamDir(ghost) = nil, want an error")
	}
	if got, want := err.Error(), fmt.Sprintf(wantTeamMsg, ws.Teams); got != want {
		t.Fatalf("message mismatch\n got: %q\nwant: %q", got, want)
	}
	assertWorkspaceError(t, err, KindTeam, "ghost", ws.Teams)
}

// TestLookupsUsePathExists pins the choice of Path.exists() over is_dir() at
// bin/company-os:236 and :242 — a non-directory of the right name is "found".
func TestLookupsUsePathExists(t *testing.T) {
	root := makeRoot(t)
	writeFile(t, filepath.Join(root, "platforms", "afile"), "")
	writeFile(t, filepath.Join(root, "teams", "afile"), "")
	ws := New(root)
	if _, err := ws.PlatformDir("afile"); err != nil {
		t.Fatalf("PlatformDir on a file = %v, want it treated as found", err)
	}
	if _, err := ws.TeamDir("afile"); err != nil {
		t.Fatalf("TeamDir on a file = %v, want it treated as found", err)
	}
}

// TestLookupsFailWhenParentIsAbsent covers the pre-first-sync federated root,
// where a manifest makes IsRoot() true but platforms/ and teams/ do not exist.
func TestLookupsFailWhenParentIsAbsent(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, ManifestName), "repos: []\n")
	ws := New(dir)
	if !ws.IsRoot() {
		t.Fatal("a manifest should mark a root")
	}
	if _, err := ws.PlatformDir("ghost"); err == nil {
		t.Fatal("PlatformDir with no platforms/ = nil, want an error")
	}
	if _, err := ws.TeamDir("ghost"); err == nil {
		t.Fatal("TeamDir with no teams/ = nil, want an error")
	}
}

// ------------------------------------------------------ listings and lookup

func TestAllPlatformsAndTeams(t *testing.T) {
	root := makeRoot(t)
	for _, n := range []string{"zeta", "alpha", "Beta"} {
		mkdir(t, filepath.Join(root, "platforms", n))
		mkdir(t, filepath.Join(root, "teams", n))
	}
	// Files are filtered out; sorted() runs over names, so "Beta" precedes the
	// lowercase names exactly as Python's sorted() over Path objects does.
	writeFile(t, filepath.Join(root, "platforms", "notes.md"), "")
	ws := New(root)

	if got := base(ws.AllPlatforms()); !equal(got, []string{"Beta", "alpha", "zeta"}) {
		t.Fatalf("AllPlatforms() = %v", got)
	}
	if got := base(ws.AllTeams()); !equal(got, []string{"Beta", "alpha", "zeta"}) {
		t.Fatalf("AllTeams() = %v", got)
	}
}

func TestAllPlatformsAbsentDirIsEmpty(t *testing.T) {
	ws := New(t.TempDir())
	if got := ws.AllPlatforms(); len(got) != 0 {
		t.Fatalf("AllPlatforms() = %v, want empty", got)
	}
	if got := ws.AllTeams(); len(got) != 0 {
		t.Fatalf("AllTeams() = %v, want empty", got)
	}
}

// TestAllPlatformsFollowsSymlinkedDirs matches Path.is_dir(), which follows
// symlinks; os.DirEntry.IsDir() does not, which is the easy way to get this
// wrong.
func TestAllPlatformsFollowsSymlinkedDirs(t *testing.T) {
	root := makeRoot(t)
	target := filepath.Join(t.TempDir(), "real-platform")
	mkdir(t, target)
	if err := os.Symlink(target, filepath.Join(root, "platforms", "linked")); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	if got := base(New(root).AllPlatforms()); !equal(got, []string{"linked"}) {
		t.Fatalf("AllPlatforms() = %v, want [linked]", got)
	}
}

func TestFindComponent(t *testing.T) {
	root := makeRoot(t)
	mkdir(t, filepath.Join(root, "platforms", "alpha", "components"))
	mkdir(t, filepath.Join(root, "platforms", "beta", "components"))
	writeFile(t, filepath.Join(root, "platforms", "beta", "components", "svc.yaml"), "id: svc\n")
	ws := New(root)

	platform, descriptor, found := ws.FindComponent("svc")
	if !found {
		t.Fatal("FindComponent(svc) not found")
	}
	if platform != "beta" {
		t.Fatalf("platform = %q, want beta", platform)
	}
	if want := filepath.Join(ws.Platforms, "beta", "components", "svc.yaml"); descriptor != want {
		t.Fatalf("descriptor = %q, want %q", descriptor, want)
	}

	if _, _, found := ws.FindComponent("ghost"); found {
		t.Fatal("FindComponent(ghost) reported found; a miss is not an error either")
	}
}

// TestFindComponentPrefersFirstPlatform pins the first-match-wins scan order of
// bin/company-os:259-262, which follows all_platforms()' sorted order.
func TestFindComponentPrefersFirstPlatform(t *testing.T) {
	root := makeRoot(t)
	for _, p := range []string{"alpha", "beta"} {
		mkdir(t, filepath.Join(root, "platforms", p, "components"))
		writeFile(t, filepath.Join(root, "platforms", p, "components", "svc.yaml"), "id: svc\n")
	}
	if platform, _, _ := New(root).FindComponent("svc"); platform != "alpha" {
		t.Fatalf("platform = %q, want alpha (first in sorted order)", platform)
	}
}

// ------------------------------------------------------------------ helpers

// assertWorkspaceError checks the classification contract: main must be able to
// reach both the exit code and the identity of the missing object by type, with
// no substring matching on the message.
func assertWorkspaceError(t *testing.T, err error, kind NotFoundKind, id, dir string) {
	t.Helper()
	if got := model.CodeOf(err); got != model.ExitWorkspace {
		t.Fatalf("model.CodeOf = %d, want %d", got, model.ExitWorkspace)
	}
	var nf *NotFoundError
	if !errors.As(err, &nf) {
		t.Fatalf("errors.As(*NotFoundError) failed for %T", err)
	}
	if nf.Kind != kind || nf.ID != id || nf.Dir != dir {
		t.Fatalf("got {%s %q %q}, want {%s %q %q}", nf.Kind, nf.ID, nf.Dir, kind, id, dir)
	}
	var me *model.Error
	if !errors.As(err, &me) {
		t.Fatal("errors.As(*model.Error) failed; main could not classify this")
	}
	if me.Msg != err.Error() {
		t.Fatalf("wrapped message %q differs from Error() %q", me.Msg, err.Error())
	}
}

func mkdir(t *testing.T, p string) {
	t.Helper()
	if err := os.MkdirAll(p, 0o755); err != nil {
		t.Fatal(err)
	}
}

func writeFile(t *testing.T, p, body string) {
	t.Helper()
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func base(paths []string) []string {
	out := make([]string, len(paths))
	for i, p := range paths {
		out[i] = filepath.Base(p)
	}
	return out
}

func equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
