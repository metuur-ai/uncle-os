package product

// Promotion target resolution (R-4.1 … R-4.5) and the provenance vocabulary
// promotion writes and the integrity gate reads (R-6.1, R-6.2, R-6.3, R-6.4,
// R-6.6).
//
// Nothing here writes a file. Target resolution is one of the read-only steps
// the effect ordering in docs/lld/team-drafted-prd-promotion.md puts ahead of
// every mutation (R-4.6), and the digest helper is a pure function of bytes so
// the promotion path and the gate path cannot answer it differently (R-6.4).

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/frontmatter"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/graph"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/model"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/workspace"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/yamlio"
)

// Frontmatter keys this unit owns. They are named once so the writer, the
// reader and the gate cannot disagree about spelling.
const (
	// KeyPromoteTo is the draft's routing block. Its only read key is
	// `platform`; `promoteTo.scope` is reserved for a later increment and is
	// deliberately neither written nor read (R-4.5).
	KeyPromoteTo = "promoteTo"
	// KeyPromotedFrom is written onto the promoted record (R-6.1).
	KeyPromotedFrom = "promotedFrom"
	// KeyPromotedTo is written back onto the draft (R-6.2).
	KeyPromotedTo = "promotedTo"

	// placeholder is the scaffolded value every unfilled draft field carries.
	// `promoteTo.platform: TODO` means "not yet decided", not "a platform
	// called TODO" (R-4.2).
	placeholder = "TODO"
)

// PromoteTarget is the resolved destination of a promotion: the platform the
// draft nominated, and the catalog directory it resolves to.
type PromoteTarget struct {
	// Platform is the id read from `promoteTo.platform` (R-4.1).
	Platform string
	// Dir is `platforms/<Platform>`, already proven to exist.
	Dir string
}

// ResolvePromoteTarget answers where a draft wants to go, or refuses.
//
// meta is the draft's loaded frontmatter and path is the draft file, used only
// to locate the refusal for the reader. The three refusals are distinct on
// purpose:
//
//   - R-4.2 absent or still `TODO` — an artifact error (exit 4). The file is
//     readable; what it says is incomplete.
//   - R-4.3 names a platform the workspace does not have — a workspace error
//     (exit 3) naming the unresolved id, raised by PlatformDir itself so the
//     wording matches every other failed platform lookup.
//   - R-4.4 disagrees with the frontmatter `platform` — an artifact error
//     reporting BOTH values, because the fix requires knowing which one is
//     wrong.
//
// The disagreement check is skipped when the frontmatter `platform` is itself
// unset or still `TODO`: that is the process-contract omission Unit 5 reports
// (with every other missing field in one pass), and reporting it here as a
// disagreement would name `TODO` as if it were a platform id. Nothing is
// written between the two checks, so the draft is still refused (R-4.6).
func ResolvePromoteTarget(ws *workspace.Workspace, meta yamlio.PyMap, path string) (PromoteTarget, error) {
	target := promoteToPlatform(meta)
	if IsUnset(target) {
		return PromoteTarget{}, model.Errorf(model.ExitArtifact,
			"%s: draft declares no target platform; set promoteTo.platform", path)
	}
	dir, err := ws.PlatformDir(target)
	if err != nil {
		return PromoteTarget{}, err
	}
	if declared := strOf(meta, "platform"); !IsUnset(declared) && declared != target {
		return PromoteTarget{}, model.Errorf(model.ExitArtifact,
			"%s: promoteTo.platform is '%s' but platform is '%s'; they must agree",
			path, target, declared)
	}
	return PromoteTarget{Platform: target, Dir: dir}, nil
}

// promoteToPlatform reads `promoteTo.platform` (R-4.1). A `promoteTo` that is
// absent, falsy, or not a mapping yields "", which the caller treats as unset —
// there is no separate diagnostic for a malformed block, because the remedy is
// the same sentence either way. `promoteTo.scope` is never consulted (R-4.5).
func promoteToPlatform(meta yamlio.PyMap) string {
	block, ok := meta.Get(KeyPromoteTo).(yamlio.PyMap)
	if !ok {
		return ""
	}
	if block.Get("platform") == nil {
		return ""
	}
	return strings.TrimSpace(yamlio.PyString(block.Get("platform")))
}

// IsUnset is "absent or still the scaffolded placeholder".
//
// It is exported for R-9.3: `today` has to decide the same question promotion
// decides, about the same fields, and two answers to "is this filled in yet?"
// would let a draft read as ready in one view and unfinished in the other.
func IsUnset(v string) bool { return v == "" || v == placeholder || v == "None" }

// ---------------------------------------------------------------- provenance

// PromotedFrom is the `promotedFrom` block the promoted record carries: where
// it came from, and what the draft looked like when it left (R-6.1).
type PromotedFrom struct {
	// Team is the origin team id.
	Team string
	// Draft is the origin draft id.
	Draft string
	// Digest is DraftDigest's output, `sha256:<hex>`.
	Digest string
}

// PyMap renders the block for writing, in the authored key order.
func (p PromotedFrom) PyMap() yamlio.PyMap {
	return yamlio.PyMap{
		{K: "team", V: yamlio.PyStr(p.Team)},
		{K: "draft", V: yamlio.PyStr(p.Draft)},
		{K: "digest", V: yamlio.PyStr(p.Digest)},
	}
}

// ReadPromotedFrom reads the block back off a loaded record.
//
// R-6.6: a record that was never promoted simply has no such block, which is
// not a finding. ok reports presence; it is never an error, so no caller can
// accidentally make the key required on hand-authored records. A present but
// malformed block reads as ok with empty fields, which the gate reports on its
// own terms.
func ReadPromotedFrom(meta yamlio.PyMap) (PromotedFrom, bool) {
	block, ok := meta.Get(KeyPromotedFrom).(yamlio.PyMap)
	if !ok {
		return PromotedFrom{}, false
	}
	return PromotedFrom{
		Team:   blockString(block, "team"),
		Draft:  blockString(block, "draft"),
		Digest: blockString(block, "digest"),
	}, true
}

// PromotedTo is the `promotedTo` block written back onto the draft alongside
// `status: promoted` (R-6.2).
type PromotedTo struct {
	// Platform is the target platform id.
	Platform string
	// Record is the id of the change record the draft became.
	Record string
}

// PyMap renders the block for writing, in the authored key order.
func (p PromotedTo) PyMap() yamlio.PyMap {
	return yamlio.PyMap{
		{K: "platform", V: yamlio.PyStr(p.Platform)},
		{K: "record", V: yamlio.PyStr(p.Record)},
	}
}

// ReadPromotedTo reads the block back off a loaded draft. Absence is not a
// finding (R-6.6); see ReadPromotedFrom.
func ReadPromotedTo(meta yamlio.PyMap) (PromotedTo, bool) {
	block, ok := meta.Get(KeyPromotedTo).(yamlio.PyMap)
	if !ok {
		return PromotedTo{}, false
	}
	return PromotedTo{
		Platform: blockString(block, "platform"),
		Record:   blockString(block, "record"),
	}, true
}

// blockString is `str(block.get(key))` with an absent key reading as "" rather
// than Python's "None", because these are ids compared for equality.
func blockString(block yamlio.PyMap, key string) string {
	v := block.Get(key)
	if v == nil {
		return ""
	}
	return yamlio.PyString(v)
}

// -------------------------------------------------------------------- digest

// DigestPrefix labels the hash so a later algorithm change is visible in the
// artifact rather than silent.
const DigestPrefix = "sha256:"

// DraftDigest is THE digest of a draft (R-6.3). Promotion calls it to write
// `promotedFrom.digest`; the promotion-integrity gate calls the same function
// to recompute it. There is deliberately no second implementation — two
// spellings of "normalized form" would diverge and turn the gate into a source
// of false failures (R-6.4).
//
// data is the raw file bytes, taken AFTER derived artifacts have been rebuilt.
func DraftDigest(data []byte) string {
	sum := sha256.Sum256(NormalizeForDigest(data))
	return DigestPrefix + hex.EncodeToString(sum[:])
}

// DraftDigestFile is DraftDigest over the contents of path. A read failure is
// an artifact error, since the caller has already established the file is the
// one it means to digest.
func DraftDigestFile(path string) (string, error) {
	doc, err := frontmatter.ParseFile(path)
	if err != nil {
		return "", model.Errorf(model.ExitArtifact, "%v", err)
	}
	return DraftDigest(reassemble(doc)), nil
}

// NormalizeForDigest is the normalized form the digest is taken over: LF line
// endings, per-line trailing whitespace stripped, no trailing blank lines, and
// the derived `tags:` block excluded (R-6.3).
//
// Excluding tags is the load-bearing part. `graph build` re-derives that block
// on every run and may emit the same tags with different bytes; including them
// would report a false drift on a draft nobody touched. The exclusion is
// semantic — the key is dropped from the parsed frontmatter, which is then
// re-emitted — rather than a line-slicing trick, because the emitter is free to
// wrap a long flow sequence across lines and a textual rule would then eat the
// next key.
//
// The re-emission runs whether or not the document HAS tags. Otherwise a draft
// digested before its first `graph build` and the same draft digested after it
// would take two different paths through this function and disagree for a
// reason that has nothing to do with its content. A document whose frontmatter
// will not parse is normalized as plain text; the digest still round-trips, and
// the malformed frontmatter is somebody else's diagnostic.
func NormalizeForDigest(data []byte) []byte {
	if stripped, ok := withoutTags(data); ok {
		data = stripped
	}
	lines := strings.Split(strings.ReplaceAll(strings.ReplaceAll(
		string(data), "\r\n", "\n"), "\r", "\n"), "\n")
	for i, l := range lines {
		lines[i] = strings.TrimRight(l, " \t")
	}
	for len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	if len(lines) == 0 {
		return nil
	}
	return []byte(strings.Join(lines, "\n") + "\n")
}

// withoutTags re-emits the document with the `tags:` key removed. ok is false
// only when there is no frontmatter to canonicalize, or it will not parse, in
// which case the caller normalizes the original bytes.
func withoutTags(data []byte) ([]byte, bool) {
	doc, err := frontmatter.Parse(data)
	if err != nil || !doc.HasFrontmatter {
		return nil, false
	}
	v, err := yamlio.PyLoadBytes(doc.YAML, "")
	if err != nil {
		return nil, false
	}
	meta, ok := v.(yamlio.PyMap)
	if !ok {
		return nil, false
	}
	kept := make(yamlio.PyMap, 0, len(meta))
	for _, p := range meta {
		if p.K != "tags" {
			kept = append(kept, p)
		}
	}
	fm, err := yamlio.PyDumpAutoFlow(kept)
	if err != nil {
		return nil, false
	}
	return []byte("---\n" + strings.TrimSpace(fm) + "\n---\n" + string(doc.Body)), true
}

// reassemble puts a parsed document back together in the form the digest is
// taken over. Parsing first is what makes DraftDigestFile agree with
// DraftDigest over the same file's raw bytes on a CRLF artifact: Parse applies
// the newline translation the frontmatter contract depends on.
func reassemble(doc frontmatter.Document) []byte {
	if !doc.HasFrontmatter {
		return doc.Body
	}
	out := make([]byte, 0, len(doc.YAML)+len(doc.Body)+9)
	out = append(out, "---\n"...)
	out = append(out, doc.YAML...)
	out = append(out, "\n---\n"...)
	return append(out, doc.Body...)
}

// ----------------------------------------------------------------- promote

// The three lifecycle values promotion moves between. `proposed` is what a
// hand-scaffolded change record opens with, which is the whole of R-5.2: a
// promoted record enters the platform lifecycle at the same point, so nothing
// downstream can tell the two apart (R-5.13 … R-5.17).
const (
	statusDraft    = "draft"
	statusProposed = "proposed"
	statusPromoted = "promoted"
)

// PRDPromote is `prd promote --team <t> <draft-id>` (R-4.6, R-5.1 … R-5.12,
// R-6.1, R-6.2, R-6.5).
//
// Steps 1-6 are read-only and step 7 onward mutate, in the order R-5.9 fixes:
//
//  7. write platforms/<p>/change-records/active/<id>/prd.md, status: proposed
//  8. rewrite the draft: status -> promoted, add promotedTo
//  9. rebuild derived artifacts over the workspace
//  10. digest the draft as it now stands on disk
//  11. patch the record with promotedFrom{team, draft, digest}
//  12. append the dated entry to teams/<t>/log.md
//
// The ordering is not stylistic. `graph build` derives a `status/` facet, and
// step 8 changes `status`, so the rebuild in step 9 is GUARANTEED to rewrite the
// draft. A digest taken before it would be recorded stale and the integrity gate
// would fail every successful promotion, without exception.
//
// Nothing is written before step 7 (R-4.6): a half-promoted workspace is the one
// state re-running cannot repair, so every refusal below happens while the tree
// is still untouched.
func PRDPromote(ws *workspace.Workspace, team, draftID string, rebuild Rebuild) (
	[]model.GateResult, error) {

	if team == "" {
		// Same reason as DraftNew: a draft is team-private, so there is no run of
		// this command without a team, and --team is not argparse-required on the
		// `prd` sub-parser.
		return nil, model.Usagef("prd", "the following arguments are required: --team")
	}
	if err := requirePRDID(draftID); err != nil {
		return nil, err
	}
	draftDir, err := DraftDir(ws, team)
	if err != nil {
		return nil, err
	}
	path := DraftPath(draftDir, draftID)
	if _, err := os.Stat(path); err != nil {
		return nil, model.Errorf(model.ExitWorkspace, "no draft at %s", path)
	}
	meta, body, err := graph.ReadFrontmatter(path)
	if err != nil {
		return nil, err
	}
	rel := relTo(ws.Root, path)

	// 1. R-5.6. A draft already promoted, or abandoned, is not a thing to
	// promote; the status it actually carries is what tells the reader which.
	if status := strOf(meta, "status"); status != statusDraft {
		return nil, model.Errorf(model.ExitConflict,
			"%s: status is '%s', not 'draft'", rel, status)
	}

	// 2-3. R-4.1 … R-4.4, all of them refusals that write nothing.
	target, err := ResolvePromoteTarget(ws, meta, rel)
	if err != nil {
		return nil, err
	}

	// 4-5. R-5.3 and R-5.4 accumulate rather than short-circuit, so one run
	// reports everything the author has left to do.
	if issues := readinessIssues(meta, body); len(issues) > 0 {
		s := model.GateResult{Ordinal: 1, Slug: model.SlugPRDPromote, Title: draftID}
		s.Findings = append(s.Findings, okFinding(model.CodePromoteNotReady, "", rel,
			model.Fields{"draft": draftID, "team": team, "path": rel}))
		for _, i := range issues {
			s.Findings = append(s.Findings, i.finding(model.SevFail))
		}
		return []model.GateResult{s}, ErrNotReady
	}

	// 6. R-5.7, with R-5.11's exception.
	dir := filepath.Join(target.Dir, "change-records", "active", draftID)
	record := filepath.Join(dir, "prd.md")
	if _, err := os.Stat(record); err == nil {
		recordMeta, _, err := graph.ReadFrontmatter(record)
		if err != nil {
			return nil, err
		}
		// R-5.11. The predicate is the integrity gate's, deliberately: what the
		// gate calls an interrupted promotion is exactly what re-running has to
		// repair, and two spellings of that state would leave a workspace the
		// gate reports on and the command refuses to fix.
		if from, ok := ReadPromotedFrom(recordMeta); ok && from.Digest != "" {
			return nil, model.Errorf(model.ExitConflict, "%s already exists",
				relTo(ws.Root, record))
		}
	}

	// 7. Seeded from the draft's body, verbatim (R-5.1).
	if err := os.MkdirAll(dir, 0o777); err != nil {
		return nil, model.Errorf(model.ExitArtifact, "cannot create %s: %v", dir, err)
	}
	if err := writeArtifact(record, promotedRecordMeta(meta), body); err != nil {
		return nil, err
	}

	// 8. R-5.8 and R-6.2: the draft is rewritten in place — not deleted, not
	// relocated, not truncated — so the team keeps its record of what it
	// proposed and the two documents can be compared later.
	to := PromotedTo{Platform: target.Platform, Record: draftID}
	draftMeta := meta.Set("status", yamlio.PyStr(statusPromoted)).Set(KeyPromotedTo, to.PyMap())
	if err := writeArtifact(path, draftMeta, body); err != nil {
		return nil, err
	}

	// 9. R-5.10: `validate` is green on return, with no separate `graph build`.
	var derived []model.GateResult
	if rebuild != nil {
		if derived, err = rebuild(ws); err != nil {
			return nil, err
		}
	}

	// 10-11. R-6.3: the digest is of the draft AS IT NOW STANDS, after the
	// rebuild has re-derived its tags.
	digest, err := DraftDigestFile(path)
	if err != nil {
		return nil, err
	}
	recordMeta, recordBody, err := graph.ReadFrontmatter(record)
	if err != nil {
		return nil, err
	}
	from := PromotedFrom{Team: team, Draft: draftID, Digest: digest}
	if err := writeArtifact(record, recordMeta.Set(KeyPromotedFrom, from.PyMap()), recordBody); err != nil {
		return nil, err
	}

	// 12. R-6.5.
	tdir, err := ws.TeamDir(team)
	if err != nil {
		return nil, err
	}
	log := filepath.Join(tdir, "log.md")
	if err := appendPromotionLog(log, today().Format(isoDate), draftID, target.Platform, draftID); err != nil {
		return nil, err
	}

	relRecord, relLog := relTo(ws.Root, record), relTo(ws.Root, log)
	s := model.GateResult{Ordinal: 1, Slug: model.SlugPRDPromote, Title: draftID}
	s.Findings = append(s.Findings,
		okFinding(model.CodePRDPromoted, "", relRecord,
			model.Fields{"path": relRecord, "prd": draftID, "platform": target.Platform}),
		okFinding(model.CodePRDDraftLinked, "", rel,
			model.Fields{"team": team, "draft": draftID, "path": rel,
				"platform": target.Platform, "prd": draftID, "digest": digest}),
		okFinding(model.CodeLogAppended, "", relLog, model.Fields{"path": relLog}))
	out := append([]model.GateResult{s}, derived...)
	// R-5.12. Like `prd complete`, the next-step line closes the whole run rather
	// than the command's own half, so it reads last after the derived output.
	return append(out, model.GateResult{
		Ordinal: len(out) + 1, Slug: model.SlugPRDPromote, Title: draftID,
		Findings: []model.Finding{
			okFinding(model.CodePRDPromoteNext, "", "",
				model.Fields{"platform": target.Platform, "prd": draftID,
					model.FieldNext: "company-os prd validate --platform " +
						target.Platform + " " + draftID}),
		},
	}), nil
}

// ErrNotReady is the readiness refusal (R-5.5).
//
// It is QUIET and it is exit 5, and both halves are the rule. Quiet, because the
// report is a stdout block naming what to supply and a second `error: …` line on
// stderr would say nothing the block has not. Exit 5, because a draft that is
// not finished is a precondition that has not been met — the same class as the
// done-gate refusal — and NOT exit 4, which means the artifact is malformed and
// is the sentence R-5.5 exists to forbid.
var ErrNotReady = model.Quiet(model.Errorf(model.ExitPrecondition,
	"draft is not ready to promote"))

// readinessIssues is steps 4 and 5: every unfilled process field (R-5.3) and
// every omitted body section (R-5.4), in one pass.
//
// The field test is `prd validate`'s, character for character — truthy and not
// `TODO` — so a draft becomes promotable at exactly the moment the record it
// would become becomes valid. Two spellings of "filled in" would produce a draft
// that promotes into a record that immediately fails `prd validate`.
//
// The section test is sectionIssues' BLOCKING half only: a missing `## ` heading
// is what R-5.4 names. An empty section under a present heading is format
// guidance the team opts into (GPF-R-4.4), and promoting is not the moment to
// start enforcing it — `prd validate` on the promoted record applies the team's
// own policy, which is R-5.13.
func readinessIssues(meta yamlio.PyMap, body []byte) []Issue {
	var out []Issue
	for _, field := range processFields {
		if truthy(meta, field) && !yamlio.PyEqual(meta.Get(field), yamlio.PyStr(placeholder)) {
			continue
		}
		out = append(out, Issue{
			Code:   model.CodePromoteFieldMissing,
			Fields: model.Fields{"field": field},
		})
	}
	blocking, _ := sectionIssues(body, PRDSections)
	for _, b := range blocking {
		out = append(out, Issue{
			Code:   model.CodePromoteSectionMissing,
			Fields: model.Fields{"section": b.Fields.Str("section")},
		})
	}
	return out
}

// promotedRecordMeta is the draft's frontmatter as a change record's (R-5.2).
//
// It is a copy and not a translation, which is the point of the draft carrying
// the same six process fields a change record does: every authored key survives
// in its authored position, and only the three the platform lifecycle owns are
// touched.
//
//   - `status` becomes `proposed` and `created` becomes today, because the
//     record's life starts now (R-5.2). `created` is written as a DATE and not a
//     string so it round-trips through the emitter the way the scaffolded
//     template's does; a quoted `'2026-01-01'` would compare differently in the
//     done-gate than a hand-scaffolded record's bare one.
//   - `promoteTo` is dropped. It is the draft's routing field — where this
//     wanted to go — and on the record it would be a claim about a decision
//     already made. `promoteTo.scope` is neither read nor carried (R-4.5).
//   - `tags` is dropped rather than copied, because tags are derived and the
//     draft's were derived for a draft. The rebuild in step 9 writes the
//     record's own.
func promotedRecordMeta(draft yamlio.PyMap) yamlio.PyMap {
	out := make(yamlio.PyMap, 0, len(draft)+2)
	for _, p := range draft {
		switch p.K {
		case KeyPromoteTo, KeyPromotedTo, KeyPromotedFrom, "tags":
			continue
		case "status":
			p.V = yamlio.PyStr(statusProposed)
		case "created":
			p.V = yamlio.PyTime(today().Format(isoDate))
		}
		out = append(out, p)
	}
	// A draft with no `status:` or no `created:` at all still produces a record
	// that has both. Set appends when the key is absent, which is the one case
	// the loop above cannot cover.
	out = out.Set("status", yamlio.PyStr(statusProposed))
	if out.Get("created") == nil {
		out = out.Set("created", yamlio.PyTime(today().Format(isoDate)))
	}
	return out
}

// writeArtifact is the read-modify-write half of RewriteFrontmatterTags, without
// the tag question: re-emit a mapping over an unchanged body.
//
// The emitter is PyDumpFrontmatter for the same reason the tag rewriter uses it
// — safe_dump's default_flow_style=None is what every committed frontmatter
// block in this workspace was written with, so anything else would rewrite files
// it did not mean to touch, and its top-level-string rule keeps a promoted
// record's authored header values on the lines the author put them on.
func writeArtifact(path string, meta yamlio.PyMap, body []byte) error {
	fm, err := yamlio.PyDumpFrontmatter(meta)
	if err != nil {
		return model.Errorf(model.ExitArtifact, "cannot serialize %s: %v", path, err)
	}
	text := "---\n" + strings.TrimSpace(fm) + "\n---\n" + string(body)
	return wrote(path, os.WriteFile(path, []byte(text), 0o666))
}

// appendPromotionLog is R-6.5: one dated, append-only line per promotion.
//
// Append-only is the requirement and not an implementation detail. The record's
// `promotedFrom` and the draft's `promotedTo` are frontmatter and can be edited;
// this line is the history that survives that. `os.O_APPEND|os.O_CREATE` is
// Python's `open(path, "a")`, which creates the file when absent — so a team
// that has never completed a change gets its log.md here.
func appendPromotionLog(path, date, draftID, platform, record string) error {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o666)
	if err != nil {
		return model.Errorf(model.ExitArtifact, "cannot open %s: %v", path, err)
	}
	line := "- " + date + ": draft `" + draftID + "` promoted to change record `" +
		record + "` on platform `" + platform + "`\n"
	if _, err := f.WriteString(line); err != nil {
		f.Close()
		return model.Errorf(model.ExitArtifact, "cannot write %s: %v", path, err)
	}
	if err := f.Close(); err != nil {
		return model.Errorf(model.ExitArtifact, "cannot write %s: %v", path, err)
	}
	return nil
}
