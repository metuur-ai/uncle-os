package graph

// validate's gates 5 and 6 — CLAUDE.md context-node drift (bin/company-os:997
// -1041) and feature-index drift (`:1044-1066`) — plus the two helpers gate 4
// needs from this package.
//
// Both gates compare a COMMITTED derived artifact against a fresh render from
// the same builder `graph build` writes with, which is what makes
// write-then-validate incapable of drifting. Neither writes anything.

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/model"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/workspace"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/yamlio"
)

// The two gates' identities. Header text is frozen by the golden snapshots.
const (
	NodeGateSlug  = "claude-node-drift"
	NodeGateTitle = "CLAUDE.md context node drift (fail-safe, absence-tolerant)"

	FeatureIndexGateSlug  = "feature-index-drift"
	FeatureIndexGateTitle = "feature-index drift (derived component->artifact map)"
)

// canonicalBlock is canonical_block (`:109-113`): newline-normalized, stripped,
// and trailing whitespace removed per line. Two spellings of one generated block
// compare equal, so re-indenting a node does not read as drift.
func canonicalBlock(text string) string {
	text = strings.ReplaceAll(strings.ReplaceAll(text, "\r\n", "\n"), "\r", "\n")
	lines := strings.Split(strings.TrimSpace(text), "\n")
	for i, l := range lines {
		lines[i] = strings.TrimRight(l, " \t\n\v\f\r")
	}
	return strings.Join(lines, "\n")
}

// BlocksEqual is blocks_equal (`:116-117`).
func BlocksEqual(a, b string) bool { return canonicalBlock(a) == canonicalBlock(b) }

// IdentityErrors is identity_errors (`:1543-1564`): `roster`, `channels` and
// `pointers` are optional, but a present one is shape-checked row by row.
//
// It returns the sentences rather than codes, exactly as PointerErrors above it
// does and for the same reason: model/codes.go declares ONE code for this render
// site (`:1025`), and each string is already the whole of one line's message.
func IdentityErrors(teamMeta yamlio.PyMap) []string {
	var errs []string
	errs = append(errs, rowErrors(teamMeta, "roster", "name", "role")...)
	errs = append(errs, rowErrors(teamMeta, "channels", "name", "id")...)
	return append(errs, PointerErrors(teamMeta)...)
}

// rowErrors is the shared body of the roster and channels checks: same list
// test, same per-row shape test, different key pair.
func rowErrors(meta yamlio.PyMap, key, first, second string) []string {
	v := meta.Get(key)
	// `if x is not None` — a present `roster: null` is skipped, unlike a present
	// empty list, which is iterated and yields nothing.
	if v == nil || yamlio.PyIsNone(v) {
		return nil
	}
	list, ok := v.(yamlio.PySeq)
	if !ok {
		return []string{key + ": must be a list"}
	}
	var errs []string
	for i, row := range list {
		m, ok := row.(yamlio.PyMap)
		if !ok || yamlio.PyFalsy(m.Get(first)) || yamlio.PyFalsy(m.Get(second)) {
			errs = append(errs, fmt.Sprintf("%s[%d]: needs '%s' and '%s'", key, i, first, second))
		}
	}
	return errs
}

// TagsInSync is gate 4's `sorted(meta.get("tags") or []) == derived` (`:1004`).
//
// The committed value is iterated with Python's `iter()` semantics (a str yields
// characters, a mapping yields keys) and then sorted. Only an all-str list can
// ever equal `derived`, which is a list of str, so a committed value carrying
// anything else is reported as drift without being sorted — Python reaches the
// same verdict, by comparing lists of different element types rather than by
// declining to sort. The one input where the two differ is a mixed-type list of
// two or more elements, which raises TypeError in Python and exits 1; this
// reports drift, which also exits 1.
func TagsInSync(meta yamlio.PyMap, derived []string) (bool, error) {
	items, err := pyIter(meta.Get("tags"), "tags")
	if err != nil {
		return false, err
	}
	if len(items) != len(derived) {
		return false, nil
	}
	committed := make([]string, 0, len(items))
	for _, it := range items {
		s, ok := it.(yamlio.PyStr)
		if !ok {
			return false, nil
		}
		committed = append(committed, string(s))
	}
	sort.Strings(committed)
	for i := range committed {
		if committed[i] != derived[i] {
			return false, nil
		}
	}
	return true, nil
}

// NodeGate is validate's gate 5 (`:997-1041`).
//
// Two things about it are load-bearing. The team identity check runs BEFORE the
// node check and against team.yaml rather than CLAUDE.md, so one root can emit a
// `<root>/team.yaml` finding immediately followed by a `<root>` one — that is
// the "gate 5 alone uses three prefix shapes" row of the LLD table, and all
// three are Subject values. And the node itself is absence-tolerant: a root with
// no CLAUDE.md, or one with no generated markers, PASSES, because the context
// node is opt-in enrichment.
func NodeGate(ws *workspace.Workspace, ordinal int) (model.GateResult, error) {
	g := model.GateResult{Ordinal: ordinal, Slug: NodeGateSlug, Title: NodeGateTitle}
	docs, err := IterGraphDocs(ws)
	if err != nil {
		return g, err
	}
	groups := GroupDocsByRoot(ws, docs)
	for _, root := range NodeRoots(ws) {
		relRoot := relTo(ws.Root, root)
		teamMeta := RootTeamMeta(ws, root)
		if teamMeta != nil {
			for _, ie := range IdentityErrors(teamMeta) {
				fields := model.Fields{"root": relRoot, "problem": ie}
				g.Findings = append(g.Findings, gateFinding(model.SevFail,
					model.CodeNodeIdentity, relRoot+"/team.yaml", fields))
			}
		}

		node := filepath.Join(root, "CLAUDE.md")
		if !exists(node) {
			g.Findings = append(g.Findings, gateFinding(model.SevOK,
				model.CodeNodeAbsent, relRoot, model.Fields{"root": relRoot}))
			continue
		}
		raw, readErr := os.ReadFile(node)
		if readErr != nil {
			return g, model.Errorf(model.ExitArtifact, "cannot read %s: %v", node, readErr)
		}
		fields := model.Fields{"root": relRoot, "path": relTo(ws.Root, node)}
		committed, marked := ExtractGeneratedBlock(readText(raw))
		if !marked {
			f := gateFinding(model.SevOK, model.CodeNodeHandOwned, relRoot+"/CLAUDE.md", fields)
			f.Path = fields.Str("path")
			g.Findings = append(g.Findings, f)
			continue
		}
		fresh, err := BuildClaudeNode(ws, root, groups[root], teamMeta, teamMeta != nil)
		if err != nil {
			return g, err
		}
		code := model.CodeNodeInSync
		sev := model.SevOK
		if !BlocksEqual(committed, fresh) {
			code, sev = model.CodeNodeDrift, model.SevFail
		}
		f := gateFinding(sev, code, relRoot+"/CLAUDE.md", fields)
		f.Path = fields.Str("path")
		g.Findings = append(g.Findings, f)
	}
	return g, nil
}

// FeatureIndexGate is validate's gate 6 (`:1044-1066`).
//
// Absence-tolerant like gate 5. The drift check `continue`s on a mismatch, so a
// drifted index is never also reported as unresolved — the references it names
// are not the current derivation's and saying anything about them would be
// noise.
func FeatureIndexGate(ws *workspace.Workspace, ordinal int) (model.GateResult, error) {
	g := model.GateResult{Ordinal: ordinal, Slug: FeatureIndexGateSlug, Title: FeatureIndexGateTitle}
	for _, pdir := range ws.AllPlatforms() {
		pname := filepath.Base(pdir)
		committedPath := filepath.Join(pdir, "generated", "feature-index.yaml")
		fresh, err := BuildFeatureIndex(ws, pdir)
		if err != nil {
			return g, err
		}
		fields := model.Fields{"platform": pname, "path": relTo(ws.Root, committedPath)}
		if !exists(committedPath) {
			g.Findings = append(g.Findings, gateFinding(model.SevOK,
				model.CodeFeatureIndexAbsent, pname, fields))
			continue
		}
		committed, err := loadMapping(committedPath)
		if err != nil {
			return g, err
		}
		same, err := sameCanonical(committed, fresh)
		if err != nil {
			return g, err
		}
		if !same {
			g.Findings = append(g.Findings, gateFinding(model.SevFail,
				model.CodeFeatureIndexDrift, pname, fields))
			continue
		}
		unresolved := FeatureIndexUnresolved(ws, fresh)
		if len(unresolved) > 0 {
			for _, u := range unresolved {
				uf := model.Fields{
					"platform": pname, "component": u.Component,
					"kind": u.Kind, "reference": u.ID,
				}
				g.Findings = append(g.Findings, gateFinding(model.SevFail,
					model.CodeFeatureIndexUnresolved, pname, uf))
			}
			continue
		}
		components, _ := fresh.Get("components").(yamlio.PyMap)
		okFields := model.Fields{"platform": pname, "components": len(components)}
		g.Findings = append(g.Findings, gateFinding(model.SevOK,
			model.CodeFeatureIndexInSync, pname, okFields))
	}
	return g, nil
}

// sameCanonical is `canonical_yaml(a) != canonical_yaml(b)` (`:1053`): a
// key-sorted structural compare, so two spellings of one document agree. It is
// the same compare WriteFeatureIndexes uses to decide whether to write, which is
// what keeps the writer and this gate from disagreeing.
func sameCanonical(a, b yamlio.PyValue) (bool, error) {
	left, err := yamlio.PyDumpCanonical(a)
	if err != nil {
		return false, model.Errorf(model.ExitArtifact, "cannot serialize feature index: %v", err)
	}
	right, err := yamlio.PyDumpCanonical(b)
	if err != nil {
		return false, model.Errorf(model.ExitArtifact, "cannot serialize feature index: %v", err)
	}
	return left == right, nil
}

// gateFinding builds one record, composing its sentence through Message and
// nowhere else.
func gateFinding(sev model.Severity, code, subject string, f model.Fields) model.Finding {
	return model.Finding{
		Severity: sev,
		Code:     code,
		Subject:  subject,
		Message:  Message(code, f),
		Fields:   f,
	}
}

// Message composes the gate sentences this package owns — gate 5, gate 6, and
// the three of gate 4's four codes that are about tags and pointers rather than
// about the process contract — from a code and its typed fields, and from
// nothing else.
//
// Gate 4's fourth code, CodeFrontmatterCoreField, is deliberately absent:
// core_field_errors is internal/product's, so its five sentences are
// product.Message's and internal/validate hands each Issue straight to it. A
// second copy here is how the doc-level gate and `prd validate` would start
// describing the same violation differently.
//
// It is exported so that internal/render can call it (or replace it) without
// this package having to know which renderer is running.
func Message(code string, f model.Fields) string {
	switch code {
	// ------------------------------------------------------------- gate 4
	case model.CodeTagsDrift:
		return "committed tags drifted from frontmatter derivation " +
			"— run: company-os graph build"
	case model.CodeFrontmatterInSync:
		return "core fields + tags in sync"
	case model.CodePointerGuidance:
		return f.Str("problem") + " — pointer guidance (not blocking)"

	// ------------------------------------------------------------- gate 5
	case model.CodeNodeIdentity:
		return f.Str("problem")
	case model.CodeNodeAbsent:
		return "no CLAUDE.md node (absent -> pass)"
	case model.CodeNodeHandOwned:
		return "hand-owned, no generated markers (-> pass)"
	case model.CodeNodeDrift:
		return "generated block drifted — run: company-os graph build"
	case model.CodeNodeInSync:
		return "context node in sync"

	// ------------------------------------------------------------- gate 6
	case model.CodeFeatureIndexAbsent:
		return "no feature-index (absent -> pass; run graph build to enable)"
	case model.CodeFeatureIndexDrift:
		return "feature-index drifted from derivation — run: company-os graph build"
	case model.CodeFeatureIndexUnresolved:
		return fmt.Sprintf(
			"feature-index component '%s' references %s '%s' which resolves to no document",
			f.Str("component"), f.Str("kind"), f.Str("reference"))
	case model.CodeFeatureIndexInSync:
		return fmt.Sprintf("feature-index in sync (%d component(s))", f.Int("components"))
	}
	return ""
}
