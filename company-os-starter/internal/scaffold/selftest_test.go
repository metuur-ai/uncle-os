package scaffold

// Inherited from examples/selftest.py (task 6.1). ST-030, `:170-172`.
//
// The Python assertion was `txt == co.PRD_TEMPLATE and "PRD_TEMPLATE" in src`.
// The first half is covered twice over by TestResolveTemplateOverridePrecedence
// (last step) and TestResolveTemplateBuiltinsWithoutWorkspace, both of which
// compare the resolved body against the builtin constant.
//
// The second half is not, and cannot be covered by those tests: they assert
// `source == SourceBuiltinPRD`, which moves whenever the constant moves. The
// invariant selftest.py actually pinned is that the label a user SEES names the
// builtin it came from — `prd new` prints "  template: <source>", and a label
// that degraded to "(builtin)" would leave a reader unable to tell which of the
// three builtins was used, or that it was a builtin at all rather than an
// override they forgot they wrote. That is asserted against a literal here,
// deliberately not against the constant.

import (
	"strings"
	"testing"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/workspace"
)

// TestBuiltinTemplateSourceLabelsNameTheirBuiltin is selftest.py:171 (ST-030),
// generalized to all three builtins because the reasoning is identical for each
// and only PRD happened to be asserted in Python.
func TestBuiltinTemplateSourceLabelsNameTheirBuiltin(t *testing.T) {
	ws := workspace.New(t.TempDir())
	cases := []struct {
		template  string
		wantInSrc string
	}{
		{TemplatePRD, "PRD_TEMPLATE"},
		{TemplateDiscoveryBrief, "DISCOVERY_TEMPLATE"},
		{TemplateRealityComponent, "templates/reality-component.md"},
	}
	for _, tc := range cases {
		t.Run(tc.template, func(t *testing.T) {
			body, source, err := ResolveTemplate(ws, tc.template, "engineering", "core")
			if err != nil {
				t.Fatalf("ResolveTemplate: %v", err)
			}
			if !strings.Contains(source, tc.wantInSrc) {
				t.Errorf("source label %q does not name its builtin (%q)", source, tc.wantInSrc)
			}
			// R-0.8: the label is byte-frozen output, and "built-in " is the
			// prefix `prd new` / `discover new` / `reality new` print.
			if !strings.HasPrefix(source, "built-in ") {
				t.Errorf("source label %q lost the \"built-in \" prefix", source)
			}
			if body == "" {
				t.Error("builtin body is empty")
			}
		})
	}
}
