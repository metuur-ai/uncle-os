package validate_test

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestDescriptionIsInvisibleToValidate is R-1.7 of okf-provenance-and-indexes:
// a document omitting `description:` validates unchanged, producing neither an
// error nor a warning.
//
// The obvious test — assert the current fixtures pass — would be tautological.
// Every fixture omits the field today and the goldens are green, so it would
// prove that nothing was added, not that omission is durably tolerated.
//
// So this asserts SYMMETRY instead: validate's rendered output is byte-identical
// whether the field is present or absent. That covers both directions at once —
// absence draws nothing (R-1.7) and presence draws nothing either (R-1.1's
// "blocks nothing", which has no other test) — and it goes red if anyone later
// adds a gate that reads `description:` from either side.
//
// It also proves the field is not a tag source. Adding a frontmatter key to a
// committed fixture is exactly the shape that trips gate 4's committed-vs-derived
// tag comparison; validating after the insert demonstrates no drift rather than
// assuming it.
//
// The comparison is over rendered output, not findings: the renderer is what a
// user and CI see, and a warn line that never reaches stdout is not what R-1.7
// is about.
//
// @spec req://uncle-os/okf-provenance-and-indexes@0.1#R-1.7
func TestDescriptionIsInvisibleToValidate(t *testing.T) {
	ws := draftCopyFixture(t, "workspace")

	before, codeBefore, _ := runValidate(t, ws.Root)

	n := addDescriptionEverywhere(t, ws.Root)
	if n == 0 {
		t.Fatal("no typed documents found in the fixture copy — the insert " +
			"did nothing and the comparison below would be vacuous")
	}
	t.Logf("inserted description: into %d documents", n)

	after, codeAfter, _ := runValidate(t, ws.Root)

	if codeBefore != codeAfter {
		t.Errorf("exit code changed when description: was added: %v -> %v",
			codeBefore, codeAfter)
	}
	if before != after {
		t.Errorf("validate output changed when description: was added.\n"+
			"--- without ---\n%s\n--- with ---\n%s", before, after)
	}
}

// addDescriptionEverywhere inserts a `description:` line after the `type:` line
// of every markdown document in root that carries a leading frontmatter block.
// Returns how many documents were touched.
//
// CLAUDE.md is skipped: it is generated, and rewriting it by hand would make the
// gate-5 drift check the thing under test rather than description.
func addDescriptionEverywhere(t *testing.T, root string) int {
	t.Helper()
	const line = "description: \"A sentence carrying a fact not in the filename.\""

	var n int
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || filepath.Ext(p) != ".md" || filepath.Base(p) == "CLAUDE.md" {
			return nil
		}
		raw, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		text := string(raw)
		body, ok := strings.CutPrefix(text, "---\n")
		if !ok {
			return nil // no frontmatter — log.md, README.md
		}
		fm, rest, ok := strings.Cut(body, "\n---\n")
		if !ok {
			return nil
		}
		lines := strings.Split(fm, "\n")
		var out []string
		var inserted bool
		for _, l := range lines {
			out = append(out, l)
			if !inserted && strings.HasPrefix(l, "type:") {
				out = append(out, line)
				inserted = true
			}
		}
		if !inserted {
			return nil
		}
		n++
		return os.WriteFile(p,
			[]byte("---\n"+strings.Join(out, "\n")+"\n---\n"+rest), 0o666)
	})
	if err != nil {
		t.Fatalf("walking the fixture copy: %v", err)
	}
	return n
}
