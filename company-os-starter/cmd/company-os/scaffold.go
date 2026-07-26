package main

// The scaffolding cluster's dispatch handlers: init, add, reality new and
// scratchpad init.
//
// Every line printed here is byte-frozen output under R-0.8, and every one of
// the first three ends with the next command in the workflow (R-1.8). The
// fourth deliberately does not: `scratchpad init` prints no next step in the
// oracle, and R-1.9 says it must keep printing exactly what it prints today,
// because R-0.8 outranks R-1.8. Do not "complete" the chain here.

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/model"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/scaffold"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/workspace"
)

// rebuildGenerated is the scaffold -> graph seam (bin/company-os:1803).
//
// It is nil until task 2.3 lands internal/graph, at which point this becomes
// graph.RebuildGenerated and nothing else in this file changes. While it is nil
// the scaffolding commands write their SOURCE artifacts correctly and produce
// none of the derived ones: no frontmatter tag rewrite, no feature-index, no
// CLAUDE.md node, and none of the "  wrote index …" / "  node …" lines the
// oracle prints before each command's own output.
var rebuildGenerated scaffold.Rebuild

// cmdInit is cmd_init (bin/company-os:1968-1995).
func cmdInit(ws *workspace.Workspace, args *Args, out io.Writer) ([]model.GateResult, error) {
	res, err := scaffold.Init(ws, scaffold.InitOptions{
		Company:  args.Company,
		Team:     args.Team,
		Platform: args.Platform,
		Prompt:   stdinPrompt(out),
		Rebuild:  rebuildGenerated,
	})
	if err != nil {
		return nil, err
	}
	writeLines(out, res.Generated)
	fmt.Fprintf(out, "initialized workspace at %s\n", res.Root)
	fmt.Fprintf(out, "  company: %s | first team: %s | first platform: %s\n",
		res.Company, res.Team, res.Platform)
	fmt.Fprintf(out, "next: cd %s && company-os discover new --team %s \"<discovery title>\"\n",
		res.Root, res.Team)
	return nil, nil
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
func cmdAdd(ws *workspace.Workspace, args *Args, out io.Writer) ([]model.GateResult, error) {
	res, err := scaffold.Add(ws, scaffold.AddKind(args.Kind), args.Name, args.Platform, rebuildGenerated)
	if err != nil {
		return nil, err
	}
	writeLines(out, res.Generated)
	switch res.Kind {
	case scaffold.AddPlatform:
		fmt.Fprintf(out, "added platform '%s'\n", res.ID)
		fmt.Fprintf(out, "next: company-os add component --platform %s <component-id>\n", res.ID)
	case scaffold.AddTeam:
		fmt.Fprintf(out, "added team '%s'\n", res.ID)
		fmt.Fprintf(out, "next: company-os discover new --team %s \"<discovery title>\"\n", res.ID)
	case scaffold.AddComponent:
		fmt.Fprintf(out, "added component '%s' to platform '%s'\n", res.ID, res.Platform)
		fmt.Fprintf(out, "next: company-os reality new --platform %s %s\n", res.Platform, res.ID)
	}
	return nil, nil
}

// cmdReality is cmd_reality (bin/company-os:2030-2058).
func cmdReality(ws *workspace.Workspace, args *Args, out io.Writer) ([]model.GateResult, error) {
	res, err := scaffold.RealityNew(ws, args.Platform, args.ComponentArg, rebuildGenerated)
	if err != nil {
		return nil, err
	}
	writeLines(out, res.Generated)
	fmt.Fprintf(out, "created %s\n", res.Path)
	fmt.Fprintf(out, "  template: %s\n", res.Source)
	fmt.Fprintf(out, "next: fill in Business rules / Current limitations, then continue: "+
		"company-os prd complete --platform %s <prd-id>\n", res.Platform)
	return nil, nil
}

// cmdScratchpad is cmd_scratchpad (bin/company-os:1141-1155). One line, no next
// step — see the file comment.
func cmdScratchpad(_ *workspace.Workspace, args *Args, out io.Writer) ([]model.GateResult, error) {
	res, err := scaffold.ScratchpadInit(args.Repo)
	if err != nil {
		return nil, err
	}
	fmt.Fprintf(out, "initialized %s and updated .gitignore\n", res.Base)
	return nil, nil
}

func writeLines(out io.Writer, lines []string) {
	for _, l := range lines {
		fmt.Fprintln(out, l)
	}
}
