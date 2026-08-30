package skills

// `skills install` — write the canonical skills the binary carries into
// `<root>/company-os/skills/`.
//
// This command has no Python oracle: the reference CLI shipped the skills as
// files a reader copied by hand, which is the gap it closes. It follows the
// scaffolding commands' conventions all the same — records out, no printing, a
// next-step line (R-1.8) — so `--json` and the TUI get it for free.
//
// The company layer is the only destination. All five skills declare
// `appliesTo: company://all-platforms` or `company://all-teams`, and derivation
// gives a company-layer skill exactly `[authority/canonical]`, which is what
// they carry — installing to a platform would put them immediately in tag
// drift until `graph build` ran.

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/model"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/workspace"
	skillfiles "github.com/metuur-ai/uncle-os/company-os-starter/skills"
)

// InstallOutcome is what happened to one skill.
type InstallOutcome struct {
	Name string
	Rel  string
	Code string
	// From and To are the installed and embedded versions. From is empty for a
	// fresh install; both are empty when the installed file could not be read.
	From string
	To   string
}

// InstallResult is the whole run, in the order the records were produced.
type InstallResult struct {
	Outcomes []InstallOutcome
	Dir      string // destination, workspace-relative
	// Generated is the derived-artifact rebuild's output, which the caller
	// prints BEFORE its own lines. See Rebuild.
	Generated []string
}

// Rebuild is the scaffold -> graph seam, declared here and satisfied in cmd/
// so this package never imports internal/graph.
//
// Writing skill files into company-os/skills/ changes a graph-docs root: the
// company-layer CLAUDE.md context node gains entries and the directory becomes
// index-eligible. Gate 5 fails on both until derivation runs, so an install
// that skipped the rebuild would leave a previously-green workspace red — which
// is exactly what acceptance caught before this existed.
type Rebuild func(ws *workspace.Workspace) ([]string, error)

// Install writes every embedded skill into `company-os/skills/`, comparing
// versions against whatever is already there.
//
// The four dispositions, and why each is what it is:
//
//   - absent          -> write it (CodeSkillsInstalled)
//   - older installed -> overwrite (CodeSkillsUpdated). A stale skill is worse
//     than a replaced one: the whole point of shipping a version is that the
//     newer text is the one the CLI's exit codes match.
//   - same version    -> leave it (CodeSkillsUnchanged). Byte differences at
//     the same version are a team's local edit, and a version that did not
//     move is not a claim to overwrite them.
//   - newer installed -> leave it (CodeSkillsLocallyNewer). The workspace is
//     ahead of the binary; overwriting would be a downgrade.
//   - unreadable      -> leave it (CodeSkillsUnreadable). Never guess at a file
//     whose frontmatter will not parse.
//
// Nothing here removes a file, and nothing writes outside the destination
// directory.
//
// The rebuild runs last and unconditionally — even when nothing changed, since
// a workspace whose derived artifacts were already stale should not be left
// that way by a command that just walked the same directory.
func Install(ws *workspace.Workspace, rebuild Rebuild) (*InstallResult, error) {
	embedded, err := skillfiles.All()
	if err != nil {
		return nil, model.Errorf(model.ExitArtifact, "cannot read embedded skills: %v", err)
	}

	dir := filepath.Join(ws.Root, "company-os", "skills")
	if err := os.MkdirAll(dir, 0o777); err != nil {
		return nil, model.Errorf(model.ExitArtifact, "cannot create %s: %v", dir, err)
	}

	res := &InstallResult{Dir: filepath.ToSlash(filepath.Join("company-os", "skills"))}
	for _, f := range embedded {
		name := f.Name + Suffix
		path := filepath.Join(dir, name)
		out := InstallOutcome{
			Name: f.Name,
			Rel:  filepath.ToSlash(filepath.Join("company-os", "skills", name)),
			To:   f.Version,
		}

		existing, readErr := os.ReadFile(path)
		switch {
		case os.IsNotExist(readErr):
			if err := os.WriteFile(path, f.Body, 0o666); err != nil {
				return nil, model.Errorf(model.ExitArtifact, "cannot write %s: %v", path, err)
			}
			out.Code = model.CodeSkillsInstalled

		case readErr != nil:
			out.Code = model.CodeSkillsUnreadable

		default:
			out.From = frontmatterVersion(existing)
			switch cmp := compareVersions(out.From, f.Version); {
			case out.From == "":
				out.Code = model.CodeSkillsUnreadable
			case cmp < 0:
				if err := os.WriteFile(path, f.Body, 0o666); err != nil {
					return nil, model.Errorf(model.ExitArtifact, "cannot write %s: %v", path, err)
				}
				out.Code = model.CodeSkillsUpdated
			case cmp > 0:
				out.Code = model.CodeSkillsLocallyNewer
			default:
				out.Code = model.CodeSkillsUnchanged
			}
		}
		res.Outcomes = append(res.Outcomes, out)
	}

	if rebuild != nil {
		lines, err := rebuild(ws)
		if err != nil {
			return nil, err
		}
		res.Generated = lines
	}
	return res, nil
}

// frontmatterVersion reads `version:` from an installed skill's frontmatter.
//
// It mirrors skillfiles.version rather than importing it: that function reads
// the bytes the binary carries, this one reads a file a team may have edited,
// and the two happening to agree today is not a reason to couple them.
func frontmatterVersion(body []byte) string {
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

// compareVersions orders two dotted-numeric versions: -1 when a < b, 0 when
// equal, 1 when a > b.
//
// Numeric per segment, so 1.10 > 1.9 — which a string comparison gets wrong and
// which will matter the first time a skill reaches its tenth revision. A
// segment that is not a number sorts before every number, so a malformed
// version reads as older and gets replaced rather than silently kept.
func compareVersions(a, b string) int {
	as, bs := strings.Split(a, "."), strings.Split(b, ".")
	for i := 0; i < len(as) || i < len(bs); i++ {
		x, y := segment(as, i), segment(bs, i)
		if x != y {
			if x < y {
				return -1
			}
			return 1
		}
	}
	return 0
}

func segment(parts []string, i int) int {
	if i >= len(parts) {
		return 0
	}
	n, err := strconv.Atoi(strings.TrimSpace(parts[i]))
	if err != nil {
		return -1
	}
	return n
}
