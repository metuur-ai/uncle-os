package validate

// The promotion-integrity gate (R-7.1 … R-7.8).
//
// Promotion copies a draft forward rather than moving it (see
// docs/lld/team-drafted-prd-promotion.md, "Copy-forward, not move"), so from the
// moment `prd promote` returns there are two documents that can drift. The
// digest written into the record's `promotedFrom` block is the only thing that
// links them, and a link nobody re-checks rots silently. This gate is the
// re-check.
//
// It lives here rather than in internal/product for the same reason gate 4 does:
// it spans two clusters. It reads internal/product's draft residency and digest
// helpers and internal/workspace's platform layout, and it composes a gate out of
// both. Nothing about it belongs to either package alone.
//
// Three rules shape the whole file:
//
//   - It recomputes with product.DraftDigestFile and NOTHING else (R-6.4). A
//     second spelling of "normalized form" here would turn the gate into a source
//     of false failures the day either side changed.
//   - It is silent on `status: draft` and `status: abandoned` (R-7.7). Only a
//     promoted draft has a link to check, and the common case is a workspace full
//     of open drafts that this gate must not narrate.
//   - "Not found" and "found but unlinked" are different findings (R-7.4 vs
//     R-7.5). The naive implementation reports an absent `promotedFrom` as a
//     digest mismatch against the empty string, which is a true statement that
//     sends the reader to look for an edit that never happened.

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/graph"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/model"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/product"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/workspace"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/yamlio"
)

// PromotionGateSlug and PromotionGateTitle name the gate. Like gate 4's
// identity, they live in this package because this package is what produces it.
const (
	PromotionGateSlug  = "promotion-integrity"
	PromotionGateTitle = "promotion integrity (promoted drafts match their change records)"
)

// The gate's finding codes are model.CodePromotion*, in model/codes.go with
// every other gate's. Only the identity above is local — the same split gate 4
// uses, and for the same reason: the codes are part of the --json surface the
// renderers switch on, the slug is not.

// statusPromoted is the one draft status this gate speaks about. `draft` and
// `abandoned` are matched by exclusion (R-7.7) rather than enumerated, so a
// status this unit has never heard of is also silent rather than crashing the
// gate on someone else's vocabulary.
const statusPromoted = "promoted"

// promotionGate walks every team's drafts and checks each promoted one against
// the record it says it became.
//
// A team with no draft directory contributes nothing (R-1.3 via DraftIDs), and a
// workspace with no promoted drafts contributes no findings at all — the header
// with nothing under it, which is R-7.8 and needs no code of its own because
// GateResult already models "ran, found nothing".
func promotionGate(ws *workspace.Workspace, ordinal int) (model.GateResult, error) {
	g := model.GateResult{Ordinal: ordinal, Slug: PromotionGateSlug, Title: PromotionGateTitle}
	for _, tdir := range ws.AllTeams() {
		team := filepath.Base(tdir)
		draftDir, err := product.DraftDir(ws, team)
		if err != nil {
			// AllTeams only yields directories that exist, so this cannot fire for
			// a team it just listed. It is returned rather than swallowed because a
			// workspace that contradicts itself mid-gate is an abort, not a finding.
			return g, err
		}
		for _, id := range product.DraftIDs(draftDir) {
			path := product.DraftPath(draftDir, id)
			meta, _, err := graph.ReadFrontmatter(path)
			if err != nil {
				return g, err
			}
			if metaString(meta, "status") != statusPromoted {
				continue
			}
			f, err := checkPromotedDraft(ws, team, id, path, meta)
			if err != nil {
				return g, err
			}
			g.Findings = append(g.Findings, f)
		}
	}
	return g, nil
}

// checkPromotedDraft is the four-way decision for one promoted draft, in the
// order the failures shadow each other: a target that does not resolve cannot be
// read, and a record with no provenance cannot be compared.
func checkPromotedDraft(
	ws *workspace.Workspace, team, id, path string, meta yamlio.PyMap,
) (model.Finding, error) {
	to, _ := product.ReadPromotedTo(meta)
	fields := model.Fields{
		"team":     team,
		"draft":    id,
		"path":     rel(ws.Root, path),
		"platform": to.Platform,
		"record":   to.Record,
	}
	subject := team + "/" + id

	// R-7.4. An absent or half-written `promotedTo` lands here too: a draft that
	// claims to be promoted and cannot say where is exactly as unresolved as one
	// naming a record that was deleted, and the reader's next move is the same.
	recordPath, location, ok := resolveRecord(ws, to)
	if !ok {
		fields["location"] = ""
		return promotionFinding(model.SevFail, model.CodePromotionTargetUnresolved, subject, fields), nil
	}
	fields["location"] = location
	fields["recordPath"] = rel(ws.Root, recordPath)

	recordMeta, _, err := graph.ReadFrontmatter(recordPath)
	if err != nil {
		return model.Finding{}, err
	}

	// R-7.5. Both the absent block and the present-but-digestless block are the
	// same event — provenance was never written — and both are repaired by
	// re-running the same command, so they are one finding rather than two.
	from, present := product.ReadPromotedFrom(recordMeta)
	if !present || from.Digest == "" {
		return promotionFinding(model.SevFail, model.CodePromotionInterrupted, subject, fields), nil
	}
	fields["recorded"] = from.Digest

	digest, err := product.DraftDigestFile(path)
	if err != nil {
		return model.Finding{}, err
	}
	fields["digest"] = digest

	// R-7.3 and, when they agree, R-7.2/R-7.6 — the archived record passes on
	// exactly the same comparison as the active one, which is why `location` is a
	// field and not a branch.
	if digest != from.Digest {
		return promotionFinding(model.SevFail, model.CodePromotionDigestDrift, subject, fields), nil
	}
	return promotionFinding(model.SevOK, model.CodePromotionInSync, subject, fields), nil
}

// resolveRecord locates the change record named by `promotedTo`, active first
// and archived second (R-7.6).
//
// The order is the lifecycle's: `prd complete` moves a record from active to
// archive, so a record found in active is the live one and a record found only
// under archive/prds/ has been completed. Both are legitimate targets.
func resolveRecord(ws *workspace.Workspace, to product.PromotedTo) (path, location string, ok bool) {
	if to.Platform == "" || to.Record == "" {
		return "", "", false
	}
	pdir, err := ws.PlatformDir(to.Platform)
	if err != nil {
		// A record inside a platform the workspace does not have is unresolved,
		// not a workspace error: the gate is reporting on the draft, and the
		// missing platform is the reason it cannot resolve.
		return "", "", false
	}
	candidates := []struct {
		path     string
		location string
	}{
		{filepath.Join(pdir, "change-records", "active", to.Record, "prd.md"), "active"},
		{filepath.Join(pdir, "archive", "prds", to.Record, "prd.md"), "archived"},
	}
	for _, c := range candidates {
		if _, err := os.Stat(c.path); err == nil {
			return c.path, c.location, true
		}
	}
	return "", "", false
}

// promotionFinding assembles one record. Path is the DRAFT's path in every case,
// including the ones whose sentence is about the record: the draft is the
// document the reader has to act on, and the record's path rides in Fields.
func promotionFinding(sev model.Severity, code, subject string, f model.Fields) model.Finding {
	return model.Finding{
		Severity: sev,
		Code:     code,
		Subject:  subject,
		Path:     f.Str("path"),
		Message:  PromotionMessage(code, f),
		Fields:   f,
	}
}

// PromotionMessage composes a promotion-integrity finding's sentence from its
// code and its typed fields, and from nothing else — the same contract every
// other gate's Message honours, so the text is a function of the record and a
// text-only finding cannot exist.
func PromotionMessage(code string, f model.Fields) string {
	target := f.Str("platform") + "/" + f.Str("record")
	switch code {
	case model.CodePromotionTargetUnresolved:
		if f.Str("platform") == "" || f.Str("record") == "" {
			return "promoted draft records no target: promotedTo names neither a platform " +
				"nor a change record, so the promotion cannot be verified"
		}
		return fmt.Sprintf(
			"unresolved target: promotedTo names change record '%s', which exists at "+
				"neither change-records/active/%s/ nor archive/prds/%s/",
			target, f.Str("record"), f.Str("record"))
	case model.CodePromotionInterrupted:
		return fmt.Sprintf(
			"interrupted promotion: change record '%s' carries no promotedFrom — "+
				"re-run `company-os prd promote --team %s %s`",
			target, f.Str("team"), f.Str("draft"))
	case model.CodePromotionDigestDrift:
		return fmt.Sprintf(
			"digest drift: draft no longer matches change record '%s' "+
				"(recorded %s, recomputed %s)",
			target, f.Str("recorded"), f.Str("digest"))
	case model.CodePromotionInSync:
		return fmt.Sprintf("matches %s change record '%s'", f.Str("location"), target)
	}
	return ""
}

// metaString reads a top-level scalar off loaded frontmatter as a Go string,
// with an absent or non-string value reading as "" rather than Python's "None",
// because the result is compared for equality against a status word.
func metaString(meta yamlio.PyMap, key string) string {
	s, ok := meta.Get(key).(yamlio.PyStr)
	if !ok {
		return ""
	}
	return string(s)
}

// rel is the workspace-relative form of an absolute path, falling back to the
// absolute path when the two share no prefix. Findings carry relative paths so a
// golden does not depend on where the workspace was checked out.
func rel(root, p string) string {
	r, err := filepath.Rel(root, p)
	if err != nil {
		return p
	}
	return r
}
