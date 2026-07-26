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
	return product.DiscoverValidate(ws, args.Team, args.ID)
}

// cmdPRD is cmd_prd (bin/company-os:573-711).
func cmdPRD(ws *workspace.Workspace, args *Args, _ io.Writer) ([]model.GateResult, error) {
	switch args.Action {
	case "new":
		return product.PRDNew(ws, args.Team, args.Platform, args.Components,
			args.Title, args.FromDiscovery)
	case "validate":
		return product.PRDValidate(ws, args.Platform, args.ID)
	}
	return product.PRDComplete(ws, args.Platform, args.ID, args.Force, rebuildSections)
}

// cmdCheck is cmd_check (bin/company-os:731-733).
func cmdCheck(ws *workspace.Workspace, args *Args, _ io.Writer) ([]model.GateResult, error) {
	return product.Check(ws, args.Team, args.Components, args.Kind)
}
