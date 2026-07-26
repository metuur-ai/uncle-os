package main

// The scaffolding cluster's dispatch handlers: init, add, reality new and
// scratchpad init.
//
// Every line printed here is byte-frozen output under R-0.8, and every one of
// the first three ends with the next command in the workflow (R-1.8). The
// fourth deliberately does not: `scratchpad init` prints no next step in the
// oracle, and R-1.9 says it must keep printing exactly what it prints today,
// because R-0.8 outranks R-1.8. Do not "complete" the chain here.
//
// All four return RECORDS rather than writing to `out`. They wrote prose
// directly until R-3.4b: one JSON encoder over the record types cannot serve a
// command that produces none, and R-3.7 requires their `--json` envelope to name
// what they created instead of defaulting to an empty document. renderPlain
// writes each finding's Message back out verbatim, so the bytes are the same
// bytes — the format strings moved, they did not change. The next command also
// lands in Fields under model.FieldNext, which is R-3.6.

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/graph"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/model"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/render"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/scaffold"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/workspace"
)

// rebuildGenerated is the scaffold -> graph seam (bin/company-os:1803).
//
// internal/scaffold declares the type and internal/graph satisfies it; the wire
// is here because this is the only package allowed to depend on both, which is
// what keeps `scaffold -> graph` from becoming an import edge and a cycle.
//
// The []string return is not a convenience. Output ORDER is load-bearing: every
// derived line the rebuild produces precedes the command's own "added platform
// 'x'", and only cmd/ may write. So graph returns records, render.Graph turns
// them into the same sentences `graph build` prints, and the caller emits them
// in front of its own.
var rebuildGenerated scaffold.Rebuild = func(ws *workspace.Workspace) ([]string, error) {
	sections, err := graph.Rebuild(ws)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := render.Graph(&buf, sections); err != nil {
		return nil, err
	}
	text := strings.TrimSuffix(buf.String(), "\n")
	if text == "" {
		return nil, nil
	}
	return strings.Split(text, "\n"), nil
}

// cmdInit is cmd_init (bin/company-os:1968-1995).
func cmdInit(ws *workspace.Workspace, args *Args, out io.Writer) ([]model.GateResult, error) {
	res, err := scaffold.Init(ws, scaffold.InitOptions{
		Company:  args.Company,
		Team:     args.Team,
		Platform: args.Platform,
		Prompt:   stdinPrompt(promptWriter(args, out)),
		Rebuild:  rebuildGenerated,
	})
	if err != nil {
		return nil, err
	}
	next := fmt.Sprintf("cd %s && company-os discover new --team %s \"<discovery title>\"",
		res.Root, res.Team)
	return append(generatedSection(res.Generated), model.GateResult{
		Ordinal: 1, Slug: model.SlugInit, Title: res.Company,
		Findings: []model.Finding{
			line(model.CodeInitCreated,
				fmt.Sprintf("initialized workspace at %s", res.Root),
				model.Fields{"root": res.Root}),
			line(model.CodeInitSummary,
				fmt.Sprintf("  company: %s | first team: %s | first platform: %s",
					res.Company, res.Team, res.Platform),
				model.Fields{
					"company": res.Company, "team": res.Team, "platform": res.Platform,
				}),
			line(model.CodeInitNext, "next: "+next,
				model.Fields{model.FieldNext: next}),
		},
	}), nil
}

// stdinPrompt returns the interactive half of _prompt (bin/company-os:1962-1965),
// or nil when no terminal is attached — which is what makes the wizard fail
// fast under CI instead of blocking on a read that will never be answered.
func stdinPrompt(out io.Writer) scaffold.Prompt {
	if !isTerminal(os.Stdin) {
		return nil
	}
	reader := bufio.NewReader(os.Stdin)
	return func(label, def string) (string, error) {
		fmt.Fprintf(out, "%s [%s]: ", label, def)
		line, err := reader.ReadString('\n')
		if err != nil && line == "" {
			return "", model.Errorf(model.ExitInteractive, "reading %s: %v", label, err)
		}
		return line, nil
	}
}

// cmdAdd is cmd_add (bin/company-os:1997-2027).
func cmdAdd(ws *workspace.Workspace, args *Args, _ io.Writer) ([]model.GateResult, error) {
	res, err := scaffold.Add(ws, scaffold.AddKind(args.Kind), args.Name, args.Platform, rebuildGenerated)
	if err != nil {
		return nil, err
	}
	var created, next string
	fields := model.Fields{"kind": string(res.Kind), "id": res.ID}
	switch res.Kind {
	case scaffold.AddPlatform:
		created = fmt.Sprintf("added platform '%s'", res.ID)
		next = fmt.Sprintf("company-os add component --platform %s <component-id>", res.ID)
	case scaffold.AddTeam:
		created = fmt.Sprintf("added team '%s'", res.ID)
		next = fmt.Sprintf("company-os discover new --team %s \"<discovery title>\"", res.ID)
	case scaffold.AddComponent:
		created = fmt.Sprintf("added component '%s' to platform '%s'", res.ID, res.Platform)
		next = fmt.Sprintf("company-os reality new --platform %s %s", res.Platform, res.ID)
		fields["platform"] = res.Platform
	}
	return append(generatedSection(res.Generated), model.GateResult{
		Ordinal: 1, Slug: model.SlugAdd, Title: res.ID,
		Findings: []model.Finding{
			line(model.CodeAddCreated, created, fields),
			line(model.CodeAddNext, "next: "+next, model.Fields{model.FieldNext: next}),
		},
	}), nil
}

// cmdReality is cmd_reality (bin/company-os:2030-2058).
func cmdReality(ws *workspace.Workspace, args *Args, _ io.Writer) ([]model.GateResult, error) {
	res, err := scaffold.RealityNew(ws, args.Platform, args.ComponentArg, rebuildGenerated)
	if err != nil {
		return nil, err
	}
	// The oracle's next-step line wraps its command in prose, so the bare command
	// and the rendered line are spelled separately (R-3.6).
	next := fmt.Sprintf("company-os prd complete --platform %s <prd-id>", res.Platform)
	return append(generatedSection(res.Generated), model.GateResult{
		Ordinal: 1, Slug: model.SlugRealityNew, Title: res.Component,
		Findings: []model.Finding{
			line(model.CodeRealityCreated, "created "+res.Path,
				model.Fields{"path": res.Path, "component": res.Component,
					"platform": res.Platform}),
			line(model.CodeRealityTemplate, "  template: "+res.Source,
				model.Fields{"source": res.Source}),
			line(model.CodeRealityNext,
				"next: fill in Business rules / Current limitations, then continue: "+next,
				model.Fields{model.FieldNext: next}),
		},
	}), nil
}

// cmdScratchpad is cmd_scratchpad (bin/company-os:1141-1155). One line, no next
// step — see the file comment.
func cmdScratchpad(_ *workspace.Workspace, args *Args, _ io.Writer) ([]model.GateResult, error) {
	res, err := scaffold.ScratchpadInit(args.Repo)
	if err != nil {
		return nil, err
	}
	return []model.GateResult{{
		Ordinal: 1, Slug: model.SlugScratchpad, Title: res.Base,
		Findings: []model.Finding{
			line(model.CodeScratchpadCreated,
				fmt.Sprintf("initialized %s and updated .gitignore", res.Base),
				model.Fields{"path": res.Base}),
		},
	}}, nil
}

// generatedSection wraps rebuild_generated's lines, which the oracle prints
// before each scaffolding command's own output. Nothing is returned for an empty
// rebuild, so the section never appears as an empty block in `--json`.
func generatedSection(lines []string) []model.GateResult {
	if len(lines) == 0 {
		return nil
	}
	s := model.GateResult{Ordinal: 0, Slug: model.SlugGenerated}
	for _, l := range lines {
		s.Findings = append(s.Findings, line(model.CodeGenerated, l, nil))
	}
	return []model.GateResult{s}
}

// line builds one whole-line record. Message is the finished line — see the file
// comment for why these five commands freeze their bytes at the producer rather
// than at the renderer.
func line(code, message string, f model.Fields) model.Finding {
	if f == nil {
		f = model.Fields{}
	}
	return model.Finding{
		Severity: model.SevOK, Code: code, Message: message, Fields: f,
	}
}

// promptWriter keeps the interactive wizard's prompts off `--json` stdout
// (R-3.2, R-3.9). They are progress, not results.
func promptWriter(args *Args, out io.Writer) io.Writer {
	if args.JSON {
		return os.Stderr
	}
	return out
}
