package scaffold

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/model"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/workspace"
)

// register_id is the one read-modify-write path `add` and `init` share, and it
// re-dumps the WHOLE registry. What it does with a registry of the wrong SHAPE
// is therefore a filesystem question, not a message question: Python raises
// before dump_yaml runs, so the file on disk is untouched. R-0.7a(j) carves out
// the traceback and the exit code, and explicitly does NOT carve out the
// filesystem effect.
//
// Measured with the vendored PyYAML 6.0.2 on a workspace scaffolded by `init`:
//
//	ids:\n- justastring   → AttributeError: 'str' object has no attribute 'get'
//	                        exit 1, registry byte-unchanged, platforms/zzz written
//	[]                     → falsy, so `or {default}` fires: exit 0, registry
//	                        replaced by the default IdRegistry document
//
// Go's `if m, ok := e.(pyMap); ok` skipped the bad entry and rewrote the file;
// its `len(data) == 0` guard missed the falsy `[]` and exited 3.
func registryPath(root string) string {
	return filepath.Join(root, "company-ontology", "ids", "registry.yaml")
}

func writeRegistry(t *testing.T, root, body string) {
	t.Helper()
	if err := os.WriteFile(registryPath(root), []byte(body), 0o666); err != nil {
		t.Fatal(err)
	}
}

// TestRegisterIDRefusesAMalformedEntryAndWritesNothing is P3.
func TestRegisterIDRefusesAMalformedEntryAndWritesNothing(t *testing.T) {
	root := initWorkspace(t)
	const body = "schemaVersion: '1.0'\nkind: IdRegistry\nids:\n- justastring\n" +
		"tags:\n- ontology/registry\n"
	writeRegistry(t, root, body)

	_, err := Add(workspace.New(root), AddPlatform, "zzz", "", nil)
	if err == nil {
		t.Fatal("expected a refusal, got nil")
	}
	if code := model.CodeOf(err); code != model.ExitArtifact {
		t.Errorf("exit code = %d, want %d (R-0.7a(j))", code, model.ExitArtifact)
	}
	if got := read(t, registryPath(root)); got != body {
		t.Errorf("registry was rewritten; Python writes nothing here\n--- got\n%s--- want\n%s", got, body)
	}
}

// TestRegisterIDRefusesAMalformedEntryOnlyWhenItIsReached mirrors the fact that
// `any(... for e in ids)` is a GENERATOR: a matching entry ahead of the bad one
// short-circuits before `.get` is ever called on it, and Python exits 0.
func TestRegisterIDRefusesAMalformedEntryOnlyWhenItIsReached(t *testing.T) {
	root := initWorkspace(t)
	writeRegistry(t, root, "schemaVersion: '1.0'\nkind: IdRegistry\nids:\n"+
		"- {id: 'platform://zzz', definedIn: platforms/zzz/platform.yaml}\n"+
		"- justastring\ntags:\n- ontology/registry\n")

	if _, err := Add(workspace.New(root), AddPlatform, "zzz", "", nil); err != nil {
		t.Fatalf("expected the already-registered id to short-circuit, got %v", err)
	}
}

// TestRegisterIDFalsyRegistryFallsBackToTheDefault is P4: `or {default}` is
// truthiness, so `[]`, `{}`, `0` and `”` all take the default branch.
func TestRegisterIDFalsyRegistryFallsBackToTheDefault(t *testing.T) {
	for _, body := range []string{"[]\n", "{}\n", "0\n", "''\n", "null\n", "false\n"} {
		t.Run(strings.TrimSpace(body), func(t *testing.T) {
			root := initWorkspace(t)
			writeRegistry(t, root, body)

			if _, err := Add(workspace.New(root), AddPlatform, "zzz", "", nil); err != nil {
				t.Fatalf("Add: %v", err)
			}
			want := "schemaVersion: '1.0'\nkind: IdRegistry\nids:\n" +
				"- id: platform://zzz\n  definedIn: platforms/zzz/platform.yaml\n" +
				"tags:\n- ontology/registry\n"
			if got := read(t, registryPath(root)); got != want {
				t.Fatalf("got:\n%s\nwant:\n%s", got, want)
			}
		})
	}
}

// TestRegisterIDRefusesANonMappingRootAndWritesNothing is the other half of
// R-0.7a(j): a TRUTHY document of the wrong root shape, where `data.setdefault`
// raises.
func TestRegisterIDRefusesANonMappingRootAndWritesNothing(t *testing.T) {
	for _, body := range []string{"- a\n- b\n", "hello\n", "42\n"} {
		t.Run(strings.TrimSpace(body), func(t *testing.T) {
			root := initWorkspace(t)
			writeRegistry(t, root, body)

			_, err := Add(workspace.New(root), AddPlatform, "zzz", "", nil)
			if err == nil {
				t.Fatal("expected a refusal, got nil")
			}
			if code := model.CodeOf(err); code != model.ExitArtifact {
				t.Errorf("exit code = %d, want %d", code, model.ExitArtifact)
			}
			if got := read(t, registryPath(root)); got != body {
				t.Errorf("registry was rewritten; Python writes nothing here\ngot:\n%s", got)
			}
		})
	}
}

// TestRegisterIDRefusesANonListIdsAndWritesNothing covers `ids:` holding a
// mapping or a scalar, where `for e in ids` raises or iterates the wrong thing.
func TestRegisterIDRefusesANonListIdsAndWritesNothing(t *testing.T) {
	root := initWorkspace(t)
	const body = "schemaVersion: '1.0'\nkind: IdRegistry\nids: notalist\n"
	writeRegistry(t, root, body)

	if _, err := Add(workspace.New(root), AddPlatform, "zzz", "", nil); err == nil {
		t.Fatal("expected a refusal, got nil")
	}
	if got := read(t, registryPath(root)); got != body {
		t.Errorf("registry was rewritten; got:\n%s", got)
	}
}
