package scaffold

import (
	"path/filepath"
	"testing"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/workspace"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/yamlio"
)

// TestScaffoldedArtifactsMatchPyYAML closes the loop on the documents nobody
// authored: it re-dumps what `init` and `add` actually wrote and asserts the
// bytes come back unchanged. A regression in a scaffold dict — a reordered key,
// a value that stopped being a string — fails here rather than in the corpus.
//
// The emitter's own corpus lives with the emitter, in internal/yamlio; this test
// stays here because only this package knows what the scaffolds contain.
//
// It used to shell out to the vendored PyYAML and assert
// `src == safe_dump(safe_load(src))`. R-9.3 deleted that vendor tree, leaving the
// test able only to skip. The assertion is now the same round trip through the
// Go emitter, which is not a weaker claim: internal/yamlio's
// TestEmitterMatchesPyYAML pins PyDump to safe_dump byte for byte against
// answers frozen from PyYAML 6.0.2, so "fixed point under PyDump" and "fixed
// point under safe_dump" are the same statement as long as that test passes.
// The one thing lost is independence — both halves now rest on the same emitter.
func TestScaffoldedArtifactsMatchPyYAML(t *testing.T) {
	root := initWorkspace(t)
	if _, err := Add(workspace.New(root), AddComponent, "billing-api", "platform-1", nil); err != nil {
		t.Fatalf("add component: %v", err)
	}
	for _, rel := range []string{
		"company-os/company.yaml",
		"company-os/standards/company-baseline.yaml",
		"platforms/platform-1/platform.yaml",
		"platforms/platform-1/governance/requirements.yaml",
		"platforms/platform-1/components/billing-api.yaml",
		"teams/core/team.yaml",
		"company-ontology/ids/registry.yaml",
	} {
		t.Run(rel, func(t *testing.T) {
			path := filepath.Join(root, filepath.FromSlash(rel))
			src := read(t, path)
			loaded, err := yamlio.PyLoadFile(path)
			if err != nil {
				t.Fatalf("PyLoadFile: %v", err)
			}
			redumped, err := yamlio.PyDump(loaded)
			if err != nil {
				t.Fatalf("PyDump: %v", err)
			}
			if src != redumped {
				t.Fatalf("scaffolded file is not what the emitter would write\n"+
					"--- re-dumped\n%s--- on disk\n%s", redumped, src)
			}
		})
	}
}
