package difftest

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

var update = flag.Bool("update", false,
	"rewrite the golden files from the current binary's behaviour")

var (
	testBinary   string
	testFixtures *fixtureSet
	examplesDir  string
)

func TestMain(m *testing.M) {
	flag.Parse()
	code, err := setup(m)
	if err != nil {
		fmt.Fprintf(os.Stderr, "difftest setup: %v\n", err)
		os.Exit(1)
	}
	os.Exit(code)
}

func setup(m *testing.M) (int, error) {
	base, err := os.MkdirTemp("", "company-os-difftest-")
	if err != nil {
		return 0, err
	}
	defer func() {
		_ = chmodWritable(base)
		_ = os.RemoveAll(base)
	}()

	// examples/ lives beside the Go module, not inside it: the fixtures predate
	// the port and are shared with acceptance.sh.
	examplesDir, err = filepath.Abs(filepath.Join("..", "..", "..", "examples"))
	if err != nil {
		return 0, err
	}
	if st, err := os.Stat(examplesDir); err != nil || !st.IsDir() {
		return 0, fmt.Errorf("examples/ not found at %s", examplesDir)
	}

	testBinary = filepath.Join(base, "company-os")
	build := exec.Command("go", "build", "-o", testBinary, "../../cmd/company-os")
	if out, err := build.CombinedOutput(); err != nil {
		return 0, fmt.Errorf("building the binary: %v\n%s", err, out)
	}

	testFixtures, err = newFixtureSet(examplesDir, filepath.Join(base, "fixtures"))
	if err != nil {
		return 0, fmt.Errorf("building fixtures: %w", err)
	}
	return m.Run(), nil
}

// TestCorpus runs every invocation and compares the whole recorded run —
// per-step exit/stdout/stderr plus the resulting file tree — against its golden.
//
// This is a characterization test, not a differential: it proves behaviour has
// not changed since a human last reviewed the golden. Re-baseline with
// `go test ./internal/difftest -update` and READ THE DIFF; a golden accepted
// without reading it protects nothing.
func TestCorpus(t *testing.T) {
	for _, in := range Corpus() {
		t.Run(in.ID, func(t *testing.T) {
			t.Parallel()
			if why := testFixtures.unavailable(in.Fixture); why != "" {
				// Loud skip: a fixture that quietly vanishes is coverage lost
				// with no diff to notice it.
				t.Skipf("fixture unavailable: %s", why)
			}
			workdir := t.TempDir()
			// `workspace sync` materializes slices at 0444/0555 — the gate [8/8]
			// invariant — which TempDir's own RemoveAll cannot delete. Cleanups
			// run LIFO, so registering this after TempDir's makes it run first.
			// It must not run before the snapshot: those mode bits are recorded.
			t.Cleanup(func() { _ = chmodWritable(workdir) })

			res, err := runInvocation(testBinary, in, testFixtures, workdir)
			if err != nil {
				t.Fatalf("running %s: %v", in.ID, err)
			}
			got := render(in, res)
			path := goldenPath(in.ID)

			if *update {
				if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
					t.Fatal(err)
				}
				return
			}

			want, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("golden missing for %s (%v)\nrun: go test ./internal/difftest -update",
					in.ID, err)
			}
			if got != string(want) {
				t.Errorf("behaviour changed for %s\n%s", in.ID, firstDiff(string(want), got))
			}
		})
	}
	// t.Cleanup on the parent runs after all parallel subtests finish, so the
	// temp fixtures survive until every invocation has used them.
}

func goldenPath(id string) string {
	return filepath.Join("testdata", filepath.FromSlash(id)+".txt")
}

// firstDiff reports the first differing line with a little context. The full
// golden is on disk; dumping both copies into the test log helps nobody.
func firstDiff(want, got string) string {
	w := strings.Split(want, "\n")
	g := strings.Split(got, "\n")
	for i := 0; i < len(w) || i < len(g); i++ {
		var wl, gl string
		var wok, gok bool
		if i < len(w) {
			wl, wok = w[i], true
		}
		if i < len(g) {
			gl, gok = g[i], true
		}
		if wl == gl && wok == gok {
			continue
		}
		var b strings.Builder
		fmt.Fprintf(&b, "  first difference at line %d:\n", i+1)
		for c := max(0, i-3); c < i; c++ {
			fmt.Fprintf(&b, "     %s\n", w[c])
		}
		if wok {
			fmt.Fprintf(&b, "  -  %s\n", wl)
		} else {
			fmt.Fprintf(&b, "  -  (golden ends here)\n")
		}
		if gok {
			fmt.Fprintf(&b, "  +  %s\n", gl)
		} else {
			fmt.Fprintf(&b, "  +  (output ends here)\n")
		}
		fmt.Fprintf(&b, "  (%d golden lines, %d output lines)\n", len(w), len(g))
		return b.String()
	}
	return "  files differ but no line differs (trailing bytes?)"
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// TestCorpusMatchesTheInheritedIDSet pins the invocation set this suite
// inherited from examples/differential.py, the Python differential harness this
// package replaces (deleted in the same change).
//
// Without this, dropping an invocation during the port — or later — would be
// invisible: the suite would still pass, just over less. testdata/corpus-ids.txt
// was captured from `differential.py --list` against the last commit that still
// had it, so it is evidence rather than a restatement of the Go corpus.
//
// Adding invocations is expected and allowed; the assertion is one-directional.
// Removing one must be a deliberate edit to the pinned list, in a reviewable
// diff, with a reason.
func TestCorpusMatchesTheInheritedIDSet(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("testdata", "corpus-ids.txt"))
	if err != nil {
		t.Fatalf("reading the inherited id list: %v", err)
	}
	inherited := map[string]bool{}
	for _, line := range strings.Split(string(raw), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		inherited[line] = true
	}
	have := map[string]bool{}
	for _, in := range Corpus() {
		if have[in.ID] {
			t.Errorf("duplicate invocation id %q", in.ID)
		}
		have[in.ID] = true
	}

	var missing []string
	for id := range inherited {
		if !have[id] {
			missing = append(missing, id)
		}
	}
	sort.Strings(missing)
	if len(missing) > 0 {
		t.Errorf("%d invocation(s) from the Python corpus are not in the Go corpus — "+
			"this is coverage lost, not a cosmetic difference:\n  %s",
			len(missing), strings.Join(missing, "\n  "))
	}
	if len(inherited) == 0 {
		t.Fatal("the inherited id list is empty; the pin is not pinning anything")
	}
	t.Logf("%d inherited invocations, %d in the Go corpus", len(inherited), len(have))
}

// TestEveryInvocationHasAGolden catches the case where a corpus entry was added
// but never baselined, which would otherwise only surface as a confusing
// "golden missing" failure inside a parallel subtest.
func TestEveryInvocationHasAGolden(t *testing.T) {
	if *update {
		t.Skip("goldens are being written")
	}
	var missing []string
	for _, in := range Corpus() {
		if _, err := os.Stat(goldenPath(in.ID)); err != nil {
			missing = append(missing, in.ID)
		}
	}
	if len(missing) > 0 {
		t.Errorf("%d invocation(s) have no golden; run `go test ./internal/difftest -update`:\n  %s",
			len(missing), strings.Join(missing, "\n  "))
	}
}

// TestNoOrphanGoldens is the other direction: a golden with no invocation is a
// test that silently stopped running.
func TestNoOrphanGoldens(t *testing.T) {
	live := map[string]bool{}
	for _, in := range Corpus() {
		live[goldenPath(in.ID)] = true
	}
	var orphans []string
	err := filepath.Walk("testdata", func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".txt") {
			return err
		}
		if filepath.Base(path) == "corpus-ids.txt" {
			return nil
		}
		if !live[path] {
			orphans = append(orphans, path)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(orphans) > 0 {
		t.Errorf("%d golden(s) have no corpus entry — delete them or restore the "+
			"invocation:\n  %s", len(orphans), strings.Join(orphans, "\n  "))
	}
}

// TestGitFixturesAreAvailable makes the git-gated portion of the corpus visible
// rather than letting it skip in silence. It is a skip, not a failure: git 2.27
// is a real environment requirement, not a bug in the suite.
func TestGitFixturesAreAvailable(t *testing.T) {
	if !testFixtures.gitOK {
		t.Skipf("the workspace-git corpus did not run: %s", testFixtures.gitSkipWhy)
	}
}
