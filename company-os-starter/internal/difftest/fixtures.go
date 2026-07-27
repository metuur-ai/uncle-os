package difftest

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// tamperStep is a pseudo-step, not an argv: the runner recognizes it and
// performs a deterministic in-tree mutation instead of executing the binary.
// It exists to drive gate [8/8]'s hash-integrity FAIL on a hand-edited slice,
// which is otherwise unreachable from the CLI surface — the whole point of the
// gate is that nothing the CLI does can produce it.
const tamperStep = "__tamper__"

// committedFixtures are checked in under examples/ and never mutated: every
// invocation runs against a fresh copy.
var committedFixtures = map[string]string{
	"workspace":       "workspace",
	"standalone-team": "standalone-team",
	"federated":       "federated",
	"banking-small":   filepath.Join("banking", "small-company"),
	"banking-rails":   filepath.Join("banking", "bank", "workspaces", "team-payments-rails"),
	"banking-fraud":   filepath.Join("banking", "bank", "workspaces", "team-fraud-detection"),
	// Failure-path fixtures (task 0.2 / R-0.9). They exist to fail; a fixture
	// that fails is passing evidence as long as it fails the same way.
	"failing-workspace":        "failing-workspace",
	"failing-federated":        "failing-federated",
	"failing-federated-nolock": "failing-federated-nolock",
}

const zeros = "0000000000000000000000000000000000000000"
const ones = "1111111111111111111111111111111111111111"

// badManifests each trigger a distinct manifest-load rejection. One
// `workspace status` and one `workspace sync` invocation per entry.
var badManifests = map[string]string{
	"not-a-mapping":    "- just\n- a\n- list\n",
	"no-repos-key":     "version: 1\nsomething: else\n",
	"empty-repos":      "version: 1\nrepos: []\n",
	"repo-not-mapping": "version: 1\nrepos:\n  - just-a-string\n",
	"missing-url": "version: 1\nrepos:\n  - name: a\n    pin: {commit: '" + zeros + "'}\n" +
		"    localDirectory: platforms/a\n    paths: [governance/]\n",
	"missing-pin": "version: 1\nrepos:\n  - name: a\n    url: file:///nowhere\n" +
		"    localDirectory: platforms/a\n    paths: [governance/]\n",
	"duplicate-name": "version: 1\nrepos:\n" +
		"  - name: a\n    url: file:///n\n    pin: {commit: '" + zeros + "'}\n" +
		"    localDirectory: platforms/a\n    paths: [governance/]\n" +
		"  - name: a\n    url: file:///n\n    pin: {commit: '" + ones + "'}\n" +
		"    localDirectory: platforms/b\n    paths: [governance/]\n",
	"bad-name-chars": "version: 1\nrepos:\n  - name: 'bad name/slash'\n    url: file:///n\n" +
		"    pin: {commit: '" + zeros + "'}\n" +
		"    localDirectory: platforms/a\n    paths: [governance/]\n",
	"renamed-root-key": "version: 1\nrepos:\n  - name: a\n    url: file:///n\n" +
		"    pin: {commit: '" + zeros + "'}\n" +
		"    root: platforms/a\n    paths: [governance/]\n",
	"floating-pin-branch": "version: 1\nrepos:\n  - name: a\n    url: file:///n\n" +
		"    pin: {branch: main}\n" +
		"    localDirectory: platforms/a\n    paths: [governance/]\n",
	"pin-both-commit-and-tag": "version: 1\nrepos:\n  - name: a\n    url: file:///n\n" +
		"    pin: {commit: '" + zeros + "', tag: v1}\n" +
		"    localDirectory: platforms/a\n    paths: [governance/]\n",
	"pin-not-a-mapping": "version: 1\nrepos:\n  - name: a\n    url: file:///n\n" +
		"    pin: main\n" +
		"    localDirectory: platforms/a\n    paths: [governance/]\n",
	"empty-paths": "version: 1\nrepos:\n  - name: a\n    url: file:///n\n" +
		"    pin: {commit: '" + zeros + "'}\n    localDirectory: platforms/a\n    paths: []\n",
	"absolute-localdir": "version: 1\nrepos:\n  - name: a\n    url: file:///n\n" +
		"    pin: {commit: '" + zeros + "'}\n" +
		"    localDirectory: /etc/passwd\n    paths: [governance/]\n",
	"escaping-localdir": "version: 1\nrepos:\n  - name: a\n    url: file:///n\n" +
		"    pin: {commit: '" + zeros + "'}\n" +
		"    localDirectory: ../outside\n    paths: [governance/]\n",
	"non-canonical-localdir": "version: 1\nrepos:\n  - name: a\n    url: file:///n\n" +
		"    pin: {commit: '" + zeros + "'}\n" +
		"    localDirectory: elsewhere/a\n    paths: [governance/]\n",
	"bare-knowledge-root": "version: 1\nrepos:\n  - name: a\n    url: file:///n\n" +
		"    pin: {commit: '" + zeros + "'}\n" +
		"    localDirectory: knowledge\n    paths: [docs/]\n",
	"overlapping-targets": "version: 1\nrepos:\n" +
		"  - name: a\n    url: file:///n\n    pin: {commit: '" + zeros + "'}\n" +
		"    localDirectory: platforms/x\n    paths: [governance/]\n" +
		"  - name: b\n    url: file:///n\n    pin: {commit: '" + ones + "'}\n" +
		"    localDirectory: platforms/x/inner\n    paths: [governance/]\n",
	"slices-not-a-list": "version: 1\nrepos:\n  - name: a\n    url: file:///n\n" +
		"    pin: {commit: '" + zeros + "'}\n    slices: nope\n",
	"slices-entry-not-mapping": "version: 1\nrepos:\n  - name: a\n    url: file:///n\n" +
		"    pin: {commit: '" + zeros + "'}\n    slices: [nope]\n",
	"slices-missing-localdir": "version: 1\nrepos:\n  - name: a\n    url: file:///n\n" +
		"    pin: {commit: '" + zeros + "'}\n" +
		"    slices:\n      - paths: [governance/]\n",
	"slices-uses-root-key": "version: 1\nrepos:\n  - name: a\n    url: file:///n\n" +
		"    pin: {commit: '" + zeros + "'}\n" +
		"    slices:\n      - root: platforms/a\n        paths: [g/]\n",
	"slices-and-localdir": "version: 1\nrepos:\n  - name: a\n    url: file:///n\n" +
		"    pin: {commit: '" + zeros + "'}\n" +
		"    localDirectory: platforms/a\n    paths: [governance/]\n" +
		"    slices:\n      - localDirectory: platforms/b\n        paths: [governance/]\n",
	"neither-slices-nor-localdir": "version: 1\nrepos:\n  - name: a\n    url: file:///n\n" +
		"    pin: {commit: '" + zeros + "'}\n",
	"malformed-yaml": "version: 1\nrepos:\n  - name: [unclosed\n",
}

// badManifestNames returns the keys sorted, so corpus ids are generated in a
// stable order regardless of Go's map iteration.
func badManifestNames() []string {
	names := make([]string, 0, len(badManifests))
	for k := range badManifests {
		names = append(names, k)
	}
	sort.Strings(names)
	return names
}

// fixtureSet resolves a fixture name to a directory on disk. Committed fixtures
// live under examples/; synthetic ones are built into a temp dir at startup.
type fixtureSet struct {
	examples string
	// base is the temp directory synthetic fixtures are built into. It is
	// normalized out of every recorded stream and file, because it changes every
	// run and `workspace sync` writes the git source path into the lock.
	base       string
	synthetic  map[string]string
	gitOK      bool
	gitSkipWhy string
}

// ErrUnavailable explains why a fixture cannot be used. Empty means usable.
// A missing fixture must be reported loudly — a corpus entry that silently
// vanishes is coverage lost without a diff to notice it.
func (f *fixtureSet) unavailable(name string) string {
	if _, ok := f.synthetic[name]; ok {
		return ""
	}
	rel, ok := committedFixtures[name]
	if !ok {
		if f.gitSkipWhy != "" {
			return fmt.Sprintf("fixture %q could not be synthesized: %s", name, f.gitSkipWhy)
		}
		return fmt.Sprintf("fixture %q is not registered", name)
	}
	dir := filepath.Join(f.examples, rel)
	if st, err := os.Stat(dir); err != nil || !st.IsDir() {
		return fmt.Sprintf("fixture %q missing at %s", name, dir)
	}
	return ""
}

func (f *fixtureSet) path(name string) string {
	if p, ok := f.synthetic[name]; ok {
		return p
	}
	return filepath.Join(f.examples, committedFixtures[name])
}

// newFixtureSet builds the fixtures that are not committed: an empty dir, the
// bad-manifest workspaces, and (git permitting) a real source repo plus the
// three federated workspaces that point at it.
func newFixtureSet(examples, base string) (*fixtureSet, error) {
	f := &fixtureSet{examples: examples, base: base, synthetic: map[string]string{}}

	empty := filepath.Join(base, "empty")
	if err := os.MkdirAll(empty, 0o755); err != nil {
		return nil, err
	}
	f.synthetic["empty"] = empty

	standalone := filepath.Join(examples, committedFixtures["standalone-team"])
	for _, name := range badManifestNames() {
		dir := filepath.Join(base, "badmanifest-"+name)
		if err := copyTree(standalone, dir); err != nil {
			return nil, fmt.Errorf("badmanifest-%s: %w", name, err)
		}
		if err := os.WriteFile(filepath.Join(dir, "workspace.yaml"),
			[]byte(badManifests[name]), 0o644); err != nil {
			return nil, err
		}
		f.synthetic["badmanifest-"+name] = dir
	}

	if why := gitUsable(); why != "" {
		f.gitSkipWhy = why
		return f, nil
	}
	if err := f.buildGitFixtures(base); err != nil {
		// A git fixture that fails to build is a skip reason, not a hard error:
		// the git-free 90% of the corpus is still worth running.
		f.gitSkipWhy = err.Error()
		return f, nil
	}
	f.gitOK = true
	return f, nil
}

func (f *fixtureSet) buildGitFixtures(base string) error {
	src := filepath.Join(base, "gitsrc")
	files := map[string]string{
		filepath.Join("governance", "requirements.yaml"): "version: 1\nplatform: testplat\nrequirements: []\n",
		filepath.Join("components", "foo.yaml"): "id: foo\ncomponentType: service\n" +
			"ownership:\n  accountableTeam: team://none\n",
		filepath.Join("docs", "sdd", "spec.md"): "# Spec\n",
		"README.md":                             "# not governance - must not be sliced\n",
		// A Python file on purpose: the slice allowlist must exclude src/ from a
		// governance slice regardless of what language the repo is written in.
		filepath.Join("src", "app.py"): "print('not governance')\n",
	}
	for rel, body := range files {
		full := filepath.Join(src, rel)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(full, []byte(body), 0o644); err != nil {
			return err
		}
	}

	// Fixed identity and dates so the commit SHA is reproducible run to run —
	// the SHA is baked into the manifest, so a moving SHA would move the golden.
	env := append(os.Environ(),
		"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@e",
		"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@e",
		"GIT_AUTHOR_DATE=2020-01-01T00:00:00+0000",
		"GIT_COMMITTER_DATE=2020-01-01T00:00:00+0000",
	)
	for _, args := range [][]string{
		{"init", "-q", "-b", "main"}, {"add", "-A"}, {"commit", "-q", "-m", "init"},
	} {
		cmd := exec.Command("git", append([]string{"-C", src}, args...)...)
		cmd.Env = env
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("git setup failed: %s", strings.TrimSpace(string(out)))
		}
	}
	shaOut, err := exec.Command("git", "-C", src, "rev-parse", "HEAD").Output()
	if err != nil {
		return fmt.Errorf("git rev-parse failed: %w", err)
	}
	sha := strings.TrimSpace(string(shaOut))

	manifest := func(commit string) string {
		return "version: 1\nrepos:\n" +
			"  - name: testplat\n" +
			"    url: file://" + src + "\n" +
			"    pin:\n" +
			"      commit: " + commit + "\n" +
			"    slices:\n" +
			"      - localDirectory: platforms/testplat\n" +
			"        paths: [governance/, components/]\n" +
			"      - localDirectory: knowledge/testplat\n" +
			"        paths: [docs/sdd]\n"
	}
	for _, e := range []struct{ key, commit string }{
		{"gitfed", sha},
		{"gitfed-badpin", sha[:8]},
		{"gitfed-missingref", strings.Repeat("b", 40)},
	} {
		dir := filepath.Join(base, e.key)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(dir, "workspace.yaml"),
			[]byte(manifest(e.commit)), 0o644); err != nil {
			return err
		}
		f.synthetic[e.key] = dir
	}
	return nil
}

var gitVersionRe = regexp.MustCompile(`(\d+)\.(\d+)`)

// gitUsable returns "" when git can drive the sync fixtures, else the reason it
// cannot. 2.27 is the floor because cone-mode sparse-checkout is what makes an
// include-only slice possible.
func gitUsable() string {
	if _, err := exec.LookPath("git"); err != nil {
		return "git not found on PATH"
	}
	out, err := exec.Command("git", "--version").Output()
	if err != nil {
		return fmt.Sprintf("git --version failed: %v", err)
	}
	m := gitVersionRe.FindStringSubmatch(string(out))
	if m == nil {
		return fmt.Sprintf("could not parse `git --version`: %q", strings.TrimSpace(string(out)))
	}
	major, _ := strconv.Atoi(m[1])
	minor, _ := strconv.Atoi(m[2])
	if major < 2 || (major == 2 && minor < 27) {
		return fmt.Sprintf("git %d.%d < 2.27 (cone-mode sparse-checkout required)", major, minor)
	}
	return ""
}
