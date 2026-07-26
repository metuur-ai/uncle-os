package validate_test

// The parity gate, in Go. examples/acceptance.sh diffs all five golden snapshots
// against the PYTHON CLI; this test diffs the same five against the Go one, so a
// records or renderer change is caught by `make test` rather than by the shell
// harness — and keeps being caught after R-9.3 deletes the reference.
//
// The comparison is byte-for-byte after acceptance.sh's normalize(), which
// rewrites line 1 only. Everything else, including the em dashes and the
// trailing newline, is compared as-is.

import (
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/model"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/render"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/validate"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/workspace"
)

// normalize is examples/acceptance.sh:21 —
// `sed "s#^validating workspace .*#validating workspace <WORKSPACE>#"`. Only the
// first line is portable-ized; a workspace path appearing anywhere else would be
// a divergence, not noise.
var rootLine = regexp.MustCompile(`(?m)^validating workspace .*$`)

func normalize(s string) string {
	return rootLine.ReplaceAllString(s, "validating workspace <WORKSPACE>")
}

func examplesDir(t *testing.T) string {
	t.Helper()
	dir, err := filepath.Abs(filepath.Join("..", "..", "..", "examples"))
	if err != nil {
		t.Fatalf("resolving examples/: %v", err)
	}
	return dir
}

// runValidate is the whole pipeline under test: producers, then the renderer,
// then the exit status run() would derive.
func runValidate(t *testing.T, root string) (string, model.ExitCode, []model.GateResult) {
	t.Helper()
	if _, err := os.Stat(root); err != nil {
		t.Skipf("fixture unavailable: %v", err)
	}
	sections, err := validate.Run(workspace.New(root))
	if err != nil {
		t.Fatalf("validate.Run(%s): %v", root, err)
	}
	var buf bytes.Buffer
	if err := render.Validate(&buf, sections); err != nil {
		t.Fatalf("render.Validate: %v", err)
	}
	code := model.ExitOK
	if model.HasFailure(sections) {
		code = model.ExitValidation
	}
	return normalize(buf.String()), code, sections
}

// TestGoldensReproduceByteForByte is the R-0.4/R-0.9 gate: five fixtures, five
// committed snapshots, no tolerance.
//
// The five are not redundant. golden-validate is the 7-gate all-pass path;
// federated-golden-validate is the same workspace shape with gate 8 present, and
// is what proves the [N/M] denominator is computed rather than stored;
// failing-workspace drives at least one [FAIL] through every one of gates 1-7
// plus the single warn site and gate 4's conditional [ok]; and the two federated
// failure fixtures split the five federated_slice_problems shapes, because the
// absent-lock branch returns early and therefore cannot co-occur with the other
// four.
func TestGoldensReproduceByteForByte(t *testing.T) {
	dir := examplesDir(t)
	cases := []struct {
		fixture string
		golden  string
		want    model.ExitCode
		gates   int
	}{
		{"workspace", "golden-validate.txt", model.ExitOK, 7},
		{"federated", "federated-golden-validate.txt", model.ExitOK, 8},
		{"failing-workspace", "failing-workspace-golden-validate.txt", model.ExitValidation, 7},
		{"failing-federated", "failing-federated-golden-validate.txt", model.ExitValidation, 8},
		{"failing-federated-nolock", "failing-federated-nolock-golden-validate.txt",
			model.ExitValidation, 8},
	}
	for _, tc := range cases {
		t.Run(tc.fixture, func(t *testing.T) {
			want, err := os.ReadFile(filepath.Join(dir, tc.golden))
			if err != nil {
				t.Fatalf("reading oracle: %v", err)
			}
			got, code, sections := runValidate(t, filepath.Join(dir, tc.fixture))
			if got != string(want) {
				t.Errorf("output diverges from %s\n--- got ---\n%s\n--- want ---\n%s",
					tc.golden, got, want)
			}
			// acceptance.sh asserts the status separately from the diff: a fixture
			// that silently started passing would still diff clean against a
			// re-baselined golden (`:83-90`).
			if code != tc.want {
				t.Errorf("exit code = %d, want %d", code, tc.want)
			}
			// Gate 8 exists only in federated mode, and the ordinals must be 1..N
			// with no renumbering of gates 1-7.
			gates := sections[1:]
			if len(gates) != tc.gates {
				t.Fatalf("got %d gates, want %d", len(gates), tc.gates)
			}
			for i, g := range gates {
				if g.Ordinal != i+1 {
					t.Errorf("gate %d has ordinal %d", i+1, g.Ordinal)
				}
			}
			if sections[0].Slug != model.SlugWorkspace {
				t.Errorf("section 0 is %q, want the banner", sections[0].Slug)
			}
			// The denominator is CARRIED on the banner (R-2.6a), so assert it
			// separately from the list: on a complete run the two agree, and it is
			// exactly when they stop agreeing — a mid-gate abort — that the
			// carried one is the only correct answer.
			if got := sections[0].Findings[0].Fields.Int("gates"); got != tc.gates {
				t.Errorf("banner carries gates=%d, want %d", got, tc.gates)
			}
		})
	}
}

// TestBannerCarriesTheRealRoot guards the one thing normalize() erases: line 1
// must actually name the workspace, or every golden above would pass against a
// renderer that printed a constant.
func TestBannerCarriesTheRealRoot(t *testing.T) {
	root := filepath.Join(examplesDir(t), "workspace")
	if _, err := os.Stat(root); err != nil {
		t.Skipf("fixture unavailable: %v", err)
	}
	sections, err := validate.Run(workspace.New(root))
	if err != nil {
		t.Fatalf("validate.Run: %v", err)
	}
	var buf bytes.Buffer
	if err := render.Validate(&buf, sections); err != nil {
		t.Fatalf("render.Validate: %v", err)
	}
	first, _, _ := strings.Cut(buf.String(), "\n")
	if first != "validating workspace "+root {
		t.Errorf("banner = %q, want the workspace root", first)
	}
}
