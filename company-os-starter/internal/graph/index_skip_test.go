package graph_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/graph"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/workspace"
)

// TestIterGraphDocsSkipsIndexFiles is R-2.5 of okf-provenance-and-indexes.
//
// A generated index.md carries frontmatter, so without a by-name skip
// IterGraphDocs would yield it and four things would go wrong at once. The
// sharpest is the fourth: CoreFieldErrors demands `id` or `prd`, a generated
// index has neither, so gate 4 would fail on a file the tool itself wrote — the
// D-2.2 failure in a second guise.
//
//  1. it is ingested as a graph document
//  2. it counts toward its own directory's >=2 threshold (D-2.3), making a
//     one-document directory eligible the moment an index appears in it
//  3. it is listed inside itself
//  4. it fails gate 4's core-field contract
//
// The index.md written here carries FULL, VALID frontmatter on purpose. The skip
// is by base name (`tags.go:222`), so valid frontmatter must not rescue it — a
// test using a malformed index would pass even if the skip were removed and the
// file were merely being dropped for being unparseable.
//
// The three incumbent skips are asserted alongside, so an edit to skipNames
// cannot silently drop one while adding index.md.
//
// @spec req://uncle-os/okf-provenance-and-indexes@0.1#R-2.5
func TestIterGraphDocsSkipsIndexFiles(t *testing.T) {
	ws := workspace.New(t.TempDir())
	dir := filepath.Join(ws.Company, "standards")
	if err := os.MkdirAll(dir, 0o777); err != nil {
		t.Fatal(err)
	}

	// One real graph document, so a zero-length result cannot be mistaken for
	// a passing skip.
	write(t, filepath.Join(dir, "real.md"),
		"---\ntype: adr\nid: adr-001\nstatus: draft\n---\n\n# Real\n")

	// Every skipped name, each carrying frontmatter valid enough to be ingested
	// if the skip were not by name.
	for _, name := range []string{"index.md", "log.md", "README.md", "CLAUDE.md"} {
		write(t, filepath.Join(dir, name),
			"---\ntype: adr\nid: skipped-"+name+"\nstatus: draft\n---\n\n# Skipped\n")
	}

	docs, err := graph.IterGraphDocs(ws)
	if err != nil {
		t.Fatalf("IterGraphDocs: %v", err)
	}

	got := map[string]bool{}
	for _, d := range docs {
		got[filepath.Base(d.Path)] = true
	}
	if !got["real.md"] {
		t.Fatalf("real.md was not ingested; the fixture is wrong, not the skip. got %v", got)
	}
	for _, name := range []string{"index.md", "log.md", "README.md", "CLAUDE.md"} {
		if got[name] {
			t.Errorf("%s was ingested as a graph document; it must be skipped by name", name)
		}
	}
}

func write(t *testing.T, path, body string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(body), 0o666); err != nil {
		t.Fatal(err)
	}
}
