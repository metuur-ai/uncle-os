package main

// `workspace sync` and `workspace status` — the federation dispatch handler.
//
// Like the scaffolding commands, both print prose rather than gate findings, so
// this file composes their lines from the typed records internal/federation
// returned — as whole-line findings, for the reasons scaffold.go's file comment
// gives. Sync is the only command in the CLI that emits part of its output and
// THEN fails: the oracle writes each repo's line as that repo completes and dies
// afterwards, so the partial record set is returned alongside the error and
// rendered before it (bin/company-os:2566-2606).

import (
	"fmt"
	"io"
	"strings"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/federation"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/model"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/workspace"
)

// cmdWorkspace is cmd_workspace (bin/company-os:2542-2553).
func cmdWorkspace(ws *workspace.Workspace, args *Args, _ io.Writer) ([]model.GateResult, error) {
	m, err := federation.LoadManifest(ws)
	if err != nil {
		return nil, err
	}
	if m == nil {
		// Absent, not malformed — and absence is legal (monorepo mode), so this
		// is a workspace error rather than an artifact one
		// (.devlocal/go-port/exit-code-map.md § `:2547`).
		return nil, model.Errorf(model.ExitWorkspace,
			"no %s at %s — this is a monorepo workspace and "+
				"needs no federation. add %s to federate.",
			workspace.ManifestName, ws.Root, workspace.ManifestName)
	}
	switch args.Action {
	case "sync":
		res, syncErr := federation.Sync(ws, m, federation.SyncOptions{
			Frozen: args.Frozen,
			Only:   args.Only,
		})
		return syncSections(res), syncErr
	case "status":
		res, err := federation.Status(ws, m)
		if err != nil {
			return nil, err
		}
		return statusSections(res), nil
	}
	return nil, nil
}

// syncSections records whatever Sync completed. A nil result means it failed
// before the header line, which is where require_git, --only and the --frozen
// lock check all sit.
//
// The blank line after the header and before the trailer is part of those lines
// (`\n` inside the f-string at `:2566` and `:2624`), not a renderer rule, so it
// stays in Message where the oracle put it.
func syncSections(res *federation.SyncResult) []model.GateResult {
	if res == nil {
		return nil
	}
	frozen := ""
	if res.Frozen {
		frozen = " --frozen"
	}
	s := model.GateResult{Ordinal: 1, Slug: model.SlugSync}
	s.Findings = append(s.Findings, line(model.CodeSyncHeader,
		fmt.Sprintf("workspace sync%s (%d repo(s))\n", frozen, res.RepoCount),
		model.Fields{"frozen": res.Frozen, "repos": res.RepoCount}))
	for _, r := range res.Repos {
		f := model.Fields{
			"name": r.Name, "sha": r.SHA, "targets": r.Targets,
			"files": r.Files, "restored": r.Restored,
		}
		if r.Restored {
			s.Findings = append(s.Findings, line(model.CodeSyncRepo,
				fmt.Sprintf("  restored %s @ %s (from lock) -> %s (%d file(s))",
					r.Name, shortSHA(r.SHA), strings.Join(r.Targets, ", "), r.Files), f))
			continue
		}
		f["pin"] = r.Pin.Kind + ":" + r.Pin.Ref
		s.Findings = append(s.Findings, line(model.CodeSyncRepo,
			fmt.Sprintf("  synced %s @ %s (%s %s) -> %s (%d file(s))",
				r.Name, shortSHA(r.SHA), r.Pin.Kind, r.Pin.Ref,
				strings.Join(r.Targets, ", "), r.Files), f))
	}
	if !res.Complete {
		// Sync died inside the repo loop; the oracle's trailer never printed.
		return []model.GateResult{s}
	}
	if res.Frozen {
		s.Findings = append(s.Findings, line(model.CodeSyncLock,
			fmt.Sprintf("\nmaterialized strictly from %s (no network)", workspace.LockName),
			model.Fields{"wrote": false}))
	} else {
		s.Findings = append(s.Findings, line(model.CodeSyncLock,
			fmt.Sprintf("\nwrote %s (%d repo(s))", workspace.LockName, res.LockRepos),
			model.Fields{"wrote": true, "lock": workspace.LockName, "repos": res.LockRepos}))
	}
	// One printed line carrying two commands, the second behind a `#` comment.
	// Only the first is the next step (R-3.6); the second is what `status` will
	// tell you anyway.
	s.Findings = append(s.Findings, line(model.CodeSyncNext,
		"next: company-os workspace status   # then: company-os validate",
		model.Fields{model.FieldNext: "company-os workspace status"}))
	return []model.GateResult{s}
}

func statusSections(res *federation.StatusResult) []model.GateResult {
	s := model.GateResult{Ordinal: 1, Slug: model.SlugStatus}
	s.Findings = append(s.Findings, line(model.CodeStatusHeader,
		fmt.Sprintf("workspace federation status (%d repo(s))\n", res.RepoCount),
		model.Fields{"repos": res.RepoCount}))
	for _, r := range res.Repos {
		f := model.Fields{
			"name": r.Name, "pin": r.Pin.Kind + ":" + r.Pin.Ref,
			"targets": r.Targets, "locked": r.Locked,
		}
		if !r.Locked {
			f["state"] = "missing"
			s.Findings = append(s.Findings, line(model.CodeStatusRepo,
				fmt.Sprintf("  %s: %s:%s — missing (never synced)",
					r.Name, r.Pin.Kind, r.Pin.Ref), f))
			continue
		}
		var label string
		switch {
		case r.PinDrift:
			label = fmt.Sprintf("drifted (manifest pin %s:%s != lock %s)",
				r.Pin.Kind, r.Pin.Ref, r.LockPin)
		case r.SliceDrift:
			label = fmt.Sprintf("drifted (slice set in %s != %s)",
				workspace.ManifestName, workspace.LockName)
		default:
			label = string(r.State)
		}
		f["sha"] = r.SHA
		f["state"] = string(r.State)
		f["pinDrift"] = r.PinDrift
		f["sliceDrift"] = r.SliceDrift
		s.Findings = append(s.Findings, line(model.CodeStatusRepo,
			fmt.Sprintf("  %s: %s:%s @ %s -> %s — %s",
				r.Name, r.Pin.Kind, r.Pin.Ref, shortSHA(r.SHA),
				strings.Join(r.Targets, ", "), label), f))
	}
	next := "company-os validate"
	if res.ActionNeeded {
		next = "company-os workspace sync"
	}
	s.Findings = append(s.Findings, line(model.CodeStatusNext, "\nnext: "+next,
		model.Fields{model.FieldNext: next}))
	return []model.GateResult{s}
}

// shortSHA is `sha[:12]`, which slices CHARACTERS. Every real SHA is ASCII, but
// the `?` placeholder a lock entry without resolvedCommit produces is shorter
// than 12 and must not be padded or truncated.
func shortSHA(sha string) string {
	r := []rune(sha)
	if len(r) <= 12 {
		return sha
	}
	return string(r[:12])
}
