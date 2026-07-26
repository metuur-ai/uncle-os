package scaffold

// The four scaffolding commands: init (:1968), add (:1997), reality new (:2030)
// and scratchpad init (:1141).
//
// Each returns a record describing what it created; cmd/company-os turns that
// into the frozen stdout lines and, from Phase 4, into the R-3.7 JSON envelope.
// Nothing here prints, so the same call is reusable from the TUI.

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/model"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/workspace"
)

// Prompt asks the user for one value and returns their answer, or "" to accept
// the default. It is nil when no terminal is attached.
//
// It is injected rather than implemented here because Python's _prompt writes
// the question to STDOUT with input() (bin/company-os:1964), which R-2.10 keeps
// above the dispatch seam, and because TTY detection is a property of the
// process, not of the scaffolder — which is what makes this testable from both
// directions without a pty.
type Prompt func(label, def string) (string, error)

// resolvePrompt is _prompt (bin/company-os:1954-1965): the flag wins; with no
// flag and a terminal the user is asked; with no flag and no terminal the run
// fails fast so scripted and CI runs stay reproducible (GPF-R-1.3).
//
// The non-interactive refusal is exit code 7, not 2: --company, --team and
// --platform are all OPTIONAL to argparse — `company-os init` with none of them
// parses fine — and the requirement materializes only from the absence of a
// TTY. The fact an agent needs is "there is no terminal here", which is what 7
// encodes (.devlocal/go-port/exit-code-map.md § D).
func resolvePrompt(ask Prompt, val, flag, label, def string) (string, error) {
	if val != "" {
		return val, nil
	}
	if ask == nil {
		return "", model.Errorf(model.ExitInteractive,
			"non-interactive run: --%s is required when no terminal is attached "+
				"(pass --company, --team, and --platform)", flag)
	}
	resp, err := ask(label, def)
	if err != nil {
		return "", err
	}
	resp = strings.TrimSpace(resp)
	if resp == "" {
		return def, nil
	}
	return resp, nil
}

// InitOptions carries `init`'s three flags plus the interaction seam.
type InitOptions struct {
	Company  string
	Team     string
	Platform string
	// Prompt is nil for a non-interactive run.
	Prompt Prompt
	// Rebuild derives the generated artifacts inside the staging workspace,
	// before anything is moved into place. Nil means no-op.
	Rebuild Rebuild
}

// InitResult is what `init` created.
type InitResult struct {
	Root     string
	Company  string
	Team     string
	Platform string
	// Generated is rebuild_generated's output, which the oracle prints before
	// init's own lines.
	Generated []string
}

// Init scaffolds a fresh workspace (bin/company-os:1968-1995, GPF-R-1.1).
//
// The staging directory is the whole point and is not an implementation detail:
// every source root is written into a temporary directory first and only then
// moved into the target, so an abort at any point — a failed prompt, a refused
// overwrite, a rebuild error — leaves nothing behind (GPF-R-1.4). The move loop
// itself is compensated: if root N fails to land, roots 1..N-1 are removed
// again, so a partial workspace is not reachable even from a mid-move failure.
func Init(ws *workspace.Workspace, opts InitOptions) (*InitResult, error) {
	target := ws.Root

	// GPF-R-1.2: refuse inside an existing workspace, mutating nothing.
	if ws.IsRoot() {
		return nil, conflict(target,
			"'%s' is already a workspace root (%s/ present) — refusing to re-init",
			target, strings.Join(workspace.CanonicalRoots, "/, "))
	}

	company, err := resolvePrompt(opts.Prompt, opts.Company, "company", "Company name", "My Company")
	if err != nil {
		return nil, err
	}
	rawTeam, err := resolvePrompt(opts.Prompt, opts.Team, "team", "First team id", "core")
	if err != nil {
		return nil, err
	}
	rawPlatform, err := resolvePrompt(opts.Prompt, opts.Platform, "platform", "First platform id", "platform-1")
	if err != nil {
		return nil, err
	}
	tid, pid := Slugify(rawTeam), Slugify(rawPlatform)

	staging, err := os.MkdirTemp("", "company-os-init-")
	if err != nil {
		return nil, model.Errorf(model.ExitArtifact, "cannot create staging directory: %v", err)
	}
	defer os.RemoveAll(staging)

	lines, err := scaffoldWorkspace(staging, company, pid, tid, opts.Rebuild)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(target, 0o777); err != nil {
		return nil, model.Errorf(model.ExitArtifact, "cannot create %s: %v", target, err)
	}
	if err := moveRoots(staging, target); err != nil {
		return nil, err
	}

	return &InitResult{
		Root: target, Company: company, Team: tid, Platform: pid, Generated: lines,
	}, nil
}

// scaffoldWorkspace is _scaffold_workspace (bin/company-os:1946-1951) plus the
// rebuild_generated call init makes against the staging root (:1984).
func scaffoldWorkspace(base, company, pid, tid string, rebuild Rebuild) ([]string, error) {
	if err := scaffoldCompany(base, company); err != nil {
		return nil, err
	}
	if err := scaffoldPlatform(base, pid); err != nil {
		return nil, err
	}
	if err := scaffoldTeam(base, tid); err != nil {
		return nil, err
	}
	if err := registerID(base, "platform://"+pid, "platforms/"+pid+"/platform.yaml"); err != nil {
		return nil, err
	}
	if err := registerID(base, "team://"+tid, "teams/"+tid+"/team.yaml"); err != nil {
		return nil, err
	}
	if rebuild == nil {
		return nil, nil
	}
	return rebuild(workspace.New(base))
}

// moveRoots is the compensated move at bin/company-os:1985-1994: each canonical
// root present in staging is moved into the target, and any failure unwinds the
// ones already moved.
func moveRoots(staging, target string) error {
	var moved []string
	for _, name := range workspace.CanonicalRoots {
		src := filepath.Join(staging, name)
		if _, err := os.Lstat(src); err != nil {
			continue
		}
		dst := filepath.Join(target, name)
		if err := move(src, dst); err != nil {
			for _, m := range moved {
				os.RemoveAll(m)
			}
			return model.Errorf(model.ExitArtifact, "cannot move %s into %s: %v", name, target, err)
		}
		moved = append(moved, dst)
	}
	return nil
}

// move is shutil.move: a rename where the two paths share a filesystem, and a
// copy-then-remove where they do not. The temporary directory is not guaranteed
// to be on the same device as the workspace, and a rename that fails with
// EXDEV must not abort the scaffold.
func move(src, dst string) error {
	if err := os.Rename(src, dst); err == nil {
		return nil
	}
	if err := copyTree(src, dst); err != nil {
		return err
	}
	return os.RemoveAll(src)
}

// copyTree reproduces shutil.copytree's mode preservation (copystat), which is
// what keeps the scaffolded directory and file modes identical on both sides of
// a cross-device move.
func copyTree(src, dst string) error {
	fi, err := os.Lstat(src)
	if err != nil {
		return err
	}
	if !fi.IsDir() {
		data, err := os.ReadFile(src)
		if err != nil {
			return err
		}
		return os.WriteFile(dst, data, fi.Mode().Perm())
	}
	if err := os.MkdirAll(dst, fi.Mode().Perm()); err != nil {
		return err
	}
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if err := copyTree(filepath.Join(src, e.Name()), filepath.Join(dst, e.Name())); err != nil {
			return err
		}
	}
	return os.Chmod(dst, fi.Mode().Perm())
}

// AddKind is `add`'s first positional.
type AddKind string

const (
	AddPlatform  AddKind = "platform"
	AddTeam      AddKind = "team"
	AddComponent AddKind = "component"
)

// AddResult is what `add` created. Platform is set for AddComponent only.
type AddResult struct {
	Kind      AddKind
	ID        string
	Platform  string
	Generated []string
}

// Add grows an existing workspace (bin/company-os:1997-2027, GPF-R-1.9). The
// unit is scaffolded from the same writers `init` uses, so the two cannot
// drift, and every writer still refuses to overwrite.
func Add(ws *workspace.Workspace, kind AddKind, name, platform string, rebuild Rebuild) (*AddResult, error) {
	switch kind {
	case AddPlatform:
		pid := Slugify(name)
		if err := scaffoldPlatform(ws.Root, pid); err != nil {
			return nil, err
		}
		if err := registerID(ws.Root, "platform://"+pid, "platforms/"+pid+"/platform.yaml"); err != nil {
			return nil, err
		}
		lines, err := runRebuild(ws, rebuild)
		if err != nil {
			return nil, err
		}
		return &AddResult{Kind: kind, ID: pid, Generated: lines}, nil

	case AddTeam:
		tid := Slugify(name)
		if err := scaffoldTeam(ws.Root, tid); err != nil {
			return nil, err
		}
		if err := registerID(ws.Root, "team://"+tid, "teams/"+tid+"/team.yaml"); err != nil {
			return nil, err
		}
		lines, err := runRebuild(ws, rebuild)
		if err != nil {
			return nil, err
		}
		return &AddResult{Kind: kind, ID: tid, Generated: lines}, nil

	case AddComponent:
		// Exit 2, not 3: --platform is optional to argparse and this is the
		// missing-required-argument error argparse would have raised had the
		// requirement been expressible there.
		if platform == "" {
			return nil, model.Errorf(model.ExitUsage,
				"add component requires --platform <platform-id>")
		}
		dir, err := ws.PlatformDir(platform)
		if err != nil {
			return nil, err
		}
		pid := filepath.Base(dir)
		cid := Slugify(name)
		if err := scaffoldComponent(ws.Root, pid, cid); err != nil {
			return nil, err
		}
		if err := registerID(ws.Root, "component://"+cid,
			"platforms/"+pid+"/components/"+cid+".yaml"); err != nil {
			return nil, err
		}
		lines, err := runRebuild(ws, rebuild)
		if err != nil {
			return nil, err
		}
		return &AddResult{Kind: kind, ID: cid, Platform: pid, Generated: lines}, nil
	}
	return nil, model.Errorf(model.ExitUsage, "add: unknown kind %q", string(kind))
}

func runRebuild(ws *workspace.Workspace, rebuild Rebuild) ([]string, error) {
	if rebuild == nil {
		return nil, nil
	}
	return rebuild(ws)
}

// RealityResult is what `reality new` created.
type RealityResult struct {
	// Path is workspace-relative and POSIX-separated, as the oracle prints it.
	Path string
	// Source is the template provenance label from ResolveTemplate.
	Source    string
	Platform  string
	Component string
	Generated []string
}

// RealityNew scaffolds a component's Representation of Reality document
// (bin/company-os:2030-2058).
//
// The `reality template not found` die at :2041 has no counterpart: it fires
// only when templates/reality-component.md is missing from the installation,
// and //go:embed makes that unreachable (R-1.11, carve-out R-0.7a(c)).
func RealityNew(ws *workspace.Workspace, platform, component string, rebuild Rebuild) (*RealityResult, error) {
	pdir, err := ws.PlatformDir(platform)
	if err != nil {
		return nil, err
	}
	out := filepath.Join(pdir, "reality", "components", component+".md")
	rel := relTo(ws.Root, out)
	if _, err := os.Stat(out); err == nil { // Path.exists(), as at :2036
		return nil, conflict(rel, "%s already exists — refusing to overwrite", rel)
	}

	text, source, err := ResolveTemplate(ws, TemplateRealityComponent, "", platform)
	if err != nil {
		return nil, err
	}

	name := component
	if _, descriptor, found := ws.FindComponent(component); found {
		if declared, err := componentName(descriptor); err != nil {
			return nil, err
		} else if declared != "" {
			name = declared
		}
	}

	content := strings.ReplaceAll(text, "reality-<component-id>", "reality-"+component)
	content = strings.ReplaceAll(content, "<YYYY-MM-DD>", time.Now().Format("2006-01-02"))
	content = strings.ReplaceAll(content, "<Component Name>", name)

	if err := os.MkdirAll(filepath.Dir(out), 0o777); err != nil {
		return nil, model.Errorf(model.ExitArtifact, "cannot create %s: %v", filepath.Dir(out), err)
	}
	if err := os.WriteFile(out, []byte(content), 0o666); err != nil {
		return nil, model.Errorf(model.ExitArtifact, "cannot write %s: %v", out, err)
	}

	lines, err := runRebuild(ws, rebuild)
	if err != nil {
		return nil, err
	}
	return &RealityResult{
		Path: rel, Source: source, Platform: platform, Component: component, Generated: lines,
	}, nil
}

// componentName is `(desc or {}).get("metadata", {}).get("name") or cid`, with
// the two shapes Python would raise on treated as "no name declared".
func componentName(descriptor string) (string, error) {
	loaded, err := loadPy(descriptor)
	if err != nil {
		return "", err
	}
	desc, ok := loaded.(pyMap)
	if !ok {
		return "", nil
	}
	meta, ok := desc.get("metadata").(pyMap)
	if !ok {
		return "", nil
	}
	name, _ := meta.get("name").(pyStr)
	return string(name), nil
}

// scratchpadDirs is the fixed set at bin/company-os:1143.
var scratchpadDirs = []string{"drafts", "brainstorms", "personal-rules", "experiments", "inbox"}

const scratchpadReadme = "# Scratchpad (local-only)\n\nPrivate working area. Never authoritative, " +
	"never committed.\nPersonal agent rules go in personal-rules/ and are applied " +
	"ON TOP of canonical skills,\nnever instead of mandatory steps.\n"

const gitignoreBlock = "\n# Local user and AI working area\nscratchpad/\n" +
	".company-os.local.yaml\n.env\n.env.local\n"

// ScratchpadResult is what `scratchpad init` created.
type ScratchpadResult struct {
	// Base is the scratchpad directory as Python's Path renders it — relative
	// to the process's working directory, with "." collapsed away, because
	// that string is printed verbatim.
	Base string
}

// ScratchpadInit creates the git-ignored local working area
// (bin/company-os:1141-1155).
//
// It is workspace-independent by design — it takes --repo, not the workspace
// root, which is why it is one of the two commands exempt from require_root
// (bin/company-os:2774). It is also the one mutating command that prints no
// next step, and R-1.9 keeps it that way: R-0.8 outranks R-1.8 here.
func ScratchpadInit(repo string) (*ScratchpadResult, error) {
	if repo == "" {
		repo = "."
	}
	base := pathJoin(repo, "scratchpad")
	for _, sub := range scratchpadDirs {
		if err := os.MkdirAll(filepath.Join(base, sub), 0o777); err != nil {
			return nil, model.Errorf(model.ExitArtifact, "cannot create %s: %v", base, err)
		}
	}
	if err := os.WriteFile(filepath.Join(base, "README.md"), []byte(scratchpadReadme), 0o666); err != nil {
		return nil, model.Errorf(model.ExitArtifact, "cannot write %s: %v",
			filepath.Join(base, "README.md"), err)
	}

	gi := pathJoin(repo, ".gitignore")
	existing, err := os.ReadFile(gi)
	if err != nil && !os.IsNotExist(err) {
		return nil, model.Errorf(model.ExitArtifact, "cannot read %s: %v", gi, err)
	}
	if !strings.Contains(string(existing), "scratchpad/") {
		f, err := os.OpenFile(gi, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o666)
		if err != nil {
			return nil, model.Errorf(model.ExitArtifact, "cannot open %s: %v", gi, err)
		}
		if _, err := f.WriteString(gitignoreBlock); err != nil {
			f.Close()
			return nil, model.Errorf(model.ExitArtifact, "cannot write %s: %v", gi, err)
		}
		if err := f.Close(); err != nil {
			return nil, model.Errorf(model.ExitArtifact, "cannot write %s: %v", gi, err)
		}
	}
	return &ScratchpadResult{Base: base}, nil
}

// pathJoin is `Path(a) / b` — filepath.Join with pathlib's normalization, which
// drops a lone "." segment so `--repo .` prints "scratchpad" and not
// "./scratchpad".
func pathJoin(a, b string) string {
	if a == "." {
		return b
	}
	return filepath.Join(a, b)
}

// relTo is Path.relative_to(root) rendered POSIX-style (R-1.12). It falls back
// to the absolute path when the target is outside the root, where Python raises
// ValueError.
func relTo(root, path string) string {
	rel, err := filepath.Rel(root, path)
	if err != nil || strings.HasPrefix(rel, "..") {
		return filepath.ToSlash(path)
	}
	return filepath.ToSlash(rel)
}
