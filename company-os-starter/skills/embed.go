// Package skillfiles carries the canonical agent skills into the binary.
//
// It exists for the same structural reason as templates/embed.go: a //go:embed
// directive may only name files at or below its own package directory, so
// nothing under internal/ can embed company-os-starter/skills/. The directive
// lives here, beside the files it embeds, and the .SKILL.md files stay exactly
// where they are — one source of text for the reference copy a reader browses
// and the bytes `skills install` writes into a workspace.
//
// Nothing but embedded bytes and their parsed identity lives here. Placement,
// version comparison and the finding records are internal/skills' business, so
// the logic stays under the AST-enforced no-exit/no-stdout rule that
// cmd/company-os/architecture_test.go applies to internal/.
package skillfiles

import (
	"embed"
	"io/fs"
	"sort"
	"strings"
)

// FS holds every canonical skill, flat, exactly as committed.
//
// The glob matches the discovery contract: `*.SKILL.md` one level deep. A skill
// added to this directory ships in the next binary with no code change; one that
// reverts to the nested `<name>/SKILL.md` layout silently stops shipping, which
// is what internal/skills' conformance test is there to catch.
//
//go:embed *.SKILL.md
var FS embed.FS

// Suffix is the discovery contract's filename suffix, duplicated from
// internal/skills.Suffix rather than imported: this package sits above
// internal/ and must not depend on it.
const Suffix = ".SKILL.md"

// File is one embedded skill: the name discovery would give it, its declared
// version, and its bytes.
type File struct {
	Name    string // file stem, e.g. "creating-prd"
	Version string // frontmatter `version:`, empty when absent or unparsable
	Body    []byte
}

// All returns every embedded skill, sorted by name.
//
// Sorting is not cosmetic: it fixes the order findings are emitted in, so a
// command's output does not depend on the filesystem's iteration order.
func All() ([]File, error) {
	names, err := fs.Glob(FS, "*"+Suffix)
	if err != nil {
		return nil, err
	}
	sort.Strings(names)
	out := make([]File, 0, len(names))
	for _, n := range names {
		b, err := FS.ReadFile(n)
		if err != nil {
			return nil, err
		}
		out = append(out, File{
			Name:    strings.TrimSuffix(n, Suffix),
			Version: version(b),
			Body:    b,
		})
	}
	return out, nil
}

// version reads the frontmatter `version:` value, unquoted.
//
// This is a deliberately narrow scan rather than a YAML load: the only field
// this package needs is a scalar on its own line, and parsing frontmatter is
// internal/yamlio's contract, which this package cannot import. A value it
// cannot find is returned empty, and the caller decides what an unversioned
// skill means — this function never guesses a default.
func version(body []byte) string {
	text := string(body)
	if !strings.HasPrefix(text, "---\n") {
		return ""
	}
	end := strings.Index(text[4:], "\n---\n")
	if end < 0 {
		return ""
	}
	for _, line := range strings.Split(text[4:4+end], "\n") {
		rest, ok := strings.CutPrefix(line, "version:")
		if !ok {
			continue
		}
		return strings.Trim(strings.TrimSpace(rest), `'"`)
	}
	return ""
}
