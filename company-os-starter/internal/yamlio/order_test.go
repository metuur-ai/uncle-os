package yamlio

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
)

// stabilityRuns is the repeat count for every "same input, same bytes" claim in
// this file. Go randomizes map iteration per range statement, not per process,
// so repeating inside one test is a real probe: a single pass over a Go map
// proves nothing, and a handful of passes can pass by luck on a two-key map.
const stabilityRuns = 64

// ---------------------------------------------------------------- fixture roots

// repoRoot locates the checkout so the tests below can run against the REAL
// committed fixtures rather than synthetic ones. The whole point of task 1.4 is
// reproducing an order that already exists on disk.
func repoRoot(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatalf("resolve repo root: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "examples")); err != nil {
		t.Skipf("no examples/ fixtures at %s: %v", root, err)
	}
	return root
}

func loadFixture(t *testing.T, path string) *Document {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	doc, err := Load(data)
	if err != nil {
		t.Fatalf("load %s: %v", path, err)
	}
	return doc
}

// lockFilesPairs returns the `files:` pairs of a lock's first repo, in document
// order.
func lockFilesPairs(t *testing.T, lockPath string) []Pair {
	t.Helper()
	doc := loadFixture(t, lockPath)
	repos := Deref(MapGet(doc.Root(), "repos"))
	if repos == nil || len(repos.Content) == 0 {
		t.Fatalf("%s: no repos", lockPath)
	}
	files := MapGet(repos.Content[0], "files")
	if files == nil {
		t.Fatalf("%s: first repo has no files map", lockPath)
	}
	return MapPairs(files)
}

// --------------------------------------------- site [A]: workspace.lock.yaml

// TestLockFilesOrderIsWalkOrderNotSorted rebuilds the `files:` map of
// examples/federated/workspace.lock.yaml the way _materialize_all +
// hash_tree do — manifest slice order, then each slice's `paths:` list order,
// then PathLess within each subtree — and asserts the result reproduces the
// committed lock byte layout.
//
// The negative half is the point: the same key set sorted globally does NOT
// match, because the manifest lists governance/ before components/. Anything
// that "just sorts" here silently rewrites a committed artifact.
func TestLockFilesOrderIsWalkOrderNotSorted(t *testing.T) {
	root := repoRoot(t)
	fed := filepath.Join(root, "examples", "federated")

	manifest := loadFixture(t, filepath.Join(fed, "workspace.yaml"))
	repos := Deref(MapGet(manifest.Root(), "repos"))
	if repos == nil || len(repos.Content) == 0 {
		t.Fatal("federated manifest has no repos")
	}
	repo := repos.Content[0]
	localDir := MapGet(repo, "localDirectory").Value

	// hash_tree, transliterated: for each allowlisted path, walk its subtree and
	// take the files in PathLess order. Slice order is the outer loop; this
	// fixture has one slice, so the paths: list order is what is under test.
	files := NewOrderedMap()
	for _, p := range Deref(MapGet(repo, "paths")).Content {
		rel := strings.Trim(p.Value, "/")
		if rel == "" {
			continue
		}
		src := filepath.Join(fed, localDir, rel)
		var found []string
		_ = filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
				return nil
			}
			r, rerr := filepath.Rel(fed, path)
			if rerr != nil {
				return nil
			}
			found = append(found, filepath.ToSlash(r))
			return nil
		})
		SortPaths(found)
		for _, f := range found {
			files.SetString(f, "sha")
		}
	}

	want := lockFilesPairs(t, filepath.Join(fed, "workspace.lock.yaml"))
	wantKeys := make([]string, len(want))
	for i, p := range want {
		wantKeys[i] = p.Key
	}

	if got := files.Keys(); !reflect.DeepEqual(got, wantKeys) {
		t.Errorf("rebuilt walk order does not reproduce the committed lock\n got: %v\nwant: %v", got, wantKeys)
	}

	// Teeth: the frozen order must actually differ from sorted, or this test
	// would pass for a sorting implementation too.
	if sorted := files.SortedKeys(); reflect.DeepEqual(sorted, wantKeys) {
		t.Fatalf("fixture no longer distinguishes walk order from sorted order (%v) — "+
			"this test has lost its teeth; pick a fixture whose paths: list is not alphabetical", sorted)
	}
}

// TestLockFilesEmitIsStable emits the same OrderedMap stabilityRuns times and
// requires byte-identical output every time. R-0.11's emission half.
func TestLockFilesEmitIsStable(t *testing.T) {
	root := repoRoot(t)
	pairs := lockFilesPairs(t, filepath.Join(root, "examples", "federated", "workspace.lock.yaml"))

	om := NewOrderedMap()
	for _, p := range pairs {
		om.SetString(p.Key, p.Value.Value)
	}

	var first []byte
	for i := 0; i < stabilityRuns; i++ {
		got, err := NewDocument(om.Node()).Bytes()
		if err != nil {
			t.Fatalf("run %d: emit: %v", i, err)
		}
		if i == 0 {
			first = got
			continue
		}
		if !bytes.Equal(first, got) {
			t.Fatalf("run %d emitted different bytes than run 0\nfirst:\n%s\ngot:\n%s", i, first, got)
		}
	}
	// Sanity: the emission carries the authored order, not an alphabetical one.
	if !bytes.HasPrefix(first, []byte(pairs[0].Key+":")) {
		t.Errorf("emitted map does not lead with the first authored key %q:\n%s", pairs[0].Key, first)
	}
}

// ---------------------------------------------------- site [B]: gate 8 order

// TestGate8FindingOrderMatchesGolden asserts that iterating the failing fixture's
// lock with MapPairs reproduces the [FAIL] line order frozen in
// examples/failing-federated-golden-validate.txt.
//
// The golden reports governance/requirements.yaml before
// components/svc-sliced.yaml — the reverse of sorted — because Python iterates
// the loaded dict in document order. This was confirmed against the live CLI by
// swapping the two lines in a copy of the lock, which swapped the two [FAIL]
// lines.
func TestGate8FindingOrderMatchesGolden(t *testing.T) {
	root := repoRoot(t)
	pairs := lockFilesPairs(t, filepath.Join(root, "examples", "failing-federated", "workspace.lock.yaml"))

	golden, err := os.ReadFile(filepath.Join(root, "examples", "failing-federated-golden-validate.txt"))
	if err != nil {
		t.Skipf("golden unavailable: %v", err)
	}

	// The per-file findings are the [FAIL] lines naming a path from the lock,
	// in the order the golden prints them.
	var goldenOrder []string
	for _, line := range strings.Split(string(golden), "\n") {
		if !strings.Contains(line, "[FAIL]") {
			continue
		}
		for _, p := range pairs {
			if strings.Contains(line, p.Key) {
				goldenOrder = append(goldenOrder, p.Key)
				break
			}
		}
	}
	if len(goldenOrder) < 2 {
		t.Fatalf("golden yielded %d per-file findings; expected the fixture's two", len(goldenOrder))
	}

	for i := 0; i < stabilityRuns; i++ {
		var got []string
		for _, p := range MapPairs(Deref(MapGet(
			Deref(MapGet(loadFixture(t, filepath.Join(root, "examples", "failing-federated", "workspace.lock.yaml")).Root(), "repos")).Content[0],
			"files"))) {
			got = append(got, p.Key)
		}
		if !reflect.DeepEqual(got, goldenOrder) {
			t.Fatalf("run %d: MapPairs order %v does not match golden [FAIL] order %v", i, got, goldenOrder)
		}
	}

	// Teeth: sorted order must differ from the golden, or a sorting
	// implementation would pass this test.
	sorted := append([]string(nil), goldenOrder...)
	sort.Strings(sorted)
	if reflect.DeepEqual(sorted, goldenOrder) {
		t.Fatal("golden finding order is alphabetical — this test no longer distinguishes " +
			"document order from sorted order")
	}
}

// ---------------------------------------------------------------- MapPairs

func TestMapPairsPreservesAuthoredOrder(t *testing.T) {
	doc, err := Load([]byte("zebra: 1\nalpha: 2\nmiddle: 3\n"))
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	want := []string{"zebra", "alpha", "middle"}
	for i := 0; i < stabilityRuns; i++ {
		var got []string
		for _, p := range MapPairs(doc.Root()) {
			got = append(got, p.Key)
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("run %d: got %v, want %v", i, got, want)
		}
	}
}

func TestMapPairsNonMapping(t *testing.T) {
	for name, src := range map[string]string{"sequence": "- a\n- b\n", "scalar": "hello\n"} {
		doc, err := Load([]byte(src))
		if err != nil {
			t.Fatalf("%s: load: %v", name, err)
		}
		if got := MapPairs(doc.Root()); got != nil {
			t.Errorf("%s: MapPairs = %v, want nil", name, got)
		}
	}
	if got := MapPairs(nil); got != nil {
		t.Errorf("MapPairs(nil) = %v, want nil", got)
	}
}

// MapPairs must agree with MapKeys/MapGet, the pre-1.4 way of doing this.
func TestMapPairsAgreesWithMapKeys(t *testing.T) {
	doc, err := Load([]byte("b: 1\na: 2\nc: 3\n"))
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	keys := MapKeys(doc.Root())
	pairs := MapPairs(doc.Root())
	if len(keys) != len(pairs) {
		t.Fatalf("MapKeys %d pairs, MapPairs %d", len(keys), len(pairs))
	}
	for i, k := range keys {
		if pairs[i].Key != k {
			t.Errorf("pair %d key = %q, want %q", i, pairs[i].Key, k)
		}
		if pairs[i].Value != MapGet(doc.Root(), k) {
			t.Errorf("pair %d value node differs from MapGet(%q)", i, k)
		}
	}
}

// --------------------------------------------------------------- OrderedMap

// Python dict semantics, asserted directly: re-assigning an existing key keeps
// its ORIGINAL position. Getting this wrong moves a file to the end of
// workspace.lock.yaml when two slices contribute it.
func TestOrderedMapSetKeepsFirstPosition(t *testing.T) {
	om := NewOrderedMap()
	om.SetString("first", "1")
	om.SetString("second", "2")
	om.SetString("first", "rewritten")

	if got, want := om.Keys(), []string{"first", "second"}; !reflect.DeepEqual(got, want) {
		t.Errorf("keys = %v, want %v", got, want)
	}
	v, ok := om.Get("first")
	if !ok || v.Value != "rewritten" {
		t.Errorf("Get(first) = %v/%v, want rewritten", v, ok)
	}
	if om.Len() != 2 {
		t.Errorf("Len = %d, want 2", om.Len())
	}
}

// dict.update: present keys keep position and take the new value; absent keys
// append in the source's order. bin/company-os:2471.
func TestOrderedMapUpdateSemantics(t *testing.T) {
	base := NewOrderedMap()
	base.SetString("a", "1")
	base.SetString("b", "2")

	other := NewOrderedMap()
	other.SetString("b", "two")
	other.SetString("c", "3")
	other.SetString("a", "one")

	base.Update(other)
	if got, want := base.Keys(), []string{"a", "b", "c"}; !reflect.DeepEqual(got, want) {
		t.Errorf("keys after update = %v, want %v", got, want)
	}
	for k, want := range map[string]string{"a": "one", "b": "two", "c": "3"} {
		v, ok := base.Get(k)
		if !ok || v.Value != want {
			t.Errorf("Get(%q) = %v, want %q", k, v, want)
		}
	}
	base.Update(nil) // must not panic
}

// Keys and SortedKeys answer two different questions over one key set: the lock
// EMITS in walk order (Keys) while aggregate_hash digests in plain string order
// (SortedKeys). Both must be stable, and they must not be the same list.
func TestOrderedMapKeysVsSortedKeys(t *testing.T) {
	om := NewOrderedMap()
	for _, k := range []string{"governance/req.yaml", "components/svc.yaml", "reality/x.md"} {
		om.SetString(k, "sha")
	}
	insertion := []string{"governance/req.yaml", "components/svc.yaml", "reality/x.md"}
	sorted := []string{"components/svc.yaml", "governance/req.yaml", "reality/x.md"}

	for i := 0; i < stabilityRuns; i++ {
		if got := om.Keys(); !reflect.DeepEqual(got, insertion) {
			t.Fatalf("run %d: Keys = %v, want %v", i, got, insertion)
		}
		if got := om.SortedKeys(); !reflect.DeepEqual(got, sorted) {
			t.Fatalf("run %d: SortedKeys = %v, want %v", i, got, sorted)
		}
	}
	// Callers must not be able to corrupt the map through a returned slice.
	om.Keys()[0] = "mutated"
	if om.Keys()[0] != "governance/req.yaml" {
		t.Error("Keys returned an aliased slice")
	}
}

func TestOrderedMapNodeIsFreshEachCall(t *testing.T) {
	om := NewOrderedMap()
	om.SetString("k", "v")
	if a, b := om.Node(), om.Node(); a == b {
		t.Error("Node returned the same node twice; it must build a fresh tree")
	}
}

// The zero OrderedMap must work — a struct field declared without the
// constructor is an easy way to reach a nil index map.
func TestOrderedMapZeroValue(t *testing.T) {
	var om OrderedMap
	om.SetString("a", "1")
	om.SetString("a", "2")
	if got, want := om.Keys(), []string{"a"}; !reflect.DeepEqual(got, want) {
		t.Errorf("keys = %v, want %v", got, want)
	}
}

// ------------------------------------------------------------------ PathLess

// pathOrderCorpus is the frozen truth table: each group is a set of paths and
// the order CPython's sorted(PurePosixPath) puts them in. Recorded from Python
// 3.12.11; TestPathLessAgainstPythonOracle re-derives it live.
var pathOrderCorpus = [][]string{
	// The separator-vs-punctuation case, which is the whole reason sort.Strings
	// is not a substitute: '/' is 0x2F, so '-' and '.' sort BEFORE it as bytes
	// but AFTER it component-wise.
	{"sdd/adr", "sdd/adr/a.md", "sdd/adr/z.md", "sdd/adr-x.md", "sdd/adr.md", "sdd/overview.md"},
	{"a/b", "a/b/c", "a-b", "a.b", "a0b"},
	// A prefix sorts before any extension of it.
	{"x", "x/y", "x/y/z"},
	// Depth does not win on its own; the first differing component does.
	{"b/a/a/a", "c"},
	// Real lock keys.
	{
		"platforms/communications/components/customer-notification-service.yaml",
		"platforms/communications/governance/requirements.yaml",
		"platforms/communications/reality/components/customer-notification-service.md",
		"platforms/communications/skills/creating-prd.SKILL.md",
	},
	// Absolute paths: rglob yields these, and the leading separator is its own
	// component in CPython.
	{"/a/b", "/a/b/c", "/a/b-c"},
	// Case is significant and uppercase sorts first, as in ASCII.
	{"A.md", "B.md", "a.md", "b.md"},
	// Digits before letters, and numeric strings sort lexically not numerically.
	{"v1/a", "v10/a", "v2/a"},
}

func TestPathLessFrozenTruthTable(t *testing.T) {
	for i, want := range pathOrderCorpus {
		for run := 0; run < 4; run++ {
			got := shuffled(want, int64(i*100+run))
			SortPaths(got)
			if !reflect.DeepEqual(got, want) {
				t.Errorf("group %d run %d: SortPaths = %v, want %v", i, run, got, want)
			}
		}
	}
}

// The comparator must actually differ from sort.Strings somewhere in the corpus,
// otherwise it is untested weight.
func TestPathLessDiffersFromStringSort(t *testing.T) {
	differs := 0
	for _, group := range pathOrderCorpus {
		asStrings := append([]string(nil), group...)
		sort.Strings(asStrings)
		if !reflect.DeepEqual(asStrings, group) {
			differs++
		}
	}
	if differs == 0 {
		t.Fatal("no corpus group distinguishes PathLess from sort.Strings")
	}
	t.Logf("%d/%d groups distinguish PathLess from sort.Strings", differs, len(pathOrderCorpus))
}

// PathLess must be a strict total order, or sort.SliceStable's result depends on
// input order and R-0.11 fails for a reason no fixture would show.
func TestPathLessIsATotalOrder(t *testing.T) {
	var all []string
	for _, g := range pathOrderCorpus {
		all = append(all, g...)
	}
	for _, a := range all {
		if PathLess(a, a) {
			t.Errorf("PathLess(%q, %q) is true; must be irreflexive", a, a)
		}
		for _, b := range all {
			if a == b {
				continue
			}
			if PathLess(a, b) == PathLess(b, a) {
				t.Errorf("PathLess(%q,%q) == PathLess(%q,%q); not a total order", a, b, b, a)
			}
			for _, c := range all {
				if PathLess(a, b) && PathLess(b, c) && !PathLess(a, c) {
					t.Errorf("not transitive: %q < %q < %q but not %q < %q", a, b, c, a, c)
				}
			}
		}
	}
}

// TestPathLessAgainstPythonOracle re-derives the ordering live from CPython over
// the frozen corpus plus a randomly generated one, so a divergence fails here
// rather than being discovered in a lock diff. Skips, never passes, when Python
// is unavailable.
func TestPathLessAgainstPythonOracle(t *testing.T) {
	oracle := filepath.Join("testdata", "pathorder_oracle.py")
	if _, err := os.Stat(oracle); err != nil {
		t.Skipf("oracle unavailable: %v", err)
	}

	groups := append([][]string(nil), pathOrderCorpus...)
	groups = append(groups, randomPathGroups(200, 6)...)

	in, err := json.Marshal(groups)
	if err != nil {
		t.Fatalf("marshal corpus: %v", err)
	}
	cmd := exec.Command("python3", oracle)
	cmd.Stdin = bytes.NewReader(in)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		t.Skipf("python3 oracle unavailable: %v (%s)", err, stderr.String())
	}

	var resp struct {
		Python string     `json:"python"`
		Sorted [][]string `json:"sorted"`
	}
	if err := json.Unmarshal(out, &resp); err != nil {
		t.Fatalf("decode oracle output: %v", err)
	}
	if len(resp.Sorted) != len(groups) {
		t.Fatalf("oracle returned %d groups, sent %d", len(resp.Sorted), len(groups))
	}

	mismatches := 0
	for i, want := range resp.Sorted {
		got := append([]string(nil), groups[i]...)
		SortPaths(got)
		if !reflect.DeepEqual(got, want) {
			mismatches++
			if mismatches <= 5 {
				t.Errorf("group %d: SortPaths = %v, CPython = %v", i, got, want)
			}
		}
	}
	t.Logf("PathLess agrees with CPython %s on %d/%d groups",
		resp.Python, len(groups)-mismatches, len(groups))
}

// randomPathGroups builds path sets biased toward the characters that surround
// '/' in ASCII, which is where component-wise and byte-wise ordering disagree.
func randomPathGroups(n, size int) [][]string {
	rng := rand.New(rand.NewSource(1404))
	segs := []string{"a", "b", "ab", "a-b", "a.b", "a_b", "A", "z", "0", "10", "2", "a-", "a."}
	groups := make([][]string, 0, n)
	for i := 0; i < n; i++ {
		seen := map[string]bool{}
		var group []string
		for len(group) < size {
			depth := 1 + rng.Intn(3)
			parts := make([]string, depth)
			for d := range parts {
				parts[d] = segs[rng.Intn(len(segs))]
			}
			p := strings.Join(parts, "/")
			if !seen[p] {
				seen[p] = true
				group = append(group, p)
			}
		}
		groups = append(groups, group)
	}
	return groups
}

func shuffled(in []string, seed int64) []string {
	out := append([]string(nil), in...)
	rng := rand.New(rand.NewSource(seed))
	rng.Shuffle(len(out), func(i, j int) { out[i], out[j] = out[j], out[i] })
	return out
}

// ------------------------------------------------------- end-to-end stability

// TestFullLockEmitIsByteStable is R-0.11's headline claim on the real artifact:
// load examples/federated/workspace.lock.yaml, re-emit it stabilityRuns times,
// and require identical bytes every run. Every mapping in that document — repos,
// pin, and the files map itself — travels through the node tree, so a Go map
// leaking into any of them would show up here.
func TestFullLockEmitIsByteStable(t *testing.T) {
	root := repoRoot(t)
	for _, fixture := range []string{
		filepath.Join("examples", "federated", "workspace.lock.yaml"),
		filepath.Join("examples", "failing-federated", "workspace.lock.yaml"),
	} {
		t.Run(fixture, func(t *testing.T) {
			var first []byte
			for i := 0; i < stabilityRuns; i++ {
				got, err := loadFixture(t, filepath.Join(root, fixture)).Bytes()
				if err != nil {
					t.Fatalf("run %d: emit: %v", i, err)
				}
				if i == 0 {
					first = got
					continue
				}
				if !bytes.Equal(first, got) {
					t.Fatalf("run %d differs from run 0\n%s", i, diffFirstLine(first, got))
				}
			}
		})
	}
}

func diffFirstLine(a, b []byte) string {
	la, lb := strings.Split(string(a), "\n"), strings.Split(string(b), "\n")
	for i := range la {
		if i >= len(lb) || la[i] != lb[i] {
			return fmt.Sprintf("line %d: %q vs %q", i+1, la[i], safeIdx(lb, i))
		}
	}
	return "length differs"
}

func safeIdx(s []string, i int) string {
	if i < len(s) {
		return s[i]
	}
	return "<missing>"
}
