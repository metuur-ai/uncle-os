package skills

import (
	"fmt"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/model"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/workspace"
)

// Gate 7's identity. The header text is frozen by every committed golden
// (examples/golden-validate.txt:38), so it lives beside the producer rather
// than being retyped at the validate call site.
const (
	GateSlug  = "skills-layering"
	GateTitle = "custom skills layering (shadowing + extends resolution)"
)

// ConflictKind is which of the two skill-layering rules a conflict breaks.
type ConflictKind string

const (
	// ConflictShadowing is GPF-R-5.2: a team or personal skill reusing a
	// canonical skill's id or name.
	ConflictShadowing ConflictKind = "shadowing"
	// ConflictDanglingExtends is GPF-R-5.3: `extends:` naming a base skill that
	// does not exist.
	ConflictDanglingExtends ConflictKind = "dangling-extends"
)

// ReasonKind is WHICH identity a shadowing skill reused. Python composes this
// into the fragment "id 'x'" or "name 'y'" inside the detection loop
// (`:852-855`); R-2.12 requires the two halves stay separate until a renderer
// puts them together.
type ReasonKind string

const (
	ReasonID   ReasonKind = "id"
	ReasonName ReasonKind = "name"
)

// Conflict is one structured skill-layering violation. It holds facts only: no
// field of it is a sentence, and nothing here is human-ordered prose.
type Conflict struct {
	Kind ConflictKind
	// Skill is the workspace-relative path of the offending skill. Set for both
	// kinds.
	Skill string

	// Shadows is the workspace-relative path of the canonical skill whose
	// identity was reused. ConflictShadowing only.
	Shadows string
	// Reason names which identity was reused, and ReasonValue is that identity.
	// ConflictShadowing only.
	Reason      ReasonKind
	ReasonValue string

	// Extends is the URI that resolved to nothing.
	// ConflictDanglingExtends only.
	Extends string
}

// Counts are the two totals gate 7's clean line reports (`:1085-1086`).
//
// Both are counted over ALL discovered skills, including the personal layer for
// Canonical — Python filters by authority, not by layer, so a team skill that
// declares `authority: canonical` counts here even though it is not eligible to
// BE shadowed (`:842-843` additionally requires the company or platform layer).
// Team is the layer count.
//
// The personal layer is deliberately absent from the reported totals: personal
// rules live in the git-ignored scratchpad, so counting them would make validate
// output depend on untracked files and drift on a fresh clone. Shadowing
// detection still scans them.
type Counts struct {
	Canonical int
	Team      int
}

// Count computes the gate's clean-line totals.
func Count(skills []Skill) Counts {
	var c Counts
	for _, s := range skills {
		if s.IsCanonical() {
			c.Canonical++
		}
		if s.Layer == LayerTeam {
			c.Team++
		}
	}
	return c
}

// Conflicts reports the skills gate's violations (skill_conflicts, `:837-866`).
// An empty result is the clean case; there is nothing to distinguish "no skills"
// from "no conflicts", and the gate treats both as passing.
//
// Emission order is the oracle's and is load-bearing: every shadowing conflict
// first, in (skill, canonical) discovery order, then every dangling extends.
func Conflicts(ws *workspace.Workspace, skills []Skill) ([]Conflict, error) {
	// Only company and platform skills can BE shadowed: a canonical team skill
	// is not an authority the layering rule protects.
	var canonical []Skill
	for _, s := range skills {
		if s.IsCanonical() && (s.Layer == LayerCompany || s.Layer == LayerPlatform) {
			canonical = append(canonical, s)
		}
	}

	var out []Conflict
	for _, s := range skills {
		if s.Layer != LayerTeam && s.Layer != LayerPersonal {
			continue
		}
		for _, k := range canonical {
			// Unreachable given the layer filters above, but it is the oracle's
			// own guard and costs one comparison to keep honest.
			if s.Path == k.Path {
				continue
			}
			c := Conflict{Kind: ConflictShadowing, Skill: s.Rel, Shadows: k.Rel}
			switch {
			// An id match wins over a name match; only one conflict is emitted
			// per (skill, canonical) pair even when both identities collide.
			case s.ID.Truthy && s.ID.Equal(k.ID):
				c.Reason, c.ReasonValue = ReasonID, s.ID.Text
			case s.Name == k.Name:
				c.Reason, c.ReasonValue = ReasonName, s.Name
			default:
				continue
			}
			out = append(out, c)
		}
	}

	for _, s := range skills {
		if !s.Extends.Truthy {
			continue
		}
		_, found, err := ResolveExtends(ws, s.Extends)
		if err != nil {
			return nil, err
		}
		if !found {
			out = append(out, Conflict{
				Kind:    ConflictDanglingExtends,
				Skill:   s.Rel,
				Extends: s.Extends.Text,
			})
		}
	}
	return out, nil
}

// Finding turns one Conflict into a record. Gate 7 uses no line prefix, so
// Subject is empty and the whole sentence is the message (LLD prefix table,
// row 7).
func (c Conflict) Finding() model.Finding {
	f := model.Finding{Severity: model.SevFail, Path: c.Skill}
	switch c.Kind {
	case ConflictShadowing:
		f.Code = model.CodeSkillShadowing
		f.Fields = model.Fields{
			"skill":       c.Skill,
			"shadows":     c.Shadows,
			"reason":      string(c.Reason),
			"reasonValue": c.ReasonValue,
		}
	case ConflictDanglingExtends:
		f.Code = model.CodeSkillDanglingExtends
		f.Fields = model.Fields{"skill": c.Skill, "extends": c.Extends}
	}
	f.Message = Message(f.Code, f.Fields)
	return f
}

// Gate runs the skills gate and returns its result (bin/company-os:1069-1088).
// Every fact validate needs is here; nothing about the gate requires reaching
// into this package's internals.
//
// Absence-tolerant, like the generated/ gates: no skills, or skills with no
// conflicts, passes with the clean line.
func Gate(ws *workspace.Workspace, ordinal int) (model.GateResult, error) {
	g := model.GateResult{Ordinal: ordinal, Slug: GateSlug, Title: GateTitle}
	skills, err := Discover(ws)
	if err != nil {
		return g, err
	}
	conflicts, err := Conflicts(ws, skills)
	if err != nil {
		return g, err
	}
	if len(conflicts) > 0 {
		for _, c := range conflicts {
			g.Findings = append(g.Findings, c.Finding())
		}
		return g, nil
	}
	counts := Count(skills)
	fields := model.Fields{"canonical": counts.Canonical, "team": counts.Team}
	g.Findings = []model.Finding{{
		Severity: model.SevOK,
		Code:     model.CodeSkillsClean,
		Message:  Message(model.CodeSkillsClean, fields),
		Fields:   fields,
	}}
	return g, nil
}

// Message composes a gate-7 finding's sentence from its code and its typed
// fields, and from nothing else.
//
// This is the only function in the package that produces human prose, and it is
// a pure function of (code, Fields) — the detection loops above concatenate
// nothing (R-2.8, R-2.12). It is what makes the R-2.3 claim testable rather
// than asserted: the two counts in the clean line are read back out of Fields
// as ints here, so changing Fields changes the rendered text and a text-only
// message could not exist.
//
// It is exported so that internal/render can call it (or replace it) without
// this package having to know which renderer is running.
func Message(code string, f model.Fields) string {
	switch code {
	case model.CodeSkillShadowing:
		return fmt.Sprintf(
			"skill shadowing: %s reuses the canonical %s '%s' of %s — extend it with "+
				"`extends: platform-skill://...` instead of replacing it",
			f.Str("skill"), f.Str("reason"), f.Str("reasonValue"), f.Str("shadows"))
	case model.CodeSkillDanglingExtends:
		return fmt.Sprintf(
			"dangling extends: %s declares extends: %s but no such base skill exists",
			f.Str("skill"), f.Str("extends"))
	case model.CodeSkillsClean:
		return fmt.Sprintf(
			"skills layered cleanly (%d canonical, %d team; no shadowing or dangling extends)",
			f.Int("canonical"), f.Int("team"))
	}
	return ""
}
