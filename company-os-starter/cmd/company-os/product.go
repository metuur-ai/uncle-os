package main

// The product cluster's dispatch handlers: discover new|validate, prd
// new|validate|complete and check ready|done.
//
// None of them formats anything — internal/product returns the record set and
// render.Product turns it into bytes — so `out` goes unused here, as it does for
// the other record-returning commands.
//
// Two of the three carry an exit code the dispatcher derives rather than is
// told: `discover validate` and `prd validate` exit 1 through HasFailure when
// they emit a [FAIL] (exit-code map § H), while `prd complete`'s done-gate
// refusal returns product.ErrDoneCheck — quiet, exit 5 — so its stdout block
// renders and nothing reaches stderr.

import (
	"io"
	"strings"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/graph"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/model"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/product"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/workspace"
)

// rebuildSections is the product -> graph seam (bin/company-os:711). It is the
// records-returning twin of scaffold.go's rebuildGenerated: `prd complete`
// splices these sections into its own output, between the archive lines and the
// next step, so they have to stay records until render.Product runs.
var rebuildSections product.Rebuild = graph.Rebuild

// cmdDiscover is cmd_discover (bin/company-os:409-464).
func cmdDiscover(ws *workspace.Workspace, args *Args, _ io.Writer) ([]model.GateResult, error) {
	if args.Action == "new" {
		return product.DiscoverNew(ws, args.Team, args.TitleArg)
	}
	// ux-simplification 1.2: --team is optional for validate when the brief
	// id is unique across the workspace. `discover new` genuinely needs its
	// team (it creates under it) and is therefore not inferred.
	team, err := resolveTeam(ws, args.Team, args.ID, "discover")
	if err != nil {
		return nil, err
	}
	return product.DiscoverValidate(ws, team, args.ID)
}

// cmdPRD is cmd_prd (bin/company-os:573-711).
func cmdPRD(ws *workspace.Workspace, args *Args, _ io.Writer) ([]model.GateResult, error) {
	switch args.Action {
	case "new":
		if args.Draft {
			// `prd new --draft "<title>"` (R-2.1) takes its title where
			// `discover new` does — the positional — while --title keeps
			// working for symmetry with the non-draft path.
			title := args.Title
			if title == "" {
				title = args.ID
			}
			return product.DraftNew(ws, args.Team, args.Platform, title,
				args.FromDiscovery, rebuildSections)
		}
		return product.PRDNew(ws, args.Team, args.Platform, args.Components,
			args.Title, args.FromDiscovery)
	case "validate":
		// ux-simplification 1.2: --platform is optional for validate and
		// complete when the PRD id is unique across the workspace.
		platform, err := resolvePlatform(ws, args.Platform, args.ID, "prd")
		if err != nil {
			return nil, err
		}
		return product.PRDValidate(ws, platform, args.ID)
	case "promote":
		// `prd promote --team <t> <draft-id>` (R-5.1). The target platform is
		// the draft's, never the flag's, so nothing here reads args.Platform.
		return product.PRDPromote(ws, args.Team, args.ID, rebuildSections)
	case "abandon":
		// `prd abandon --team <t> <draft-id>` (R-8.1). Like promote, it reads no
		// platform: a draft that is never going anywhere has no target.
		return product.PRDAbandon(ws, args.Team, args.ID, rebuildSections)
	}
	// complete (default action)
	platform, err := resolvePlatform(ws, args.Platform, args.ID, "prd")
	if err != nil {
		return nil, err
	}
	return product.PRDComplete(ws, platform, args.ID, args.Force, rebuildSections)
}

// resolvePlatform returns the platform a prd subcommand should operate on.
//
// ux-simplification 1.2: when --platform was supplied, it wins unchanged —
// byte-identical behavior for every existing flag-carrying invocation. When
// omitted and an id is available, the workspace is scanned for exactly one
// match (workspace.FindPRD). Ambiguity is a usage error naming every candidate
// so the user can pick without another command; absence is a workspace error
// in the style of the existing "no active PRD at …" path.
//
// When both the flag and the id are empty, the empty string passes through so
// the downstream requirePRDID / PlatformDir("") reports the missing argument in
// the existing voice.
func resolvePlatform(ws *workspace.Workspace, flag, id, scope string) (string, error) {
	if flag != "" {
		return flag, nil
	}
	if id == "" {
		return "", nil
	}
	platform, candidates := ws.FindPRD(id)
	switch {
	case platform != "":
		return platform, nil
	case len(candidates) > 1:
		return "", model.Usagef(scope,
			"PRD '%s' found under multiple platforms: %s — pass --platform to pick one",
			id, strings.Join(candidates, ", "))
	default:
		return "", model.Errorf(model.ExitWorkspace,
			"no PRD '%s' found under any platform", id)
	}
}

// resolveTeam is the team-side twin of resolvePlatform (ux-simplification 1.2).
// It backs `discover validate` without --team. The scan covers
// teams/*/product/discovery/ — the single location DiscoverNew writes to and
// DiscoverValidate reads from.
func resolveTeam(ws *workspace.Workspace, flag, id, scope string) (string, error) {
	if flag != "" {
		return flag, nil
	}
	if id == "" {
		return "", nil
	}
	team, candidates := ws.FindDiscovery(id)
	switch {
	case team != "":
		return team, nil
	case len(candidates) > 1:
		return "", model.Usagef(scope,
			"discovery brief '%s' found under multiple teams: %s — pass --team to pick one",
			id, strings.Join(candidates, ", "))
	default:
		return "", model.Errorf(model.ExitWorkspace,
			"no discovery brief '%s' found under any team", id)
	}
}

// cmdCheck is cmd_check (bin/company-os:731-733).
func cmdCheck(ws *workspace.Workspace, args *Args, _ io.Writer) ([]model.GateResult, error) {
	return product.Check(ws, args.Team, args.Components, args.Kind)
}
