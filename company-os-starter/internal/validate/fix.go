package validate

// FixDerived is the body of `validate --fix` (ux-simplification 2.1): the
// regeneration step that runs BEFORE the gates so a workspace whose derived
// state has drifted — a hand-edited tag, a stale effective-governance after a
// deviation declaration, a deleted per-directory index — is brought back in
// sync without the user remembering `governance resolve` and `graph build`.
//
// It reuses the SAME code paths those two commands use — governance.Resolve
// for each team, graph.Rebuild for tags + feature-indexes + per-directory
// indexes + CLAUDE.md nodes — and never introduces a new writer. The
// acceptance double-build (§4 of acceptance.sh) already proves these paths are
// idempotent; --fix is the same derivation invoked opportunistically.
//
// Federated slices are untouched: graph.Rebuild already skips knowledge/
// (iterGraphDocs excludes it) and WriteIndexes skips manifest-declared slice
// roots, so --fix cannot rewrite synced content. Gate 9 still fails on a
// hand-edited slice after --fix, which is the design intent.
//
// The return value is the count of files whose bytes actually changed on disk.
// A second --fix run on a clean tree returns 0 — idempotence made observable.

import (
	"path/filepath"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/governance"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/graph"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/model"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/workspace"
)

// FixDerived regenerates every derived artifact the acceptance double-build
// already proves is derived, and returns how many files changed on disk.
func FixDerived(ws *workspace.Workspace) (int, error) {
	changed := 0

	// 1. effective-governance.yaml for every team. governance.Resolve already
	//    implements the semantic guard (writeGuarded neutralizes generatedAt),
	//    so Written is true only when the DERIVED CONTENT changed — exactly
	//    the question --fix needs answered.
	for _, tdir := range ws.AllTeams() {
		teamID := filepath.Base(tdir)
		resolved, err := governance.Resolve(ws, teamID)
		if err != nil {
			return changed, err
		}
		if resolved.Written {
			changed++
		}
	}

	// 2. Tags, feature-indexes, per-directory indexes, CLAUDE.md nodes — the
	//    four derivations graph.Rebuild performs, through the same retag +
	//    rebuild code path graph build uses. Each finding that carries one of
	//    the "written" or "removed" codes names a file whose bytes changed.
	sections, err := graph.Rebuild(ws)
	if err != nil {
		return changed, err
	}
	for _, s := range sections {
		for _, f := range s.Findings {
			if writeCode(f.Code) {
				changed++
			}
		}
	}

	return changed, nil
}

// writeCode reports whether a finding code names a derivation that changed a
// file on disk. The five codes are the ones graph.Rebuild emits for actual
// writes; the "hand-owned" and "in-sync" codes are skips, not writes.
func writeCode(code string) bool {
	switch code {
	case model.CodeGraphTagged,
		model.CodeGraphIndexWritten,
		model.CodeGraphNodeWritten,
		model.CodeGraphDirIndexWritten,
		model.CodeGraphDirIndexRemoved:
		return true
	}
	return false
}
