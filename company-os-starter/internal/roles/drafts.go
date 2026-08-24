package roles

// The `today --team <t>` open-draft seam (R-9.1 … R-9.5).
//
// The listing itself is built by internal/product, which owns draft residency
// (R-1.1 … R-1.3) and the placeholder predicate R-9.3 needs. It is INJECTED
// rather than imported because internal/product depends, through its governance
// checklist, on internal/governance -> internal/ids -> internal/roles: importing
// it back would close that loop. This is the same seam product.Rebuild already
// is, wired in the same place (cmd/company-os), for the same reason — and it
// leaves the one thing that is genuinely a view decision, WHERE the section goes
// and whether it appears at all, in the view package.

import (
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/model"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/workspace"
)

// DraftsSection builds one team's open-draft section at the given ordinal. The
// bool is R-9.4: false means the team has no open drafts and the section is to
// be omitted, not rendered empty. An unknown team is an error, not an empty
// listing — "team `platfrom` has no drafts" is a true sentence about a team that
// does not exist and a useless one.
type DraftsSection func(ws *workspace.Workspace, team string, ordinal int) (
	model.GateResult, bool, error)
