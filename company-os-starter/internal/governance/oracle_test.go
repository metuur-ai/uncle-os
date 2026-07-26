package governance_test

// The differential harness compares two BINARIES over a corpus of invocations.
// This file makes the same comparison at the LIBRARY seam, against
// bin/company-os, over the committed fixtures — so a divergence fails `go test`
// instead of waiting for someone to run the harness.
//
// It carries the two claims task 2.5 exists to make:
//
//   - `deviation declare` and `exception request` reproduce the oracle's
//     read-modify-write BYTE FOR BYTE. R-0.7a(g) sanctions an authored file
//     coming back out under a different layout; this asserts the carve-out is
//     not needed, because PyLoadFile/PyWriteFile is PyYAML's own reflow. If the
//     emitter ever drifts, these fail here rather than in the harness.
//   - `governance resolve` writes an effective-governance.yaml that agrees with
//     the oracle's on everything except generatedAt, which is the R-0.7c guard's
//     whole subject.

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/governance"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/model"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/render"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/workspace"
)

// tsRE is the harness's only content normalization that matters here
// (examples/differential.py:106): the generated UTC timestamp.
var tsRE = regexp.MustCompile(`\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z`)

func repoRoot(t *testing.T) string {
	t.Helper()
	abs, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatalf("resolving repo root: %v", err)
	}
	return abs
}

// oracleCLI runs the reference CLI, skipping rather than failing when no oracle
// is available — a missing oracle must never look like agreement.
func oracleCLI(t *testing.T, wsPath string, args ...string) (stdout string, code int) {
	t.Helper()
	root := repoRoot(t)
	cli := filepath.Join(root, "company-os-starter", "bin", "company-os")
	if _, err := os.Stat(cli); err != nil {
		t.Skipf("reference CLI not present: %v", err)
	}
	if _, err := exec.LookPath("python3"); err != nil {
		t.Skip("python3 not available; the reference oracle cannot run")
	}
	cmd := exec.Command("python3", append([]string{cli, "--root", wsPath}, args...)...)
	cmd.Env = append(os.Environ(),
		"PYTHONPATH="+filepath.Join(root, "company-os-starter", "vendor"))
	var out, errb bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errb
	err := cmd.Run()
	if err != nil {
		var ee *exec.ExitError
		if !asExit(err, &ee) {
			t.Fatalf("reference CLI failed on %v: %v\n%s", args, err, errb.String())
		}
		code = ee.ExitCode()
	}
	return out.String(), code
}

func asExit(err error, target **exec.ExitError) bool {
	ee, ok := err.(*exec.ExitError)
	if ok {
		*target = ee
	}
	return ok
}

// copyFixture copies a committed fixture so the oracle and the port each get a
// pristine tree, exactly as the harness does.
func copyFixture(t *testing.T, name string) string {
	t.Helper()
	src := filepath.Join(repoRoot(t), "examples", filepath.FromSlash(name))
	if _, err := os.Stat(src); err != nil {
		t.Skipf("fixture absent: %v", err)
	}
	dst := filepath.Join(t.TempDir(), "ws")
	err := filepath.WalkDir(src, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, p)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		if !d.Type().IsRegular() {
			return nil
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		return os.WriteFile(target, b, 0o644)
	})
	if err != nil {
		t.Fatalf("copying %s: %v", src, err)
	}
	return dst
}

func renderSections(t *testing.T, sections []model.GateResult) string {
	t.Helper()
	var b bytes.Buffer
	if err := render.Governance(&b, sections); err != nil {
		t.Fatalf("render.Governance: %v", err)
	}
	return b.String()
}

// TestDeclareReproducesTheOracleBytes is the task's headline claim: the
// read-modify-write of an AUTHORED file comes out byte-identical.
//
// The workspace fixture is the demanding one. Its deviations.yaml is committed
// in FLOW style with a folded multi-line rationale, an unquoted YAML-1.1
// timestamp (`reviewDate: 2035-01-15`), a quoted one on the next entry, and a
// `://` scalar inside a flow mapping — four independent emitter behaviours, all
// of which safe_dump rewrites. Reproducing "the file the oracle leaves behind"
// therefore means reproducing a reflow, not preserving the author's layout.
func TestDeclareReproducesTheOracleBytes(t *testing.T) {
	cases := []struct {
		fixture, team, rule, rationale string
	}{
		{"workspace", "customer-engagement",
			"platform-standard://communications/prd-structure", ""},
		{"workspace", "customer-engagement",
			"platform-standard://communications/delivery-reliability", "we ship a different shape"},
		{"standalone-team", "solo", "company-control://change-log", ""},
		{"banking/small-company", "core", "company-control://change-log", ""},
		{"federated", "customer-engagement",
			"platform-standard://communications/prd-structure", ""},
	}
	for _, c := range cases {
		t.Run(c.fixture+"/"+c.team, func(t *testing.T) {
			rel := filepath.Join("teams", c.team, "governance", "deviations.yaml")

			pyWS := copyFixture(t, c.fixture)
			args := []string{"deviation", "declare", c.rule, "--team", c.team}
			if c.rationale != "" {
				args = append(args, "--rationale", c.rationale)
			}
			pyOut, pyCode := oracleCLI(t, pyWS, args...)
			if pyCode != 0 {
				t.Fatalf("oracle exited %d", pyCode)
			}

			goWS := copyFixture(t, c.fixture)
			sections, err := governance.DeclareSections(
				workspace.New(goWS), c.team, c.rule, c.rationale)
			if err != nil {
				t.Fatalf("DeclareSections: %v", err)
			}
			if got := renderSections(t, sections); got != pyOut {
				t.Errorf("stdout diverges\n go: %q\npy: %q", got, pyOut)
			}
			assertSameBytes(t, filepath.Join(pyWS, rel), filepath.Join(goWS, rel))
		})
	}
}

// TestRequestReproducesTheOracleBytes is the same claim for exceptions.yaml,
// which is committed in BLOCK style — so it also proves the emitter does not
// gratuitously flow a document that arrived block.
func TestRequestReproducesTheOracleBytes(t *testing.T) {
	cases := []struct {
		fixture, team, rule, component, reason, expires string
	}{
		{"workspace", "customer-engagement",
			"platform-standard://communications/delivery-reliability",
			"customer-notification-service", "", "2035-01-01"},
		{"workspace", "customer-engagement",
			"platform-standard://communications/message-schema",
			"customer-notification-service", "legacy consumer", "not-a-date"},
		{"standalone-team", "solo", "company-control://security-service-baseline",
			"none", "", "2035-01-01"},
		{"banking/small-company", "core", "platform-standard://product/release-safety",
			"banking-app", "", "2035-01-01"},
	}
	for _, c := range cases {
		t.Run(c.fixture+"/"+c.rule, func(t *testing.T) {
			rel := filepath.Join("teams", c.team, "governance", "exceptions.yaml")

			pyWS := copyFixture(t, c.fixture)
			args := []string{"exception", "request", c.rule, "--team", c.team,
				"--component", c.component, "--expires", c.expires}
			if c.reason != "" {
				args = append(args, "--reason", c.reason)
			}
			pyOut, pyCode := oracleCLI(t, pyWS, args...)
			if pyCode != 0 {
				t.Fatalf("oracle exited %d", pyCode)
			}

			goWS := copyFixture(t, c.fixture)
			sections, err := governance.RequestSections(
				workspace.New(goWS), c.team, c.rule, c.component, c.reason, c.expires)
			if err != nil {
				t.Fatalf("RequestSections: %v", err)
			}
			if got := renderSections(t, sections); got != pyOut {
				t.Errorf("stdout diverges\n go: %q\npy: %q", got, pyOut)
			}
			assertSameBytes(t, filepath.Join(pyWS, rel), filepath.Join(goWS, rel))
		})
	}
}

// TestResolveMatchesTheOracle covers every fixture the differential corpus
// resolves, on stdout and on the generated artifact.
//
// The artifact is compared with the timestamp normalized, which is exactly the
// harness's rule and exactly the field the R-0.7c guard holds equal. Everything
// else must agree byte for byte.
func TestResolveMatchesTheOracle(t *testing.T) {
	cases := []struct{ fixture, team string }{
		{"workspace", "customer-engagement"},
		{"standalone-team", "solo"},
		{"federated", "customer-engagement"},
		{"banking/small-company", "core"},
		{"banking/bank/workspaces/team-payments-rails", "payments-rails"},
		{"banking/bank/workspaces/team-fraud-detection", "fraud-detection"},
		{"failing-workspace", "ghost"},
	}
	for _, c := range cases {
		t.Run(c.fixture, func(t *testing.T) {
			rel := filepath.Join("teams", c.team, "generated", "effective-governance.yaml")

			pyWS := copyFixture(t, c.fixture)
			pyOut, pyCode := oracleCLI(t, pyWS, "governance", "resolve", "--team", c.team)
			if pyCode != 0 {
				t.Fatalf("oracle exited %d", pyCode)
			}

			goWS := copyFixture(t, c.fixture)
			sections, err := governance.ResolveSections(workspace.New(goWS), c.team)
			if err != nil {
				t.Fatalf("ResolveSections: %v", err)
			}
			if got := renderSections(t, sections); got != pyOut {
				t.Errorf("stdout diverges\n go: %q\npy: %q", got, pyOut)
			}
			want := tsRE.ReplaceAllString(readFile(t, filepath.Join(pyWS, rel)), "<TS>")
			got := tsRE.ReplaceAllString(readFile(t, filepath.Join(goWS, rel)), "<TS>")
			if got != want {
				t.Errorf("%s diverges\n go:\n%s\npy:\n%s", rel, got, want)
			}
		})
	}
}

// TestExplainMatchesTheOracle covers the read side, including the difflib
// suggestion path — `governance explain` is internal/ids' Suggest's first live
// caller (`bin/company-os:365`).
func TestExplainMatchesTheOracle(t *testing.T) {
	cases := []struct {
		name, fixture, component string
		resolveFirst             string
	}{
		{"known", "workspace", "customer-notification-service", ""},
		{"unknown-with-suggestions", "workspace", "customer-notification-servic", ""},
		{"unknown-no-suggestions", "workspace", "zzzzzzzz", ""},
		{"omitted-positional", "workspace", "None", ""},
		{"banking", "banking/small-company", "banking-app", ""},
		{"after-resolve", "standalone-team", "anything", "solo"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			pyWS := copyFixture(t, c.fixture)
			goWS := copyFixture(t, c.fixture)
			if c.resolveFirst != "" {
				if _, code := oracleCLI(t, pyWS, "governance", "resolve", "--team", c.resolveFirst); code != 0 {
					t.Fatalf("oracle resolve exited %d", code)
				}
				if _, err := governance.ResolveSections(workspace.New(goWS), c.resolveFirst); err != nil {
					t.Fatalf("ResolveSections: %v", err)
				}
			}

			// The omitted positional is argparse's None; the CLI renders it as
			// the four characters "None", so the oracle is invoked with no
			// positional at all and the port with that rendering.
			args := []string{"governance", "explain"}
			if c.name != "omitted-positional" {
				args = append(args, c.component)
			}
			pyOut, _ := oracleCLI(t, pyWS, args...)

			sections, err := governance.Explain(workspace.New(goWS), c.component)
			if err != nil {
				// The oracle prints its diagnostic on stderr and stdout is
				// empty; the port carries it in the error. Compare the message
				// against die()'s rendering minus its own prefix.
				want := oracleStderr(t, pyWS, args...)
				if got := "error: " + err.Error() + "\n"; got != want {
					t.Errorf("diagnostic diverges\n go: %q\npy: %q", got, want)
				}
				if pyOut != "" {
					t.Errorf("oracle printed on stdout before dying: %q", pyOut)
				}
				return
			}
			if got := renderSections(t, sections); got != pyOut {
				t.Errorf("stdout diverges\n go: %q\npy: %q", got, pyOut)
			}
		})
	}
}

// oracleStderr is oracleCLI's other stream, for the paths where the reference
// dies.
func oracleStderr(t *testing.T, wsPath string, args ...string) string {
	t.Helper()
	root := repoRoot(t)
	cli := filepath.Join(root, "company-os-starter", "bin", "company-os")
	cmd := exec.Command("python3", append([]string{cli, "--root", wsPath}, args...)...)
	cmd.Env = append(os.Environ(),
		"PYTHONPATH="+filepath.Join(root, "company-os-starter", "vendor"))
	var errb bytes.Buffer
	cmd.Stderr = &errb
	if err := cmd.Run(); err == nil {
		t.Fatalf("oracle %v succeeded; it no longer reaches the failure path", args)
	}
	return errb.String()
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading %s: %v", path, err)
	}
	return string(b)
}

func assertSameBytes(t *testing.T, want, got string) {
	t.Helper()
	a, b := readFile(t, want), readFile(t, got)
	if a == b {
		return
	}
	t.Errorf("the read-modify-write diverges from the oracle\n"+
		"py:\n%s\ngo:\n%s\nfirst difference at %d", a, b, commonPrefix(a, b))
}

func commonPrefix(a, b string) int {
	n := 0
	for n < len(a) && n < len(b) && a[n] == b[n] {
		n++
	}
	return n
}
