package product

// Context inference for `prd new` (ux-simplification Phase 2, Unit 2).
//
// Phase 1 gave `prd validate`, `prd complete` and `discover validate` the right
// to work out WHERE an artifact already lives. This unit gives `prd new` the
// right to work out where a new one BELONGS — which is a different claim, and
// the reason the two resolvers do not share code.
//
// The Phase 1 pair (cmd/company-os/product.go) search for an existing artifact
// by id: FindPRD, FindDiscovery. `prd new` has no artifact to find yet, so
// inference here enumerates the workspace instead (R-2.9).
//
// This deliberately softens the rule stated at cmd/company-os/args.go's
// discover-new suspension — "inference cannot help a create". That holds when
// the flag names WHERE THE THING WILL BE WRITTEN and there is a real choice:
// `discover new --team` picks one of several teams to write into, and guessing
// would scatter briefs. It does not hold when the workspace admits exactly one
// answer, which is the only case any function here acts on.
//
// Three rules, shared with the Phase 1 resolvers:
//
//  1. An explicit flag always wins, unread and unvalidated (R-2.6).
//  2. Inference fires only on a UNIQUE match. There is no best guess (R-2.7),
//     because `prd new` WRITES — a wrong platform files the change record in
//     the wrong catalog, and that is worse than an error message.
//  3. Zero candidates return "" rather than an error, so the existing
//     downstream diagnostic keeps its voice (PlatformDir("") -> notFound).
//
// `--components` is deliberately NOT inferred; see InferComponents.

import (
	"path/filepath"
	"strings"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/model"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/workspace"
)

// InferPlatformForNew resolves the platform a new PRD is written into.
//
// R-2.1/R-2.2/R-2.3. Enumerates platforms/ (R-2.9) rather than searching for a
// record by id, because at `prd new` time there is no record.
func InferPlatformForNew(ws *workspace.Workspace, flag string) (string, error) {
	if flag != "" {
		return flag, nil
	}
	// AllPlatforms returns DIRECTORY PATHS, not ids (workspace.subdirs joins
	// each entry onto the parent). PlatformDir wants the id, so every path is
	// reduced to its base name before it is used or reported.
	var platforms []string
	for _, dir := range ws.AllPlatforms() {
		platforms = append(platforms, filepath.Base(dir))
	}
	switch len(platforms) {
	case 1:
		return platforms[0], nil
	case 0:
		// Let PlatformDir("") report it. A workspace with no platforms at all is
		// not an ambiguity the user can resolve by passing a flag, so a usage
		// error here would name a fix that does not exist.
		return "", nil
	default:
		return "", model.Usagef("prd",
			"workspace has multiple platforms: %s — pass --platform to pick one",
			strings.Join(platforms, ", "))
	}
}

// InferTeamForNew resolves the owning team of a new PRD from the discovery
// brief it is being written from (R-2.4).
//
// This is a LOOKUP, not an enumeration: the brief already sits inside exactly
// one team's product/discovery/, so the answer is a fact on disk rather than a
// guess among candidates. That makes it the strongest of the three inferences,
// and it is why it is the only one that does not need a uniqueness argument.
//
// Without --from-discovery there is nothing to look up and "" passes through,
// preserving today's behavior exactly: PRDNew's carryDiscovery returns early
// when fromDiscovery is empty, and Gather reports the missing team downstream.
// Inferring a lone team here instead was considered and rejected as scope
// beyond R-2.4 — see the amendment note in the EARS.
func InferTeamForNew(ws *workspace.Workspace, flag, fromDiscovery string) (string, error) {
	if flag != "" {
		return flag, nil
	}
	if fromDiscovery == "" {
		return "", nil
	}
	team, candidates := ws.FindDiscovery(fromDiscovery)
	switch {
	case team != "":
		return team, nil
	case len(candidates) > 1:
		return "", model.Usagef("prd",
			"discovery brief '%s' found under multiple teams: %s — pass --team to pick one",
			fromDiscovery, strings.Join(candidates, ", "))
	default:
		// carryDiscovery re-reads the brief and produces the canonical
		// "discovery brief not found" workspace error, so returning "" here
		// keeps a single spelling of that message.
		return "", nil
	}
}

// InferComponentsForNew resolves the component list of a new PRD (R-2.5).
//
// R-2.15 permitted dropping this if it could not be shown to resolve safely.
// It was first dropped on the argument that `--components` has a legal empty
// default, so omitting it was an existing working invocation that R-0.13
// protects. Testing disproved that: omitting it yields `components: []`, which
// FAILS `prd validate`'s process contract AND gate 3 ("missing frontmatter
// ['components']"). There is no working behavior to protect — the flag-free
// path produced an invalid artifact — so inference here repairs a defect rather
// than altering a success.
//
// The source is effectiveGovernance, NOT ownership/components.yaml, and that is
// deliberate: PRDNew calls Gather immediately afterwards, which resolves the
// component list against exactly this file. Inferring from the authored
// ownership registry instead could propose a component that Gather then reports
// as "not in any generated governance file" — the tool contradicting itself
// inside one command. Same source in, same source out.
//
// A stale generated file therefore suppresses inference rather than corrupting
// it: no components, and the existing downstream diagnostic still fires.
func InferComponentsForNew(ws *workspace.Workspace, flag, team string) (string, error) {
	if flag != "" {
		return flag, nil
	}
	if team == "" {
		return "", nil
	}
	eff, err := effectiveGovernance(ws, team)
	if err != nil {
		// Not an error here. Gather is about to read the same file and owns the
		// diagnostic; failing twice in two voices helps nobody.
		return "", nil
	}
	all, ok := eff.Get("components").(pyMap)
	if !ok {
		return "", nil
	}
	// PyMap is an ordered []PyPair, not a Go map, so key order here is the
	// generated file's order — which keeps the multi-component error message
	// stable across runs.
	var owned []string
	for _, kv := range all {
		owned = append(owned, kv.K)
	}
	if len(owned) == 1 {
		return owned[0], nil
	}
	// Zero owned components: nothing to infer, and the empty list reproduces
	// today's behavior. Several: the checklist a wrong pick would generate is
	// the most expensive wrong answer `prd new` can give (R-2.7), so refuse and
	// name them.
	if len(owned) > 1 {
		return "", model.Usagef("prd",
			"team '%s' owns multiple components: %s — pass --components to pick",
			team, strings.Join(owned, ", "))
	}
	return "", nil
}
