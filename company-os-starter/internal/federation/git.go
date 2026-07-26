package federation

// Git plumbing (bin/company-os:2249-2325). The only subprocess users in the
// whole module.
//
// Every failure here is model.ExitExternalTool (6) with one measured exception:
// an abbreviated commit SHA in workspace.yaml is an ARTIFACT fault (4), not a
// tool fault — both `git fetch` and `git rev-parse` succeeded, and what is wrong
// is the pin the human wrote. See .devlocal/go-port/exit-code-map.md § `:2318`.

import (
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/model"
)

// MinGit is the floor for cone-mode sparse-checkout plus a stable partial
// clone, both of which landed in git 2.27 (GPF-R-7.7).
var MinGit = [2]int{2, 27}

var gitVersionRe = regexp.MustCompile(`(\d+)\.(\d+)(?:\.(\d+))?`)

// gitAvailable is shutil.which("git") (bin/company-os:2249).
func gitAvailable() bool {
	_, err := exec.LookPath("git")
	return err == nil
}

// runGit is _run_git (bin/company-os:2253-2264). check=false returns the failed
// result instead of erroring, matching the keyword default's one caller.
func runGit(args []string, cwd string, check bool) (stdout string, err error) {
	cmd := exec.Command("git", args...)
	if cwd != "" {
		cmd.Dir = cwd
	}
	var outBuf, errBuf strings.Builder
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	runErr := cmd.Run()
	if _, isExit := runErr.(*exec.ExitError); runErr != nil && !isExit {
		// exec.Command defers the PATH lookup to Run, so a missing git surfaces
		// here rather than at construction — Python's FileNotFoundError branch.
		return "", model.Errorf(model.ExitExternalTool,
			"git not found on PATH — federation commands require git; "+
				"monorepo commands do not. (GPF-R-7.7)")
	}
	code := cmd.ProcessState.ExitCode()
	if check && code != 0 {
		// `(r.stderr or r.stdout).strip()` — a truthiness pick, so an empty
		// stderr falls through to stdout rather than printing nothing.
		detail := errBuf.String()
		if detail == "" {
			detail = outBuf.String()
		}
		return "", model.Errorf(model.ExitExternalTool,
			"`git %s` failed (exit %d):\n%s",
			strings.Join(args, " "), code, strings.TrimSpace(detail))
	}
	return outBuf.String(), nil
}

// gitVersion is _git_version (bin/company-os:2267-2271).
func gitVersion() ([3]int, error) {
	out, err := runGit([]string{"--version"}, "", true)
	if err != nil {
		return [3]int{}, err
	}
	m := gitVersionRe.FindStringSubmatch(out)
	if m == nil {
		return [3]int{}, model.Errorf(model.ExitExternalTool,
			"could not parse `git --version`")
	}
	var v [3]int
	for i := 0; i < 3; i++ {
		if m[i+1] != "" {
			v[i], _ = strconv.Atoi(m[i+1])
		}
	}
	return v, nil
}

// RequireGit guards the federation entry points (bin/company-os:2274-2287).
// Monorepo commands never call it, so a missing or old git breaks federation
// only (GPF-R-7.7).
func RequireGit() error {
	if !gitAvailable() {
		return model.Errorf(model.ExitExternalTool,
			"git is required for federation (workspace sync/status) but was not "+
				"found on PATH. monorepo commands do not need git. "+
				"install git, then: company-os workspace sync")
	}
	ver, err := gitVersion()
	if err != nil {
		return err
	}
	if ver[0] < MinGit[0] || (ver[0] == MinGit[0] && ver[1] < MinGit[1]) {
		return model.Errorf(model.ExitExternalTool,
			"git %s is too old for federation; >= %s is required (cone-mode "+
				"sparse-checkout + partial clone). upgrade git, then: "+
				"company-os workspace sync",
			joinInts(ver[:]), joinInts([]int{MinGit[0], MinGit[1]}))
	}
	return nil
}

func joinInts(v []int) string {
	parts := make([]string, len(v))
	for i, n := range v {
		parts[i] = strconv.Itoa(n)
	}
	return strings.Join(parts, ".")
}

// sparseDirs is _sparse_dirs (bin/company-os:2290-2299): the top-level
// directories of the allowlist, deduplicated in first-seen order. Cone mode
// keeps root files anyway; the precise allowlist is enforced by the copy step.
func sparseDirs(paths []string) []string {
	var dirs []string
	for _, p := range paths {
		top := strings.SplitN(strings.Trim(p, "/"), "/", 2)[0]
		if top == "" || containsString(dirs, top) {
			continue
		}
		dirs = append(dirs, top)
	}
	return dirs
}

// fetchPinned is _fetch_pinned (bin/company-os:2302-2320): a blobless, shallow
// fetch of the pinned object, returning the resolved commit SHA. Tags resolve
// here (GPF-R-6.4), which is what makes the lock reproducible.
func fetchPinned(cache, url string, pin Pin) (string, error) {
	if err := mkdirAll(filepath.Dir(cache)); err != nil {
		return "", err
	}
	if !isDir(filepath.Join(cache, ".git")) {
		if _, err := runGit([]string{"init", "--quiet", cache}, "", true); err != nil {
			return "", err
		}
		if _, err := runGit([]string{"remote", "add", "origin", url}, cache, true); err != nil {
			return "", err
		}
	} else if _, err := runGit([]string{"remote", "set-url", "origin", url}, cache, true); err != nil {
		return "", err
	}
	fetchRef := pin.Ref
	if pin.Kind != "commit" {
		fetchRef = "refs/tags/" + pin.Ref
	}
	// --filter=blob:none keeps the fetch to commits+trees on capable servers
	// (ignored with a warning elsewhere); --depth 1 keeps it shallow (GPF-R-7.1).
	if _, err := runGit([]string{"fetch", "--filter=blob:none", "--depth", "1",
		"origin", fetchRef}, cache, true); err != nil {
		return "", err
	}
	out, err := runGit([]string{"rev-parse", "FETCH_HEAD^{commit}"}, cache, true)
	if err != nil {
		return "", err
	}
	resolved := strings.TrimSpace(out)
	if pin.Kind == "commit" && !strings.HasPrefix(resolved, pin.Ref) && pin.Ref != resolved {
		return "", model.Errorf(model.ExitArtifact,
			"pinned commit '%s' resolved to '%s' — pin a full commit SHA",
			pin.Ref, resolved)
	}
	return resolved, nil
}

// checkoutSlice is _checkout_slice (bin/company-os:2323-2328): limit the cache
// working tree to the allowlisted top dirs, then check out the pinned SHA. Same
// SHA plus same sparse set gives an identical tree.
func checkoutSlice(cache, sha string, dirs []string) error {
	if _, err := runGit([]string{"sparse-checkout", "init", "--cone"}, cache, true); err != nil {
		return err
	}
	if _, err := runGit(append([]string{"sparse-checkout", "set"}, dirs...), cache, true); err != nil {
		return err
	}
	_, err := runGit([]string{"checkout", "--quiet", "--detach", sha}, cache, true)
	return err
}

func containsString(haystack []string, needle string) bool {
	for _, s := range haystack {
		if s == needle {
			return true
		}
	}
	return false
}
