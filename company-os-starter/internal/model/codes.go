package model

// Every finding code and every section slug the CLI can emit, in one file.
//
// R-2.4 makes these a CONTRACT — they survive message rewordings, and R-3.4
// publishes them verbatim through `--json`. A contract that lives in six
// packages cannot be reviewed as a contract: the previous arrangement put each
// command's codes next to the command that produced them, which read well
// per-package and meant that checking for a collision, an inconsistent prefix,
// or a code nobody renders required opening every producer. internal/render
// already imported every one of those packages solely to name them.
//
// Producers import model, which imports nothing — so the codes stay reachable
// from every layer with no dependency inversion, and a reviewer reads them as
// one list.
//
// This is a relocation, not a redefinition: every string below is byte-identical
// to what it was in its producing package. Changing a VALUE is a contract change
// and belongs to task 3.4, not here.

// ---------------------------------------------------------------- validate
//
// One code per render site in cmd_validate (bin/company-os:922-1107). The
// `fail(c)` and `fail(pr)` sites at `:1078` and `:1100` pass through
// pre-composed prose from skill_conflicts and federated_slice_problems; R-2.12
// requires decomposing those, so each distinct sentence they produce gets its
// own code here rather than one code per call site.
const (
	// SlugWorkspace names the banner section that precedes gate 1. It is not a
	// gate: it carries the workspace root for `validating workspace <root>`
	// (`:924`), the one line of the report no gate can derive, and the renderer
	// identifies it by this slug so that the [N/M] denominator stays the number
	// of real gates.
	SlugWorkspace = "workspace"
	// CodeValidateRoot is that banner line.
	CodeValidateRoot = "validate.root"

	// Gate 1 — ownership reconciliation (:941, :946, :951).
	CodeOwnershipDescriptorMissing   = "ownership.descriptor-missing"
	CodeOwnershipAccountableMismatch = "ownership.accountable-mismatch"
	CodeOwnershipAgrees              = "ownership.agrees"

	// Gate 2 — deviation and exception expiry (:960, :963, :968, :971, :974).
	CodeDeviationExpired  = "expiry.deviation-expired"
	CodeDeviationCurrent  = "expiry.deviation-current"
	CodeExceptionNoExpiry = "expiry.exception-no-expiry"
	CodeExceptionExpired  = "expiry.exception-expired"
	CodeExceptionValid    = "expiry.exception-valid"

	// Gate 3 — active PRD contracts (:990, :993).
	CodePRDFrontmatterMissing = "prd.frontmatter-missing"
	CodePRDContractPresent    = "prd.contract-present"

	// Gate 4 — frontmatter core and tag derivation (:1002, :1006, :1010, :1013).
	CodeFrontmatterCoreField = "frontmatter.core-field"
	CodeTagsDrift            = "frontmatter.tags-drift"
	CodeFrontmatterInSync    = "frontmatter.in-sync"
	CodePointerGuidance      = "frontmatter.pointer-guidance"

	// Gate 5 — CLAUDE.md context node drift (:1025, :1029, :1033, :1037, :1041).
	CodeNodeIdentity  = "node.identity"
	CodeNodeAbsent    = "node.absent"
	CodeNodeHandOwned = "node.hand-owned"
	CodeNodeDrift     = "node.drift"
	CodeNodeInSync    = "node.in-sync"

	// Gate 6 — feature-index drift (:1050, :1055, :1062, :1066).
	CodeFeatureIndexAbsent     = "feature-index.absent"
	CodeFeatureIndexDrift      = "feature-index.drift"
	CodeFeatureIndexUnresolved = "feature-index.unresolved-reference"
	CodeFeatureIndexInSync     = "feature-index.in-sync"

	// Gate 7 — custom skills layering (:1078 decomposed, :1087).
	CodeSkillShadowing       = "skills.shadowing"
	CodeSkillDanglingExtends = "skills.dangling-extends"
	CodeSkillsClean          = "skills.clean"

	// Gate 8 — federated slice integrity (:1100 decomposed, :1103).
	CodeSliceLockMissing      = "federation.lock-missing"
	CodeRepoNotLocked         = "federation.repo-not-locked"
	CodeSliceSetDrift         = "federation.slice-set-drift"
	CodeSliceFileMissing      = "federation.slice-file-missing"
	CodeSliceHandEdited       = "federation.slice-hand-edited"
	CodeFederationSlicesMatch = "federation.slices-match"
)

// ------------------------------------------------------- role glossary
//
// The --role legend (`bin/company-os:1260`), shared verbatim by `today` and
// `ids list` — which is why it is neither command's.
const (
	// SlugGlossary names the legend section.
	SlugGlossary = "glossary"
	// CodeGlossaryTerm is one canonical term paired with its plain-language
	// label.
	CodeGlossaryTerm = "role.glossary-term"
)

// ------------------------------------------------------------------ today
//
// `today` prints no gate headers, so GateResult.Ordinal and GateResult.Title
// exist for --json and the TUI; Slug is what the renderers switch on. One code
// per print site in cmd_today (bin/company-os:1168-1203).
const (
	// SlugHeader is the "== today (<role>) ==" banner.
	SlugHeader = "header"
	// SlugPlatform is one platform's active-PRD and outcome-review block.
	SlugPlatform = "platform"
	// SlugTeam is one team's governance block, or its missing-governance warning.
	SlugTeam = "team"
	// SlugOnboarding is the trailing pointer to a matching onboarding guide.
	SlugOnboarding = "onboarding"

	// CodeHeader is `:1169`.
	CodeHeader = "today.header"
	// CodePlatform is `:1177`, the per-platform line carrying the active count.
	CodePlatform = "today.platform"
	// CodeActivePRD is `:1180`.
	CodeActivePRD = "today.active-prd"
	// CodeOutcomeReview is `:1187`, emitted only for a pending outcome.
	CodeOutcomeReview = "today.outcome-review"
	// CodeGovernanceMissing is the warn at `:1192`. It is the only non-ok
	// severity `today` can produce, and it does not make the command fail.
	CodeGovernanceMissing = "today.governance-missing"
	// CodeTeam is `:1194`.
	CodeTeam = "today.team"
	// CodeComponent is `:1197`.
	CodeComponent = "today.component"
	// CodeOnboarding is `:1203`.
	CodeOnboarding = "today.onboarding"
)

// -------------------------------------------------------------- ids list
//
// cmd_ids (`bin/company-os:1275-1302`). The --role legend that may precede the
// listing is SlugGlossary above, shared verbatim with `today`.
const (
	// SlugRegistry names the listing section.
	SlugRegistry = "registry"

	// CodeRegistryEmpty is the whole output when the registry is missing,
	// empty, or malformed (`:1280-1283`).
	CodeRegistryEmpty = "ids.registry-empty"
	// CodeListingHeader names the registry the listing was read from (`:1298`).
	CodeListingHeader = "ids.listing-header"
	// CodeEntry is one registered ID (`:1300`).
	CodeEntry = "ids.entry"
	// CodeCount is the trailing tally, with its "of N" suffix (`:1301-1302`).
	CodeCount = "ids.count"
)

// ----------------------------------------------------------- skills list
//
// cmd_skills (`bin/company-os:869-917`). Each Section is one contiguous block
// of the merged view, which is what gives the TUI something to scroll by and
// --json something to key on.
const (
	SectionBanner   = "banner"
	SectionLayers   = "layers"
	SectionMerged   = "merged-guidance"
	SectionPersonal = "personal-rules"
	SectionSummary  = "summary"

	// CodeBanner is the "== agent skills ... ==" title (`:871`).
	CodeBanner = "skills.banner"
	// CodeLayersHeader opens the origin-labeled section (`:874`).
	CodeLayersHeader = "skills.layers-header"
	// CodeLayerEmpty is a layer with no skills (`:878`). Every layer is listed
	// whether or not it is populated, so the reader sees which are unused.
	CodeLayerEmpty = "skills.layer-empty"
	// CodeLayerEntry is one discovered skill in the origin-labeled section
	// (`:889`).
	CodeLayerEntry = "skills.layer-entry"
	// CodeMergedHeader opens the merged guidance section (`:894`).
	CodeMergedHeader = "skills.merged-header"
	// CodeSkillHeader is one skill's label in the merged view (`:899`).
	CodeSkillHeader = "skills.skill-header"
	// CodeBaseHeader announces a resolved `extends` base (`:903`).
	CodeBaseHeader = "skills.base-header"
	// CodeBaseStep is a step inherited from that base (`:905`).
	CodeBaseStep = "skills.base-step"
	// CodeDanglingExtendsWarning is the inline warning for an `extends` that
	// does not resolve (`:907`). It is the merged view's own notice, distinct
	// from gate 7's CodeSkillDanglingExtends finding: this one never affects an
	// exit code.
	CodeDanglingExtendsWarning = "skills.dangling-extends-warning"
	// CodeStep is one of the skill's own steps (`:910`).
	CodeStep = "skills.step"
	// CodePersonalHeader opens the personal-rules section (`:913`).
	CodePersonalHeader = "skills.personal-header"
	// CodePersonalEntry is one personal rule (`:915`).
	CodePersonalEntry = "skills.personal-entry"
	// CodeSummary is the trailing tally (`:916-917`).
	CodeSummary = "skills.summary"
)

// ---------------------------------------------------------- graph build
//
// cmd_graph (`bin/company-os:1787-1797`) and rebuild_generated (`:1803-1810`)
// share one derivation and therefore one code set. The scaffolding commands
// print the middle two sections' lines before their own output, so the section
// boundaries are also the ordering contract between the two callers.
const (
	// SectionTags is the per-document "tagged …" block. Only `graph build`
	// produces it; rebuild_generated re-tags without announcing it.
	SectionTags = "tags"
	// SectionFeatureIndexes is write_feature_indexes' output.
	SectionFeatureIndexes = "feature-indexes"
	// SectionClaudeNodes is write_claude_nodes' output.
	SectionClaudeNodes = "claude-nodes"

	// CodeGraphTagged is `:1793`, one document whose derived tags changed.
	CodeGraphTagged = "graph.tagged"
	// CodeGraphIndexWritten is `:1534`, one platform's regenerated
	// feature-index.
	CodeGraphIndexWritten = "graph.feature-index-written"
	// CodeGraphNodeWritten is `:1784`, one root's regenerated CLAUDE.md node.
	CodeGraphNodeWritten = "graph.node-written"
	// CodeGraphNodeMarkersUnbalanced is the warn at `:1664`. It is the only
	// non-ok severity `graph build` can produce and it does not fail the run —
	// the fail-safe answer to a malformed node is to leave it alone.
	CodeGraphNodeMarkersUnbalanced = "graph.node-markers-unbalanced"
	// CodeGraphSummary is the trailing tally at `:1797`.
	CodeGraphSummary = "graph.summary"
)

// ---------------------------------------------------------- governance
//
// cmd_governance (`bin/company-os:331-370`), cmd_deviation (`:1112-1125`) and
// cmd_exception (`:1128-1138`). The four actions share a file — they are one
// cluster in the methodology (a rule applies, a team deviates, a team is
// excepted) — but each is its own section because the renderers key on Slug.
//
// `governance resolve` and `exception request` print no next-step line, and
// R-1.9 keeps it that way; there is deliberately no code for one here.
const (
	// SlugResolve is `governance resolve`'s output block.
	SlugResolve = "governance-resolve"
	// SlugExplain is one team's answer to `governance explain`. The command
	// emits one section per team whose generated governance names the
	// component, because `:353-363` does not stop at the first hit.
	SlugExplain = "governance-explain"
	// SlugDeviation is `deviation declare`'s output block.
	SlugDeviation = "deviation-declare"
	// SlugException is `exception request`'s output block.
	SlugException = "exception-request"

	// CodeGovernanceResolved is `:337`, the headline carrying the component
	// count.
	CodeGovernanceResolved = "governance.resolved"
	// CodeGovernanceWrote is `:338`, naming the generated file.
	CodeGovernanceWrote = "governance.wrote"
	// CodeGovernanceComponent is `:342-344`, one component's platform list and
	// requirement tally.
	CodeGovernanceComponent = "governance.component"
	// CodeGovernanceNoDescriptor is the warn at `:346`. It is the only non-ok
	// severity `governance resolve` can produce and it does not fail the run.
	CodeGovernanceNoDescriptor = "governance.no-descriptor"

	// CodeExplainComponent is `:357`, the per-team header.
	CodeExplainComponent = "governance.explain-component"
	// CodeExplainRequirement is `:361-363`, one requirement. It renders TWO
	// lines — the requirement and the sentence explaining why it applies —
	// because the pair is one record: the second line has no meaning without
	// the first and never appears alone.
	CodeExplainRequirement = "governance.explain-requirement"

	// CodeDeviationDeclared is `:1122`, naming the rule and the file.
	CodeDeviationDeclared = "deviation.declared"
	// CodeDeviationReviewDue is `:1123-1124`, the review date plus the
	// next command in the chain (R-1.8).
	CodeDeviationReviewDue = "deviation.review-due"

	// CodeExceptionDrafted is `:1137`, naming the file and the expiry.
	CodeExceptionDrafted = "exception.drafted"
	// CodeExceptionApproval is `:1138`, the standing note that an exception is
	// not valid until the rule owner approves it.
	CodeExceptionApproval = "exception.approval-note"
)

// ------------------------------------------------------------- product
//
// cmd_discover (`bin/company-os:409-464`), cmd_prd (`:551-711`) and cmd_check
// (`:714-733`). One cluster: a brief becomes a PRD becomes archived reality, and
// `check ready|done` renders the same governance checklist `prd new` injects.
//
// The core_field_errors codes are deliberately NOT product-specific. That helper
// (`:128-145`) is the process-level artifact contract and validate gate 4 calls
// it too; it lives in internal/product because that is where its first consumer
// is, and gate 4 imports it rather than growing a second copy. Its five
// sentences get five codes here so a `--json` consumer can branch on which field
// is missing instead of regex-ing one blob.
const (
	// SlugDiscoverNew is `discover new`'s output block.
	SlugDiscoverNew = "discover-new"
	// SlugDiscoverValidate is `discover validate`'s.
	SlugDiscoverValidate = "discover-validate"
	// SlugPRDNew, SlugPRDValidate and SlugPRDComplete are the three `prd`
	// actions.
	SlugPRDNew      = "prd-new"
	SlugPRDValidate = "prd-validate"
	SlugPRDComplete = "prd-complete"
	// SlugCheckBaseline is compose_checklist's team-baseline half (`:717-722`)
	// and SlugCheckGovernance its applicable-governance half (`:724-729`). They
	// are two sections rather than one because the oracle separates them with a
	// blank line and a second banner, which is exactly a section boundary.
	SlugCheckBaseline   = "check-baseline"
	SlugCheckGovernance = "check-governance"
)

// core_field_errors (`bin/company-os:128-145`), one code per sentence.
const (
	// CodeCoreTypeMissing is `:132`.
	CodeCoreTypeMissing = "core.type-missing"
	// CodeCoreIdentityMissing is `:134-135` — neither `id` nor `prd`.
	CodeCoreIdentityMissing = "core.identity-missing"
	// CodeCoreStatusMissing is `:138`, for the four LIFECYCLE_TYPES.
	CodeCoreStatusMissing = "core.status-missing"
	// CodeCoreUpdatedMissing is `:140-141`, the field `prd complete`'s
	// done-gate reads.
	CodeCoreUpdatedMissing = "core.updated-missing"
	// CodeCoreRoleMissing is `:143`.
	CodeCoreRoleMissing = "core.role-missing"
)

// `discover new|validate` (`:411-464`).
const (
	// CodeTemplateSource is the `  template: <label>` provenance line, shared
	// verbatim by `discover new` (`:419`) and `prd new` (`:613`) — which is why
	// it belongs to neither. The label itself is frozen output; see the
	// SourceBuiltin* constants in internal/scaffold.
	CodeTemplateSource = "product.template-source"
	// CodeDiscoveryCreated is `:418`.
	CodeDiscoveryCreated = "discovery.created"
	// CodeDiscoveryNext is `:420-423`.
	CodeDiscoveryNext = "discovery.next"
	// CodeDiscoveryValidated is `:459`.
	CodeDiscoveryValidated = "discovery.validated"
	// CodeDiscoveryValidateNext is `:460-462`.
	CodeDiscoveryValidateNext = "discovery.validate-next"
)

// `prd new|validate|complete` (`:551-711`).
const (
	// CodePRDCreated is `:611`.
	CodePRDCreated = "prd.created"
	// CodePRDGovernanceUnresolved is the warn at `:605`: a component the team's
	// effective governance does not name.
	CodePRDGovernanceUnresolved = "prd.governance-unresolved"
	// CodePRDRealityNote is `:617-620`, the R-1.8 pointer at `reality new` for a
	// component with no reality doc yet.
	CodePRDRealityNote = "prd.reality-note"
	// CodePRDNext is `:621-622`.
	CodePRDNext = "prd.next"
	// CodePRDContractOK is `:666`.
	CodePRDContractOK = "prd.contract-ok"
	// CodePRDValidateNext is `:667-669`.
	CodePRDValidateNext = "prd.validate-next"
	// CodePRDProcessField is `:630`, one of the six process-contract fields
	// missing or still TODO.
	CodePRDProcessField = "prd.process-field"

	// CodeSectionHeadingMissing is the ALWAYS-blocking artifact-contract failure
	// at `:441` / `:640`: a required `## ` heading is absent whatever template
	// produced the document (GPF-R-4.4).
	CodeSectionHeadingMissing = "product.section-heading-missing"
	// CodeSectionEmpty is `:446` / `:645`. It renders two different sentences —
	// the bare one when the team opted into blocking enforcement, the
	// "format guidance only" one when it did not — off the `enforced` field, so
	// a single code covers both severities rather than splitting one check in
	// two.
	CodeSectionEmpty = "product.section-empty"
)

// `prd complete`'s done-gate (`:672-711`) — invariant #4.
const (
	// CodeDoneCheckHeader is `:700`, the banner that precedes the refusals.
	CodeDoneCheckHeader = "done.header"
	// CodeDoneChecklistUnchecked is `:681-682`.
	CodeDoneChecklistUnchecked = "done.checklist-unchecked"
	// CodeDoneRealityMissing is `:687-688`.
	CodeDoneRealityMissing = "done.reality-missing"
	// CodeDoneRealityStale is `:692-693`.
	CodeDoneRealityStale = "done.reality-stale"
	// CodeDoneRealityDateInvalid has no counterpart in the oracle. It is the
	// R-1.14 / OKF Phase 0 fix: the done-gate compared the two dates as raw
	// STRINGS, which is right for well-formed ISO dates by lexical accident and
	// silently wrong for `18/07/2026`, an empty value, or a YAML-parsed
	// datetime. A value that will not parse now names itself instead of being
	// silently ordered.
	CodeDoneRealityDateInvalid = "done.reality-date-invalid"
	// CodeDoneFix is `:705-706`, the R-1.8 pointer for each missing reality doc.
	CodeDoneFix = "done.fix"

	// CodePRDArchived is `:708`.
	CodePRDArchived = "prd.archived"
	// CodeOutcomeScheduled is `:709`.
	CodeOutcomeScheduled = "prd.outcome-scheduled"
	// CodeLogAppended is `:710`.
	CodeLogAppended = "prd.log-appended"
	// CodePRDCompleteNext is `:712`. It is printed AFTER rebuild_generated's
	// lines, not before them — the one scaffolding-adjacent command whose own
	// output brackets the derived output rather than following it.
	CodePRDCompleteNext = "prd.complete-next"
)

// `check ready|done` (`:714-733`).
const (
	// CodeCheckBaselineHeader is `:719`, naming the definition-of-<kind> file.
	CodeCheckBaselineHeader = "check.baseline-header"
	// CodeCheckBaselineText is `:721`, the whole team baseline document.
	CodeCheckBaselineText = "check.baseline-text"
	// CodeCheckBaselineMissing is the warn at `:723`.
	CodeCheckBaselineMissing = "check.baseline-missing"
	// CodeCheckGovernanceHeader is `:725`.
	CodeCheckGovernanceHeader = "check.governance-header"
	// CodeCheckChecklist is `:728` — the rendered governance checklist, or the
	// "(none resolved)" placeholder when nothing resolved.
	CodeCheckChecklist = "check.checklist"
	// CodeCheckUnresolved is the warn at `:731`.
	CodeCheckUnresolved = "check.component-unresolved"
)

// ------------------------------------------------- scaffolding + federation
//
// cmd_init (`bin/company-os:1968-1995`), cmd_add (`:1997-2027`), cmd_reality
// (`:2030-2058`), cmd_scratchpad (`:1141-1155`) and cmd_workspace
// (`:2542-2553`).
//
// These five printed prose straight to stdout until R-3.4b: one JSON encoder
// over the record types can only reach a command that produces record types, and
// R-3.7 needs their envelope to describe what they created rather than default to
// an empty document. So they return findings like every other command, and the
// text side renders each finding's Message verbatim — which is what those lines
// already were.
//
// Their Message is therefore the WHOLE line including any leading `next: ` or
// two-space indent, unlike the product cluster where the renderer supplies the
// prefix. That is deliberate: these six line shapes have no grammar worth
// factoring, and freezing the bytes at the producer is what makes the conversion
// provably output-neutral (R-0.8, R-3.3).
const (
	// SlugGenerated carries rebuild_generated's derived lines, which the oracle
	// prints BEFORE each scaffolding command's own output.
	SlugGenerated = "generated"
	// SlugInit, SlugAdd, SlugRealityNew and SlugScratchpad are the four
	// scaffolding commands' own blocks.
	SlugInit       = "init"
	SlugAdd        = "add"
	SlugRealityNew = "reality-new"
	SlugScratchpad = "scratchpad-init"
	// SlugSync and SlugStatus are `workspace sync` and `workspace status`.
	SlugSync   = "workspace-sync"
	SlugStatus = "workspace-status"

	// CodeGenerated is one line of rebuild_generated's output, passed through
	// verbatim. It is already render.Graph's rendering of a graph record set;
	// scaffold.Rebuild hands it back as text, so this code carries text.
	CodeGenerated = "scaffold.generated"

	// CodeInitCreated is `:1993`, CodeInitSummary `:1994`, CodeInitNext `:1995`.
	CodeInitCreated = "init.created"
	CodeInitSummary = "init.summary"
	CodeInitNext    = "init.next"

	// CodeAddCreated is `:2011`/`:2018`/`:2026` and CodeAddNext the next-step
	// line that follows each. One code covers all three units: the unit kind is
	// in Fields, not in the code.
	CodeAddCreated = "add.created"
	CodeAddNext    = "add.next"

	// `add team <id> --repair`. Three codes, not one, because the useful
	// question is which files it touched: CodeAddRepairWrote per file created,
	// CodeAddRepairSkipped per file left alone (the evidence that nothing was
	// overwritten), and CodeAddRepairNoop when everything was already present —
	// a repair that printed nothing would read as a failure.
	CodeAddRepairWrote   = "add.repair-wrote"
	CodeAddRepairSkipped = "add.repair-skipped"
	CodeAddRepairNoop    = "add.repair-noop"

	// CodeRealityCreated is `:2055`, CodeRealityTemplate `:2056` and
	// CodeRealityNext `:2057`.
	CodeRealityCreated  = "reality.created"
	CodeRealityTemplate = "reality.template-source"
	CodeRealityNext     = "reality.next"

	// CodeScratchpadCreated is `:1155`. There is deliberately no next-step code:
	// `scratchpad init` prints none today and R-1.9 outranks R-1.8 here.
	CodeScratchpadCreated = "scratchpad.created"

	// CodeSyncHeader is `:2566`, CodeSyncRepo one repo's line (`:2589`/`:2599`),
	// CodeSyncLock the `wrote`/`materialized` trailer (`:2624`/`:2627`) and
	// CodeSyncNext `:2632`.
	CodeSyncHeader = "sync.header"
	CodeSyncRepo   = "sync.repo"
	CodeSyncLock   = "sync.lock"
	CodeSyncNext   = "sync.next"

	// CodeStatusHeader is `:2637`, CodeStatusRepo one repo's line and
	// CodeStatusNext the trailing pointer, which names `sync` or `validate`
	// depending on whether anything needs action.
	CodeStatusHeader = "status.header"
	CodeStatusRepo   = "status.repo"
	CodeStatusNext   = "status.next"
)

// FieldNext is the Fields key carrying the next command in the workflow, bare —
// no `next: ` prefix and none of the prose the sentence wraps it in.
//
// R-1.8 makes that command the system's principal affordance and R-3.2 forbids
// the prose carrying it from reaching `--json` stdout, so without a structured
// home it would simply vanish for the one consumer `--json` exists for (R-3.6).
// A key rather than a code set is what keeps the JSON encoder uniform: it lifts
// every finding that has this key and needs to know nothing about which codes
// those are. Producers that compose a next-step sentence set it; everyone else
// leaves it absent.
const FieldNext = "next"

// NextCommands collects, in order, the next-command guidance carried by a record
// set. It is the whole of R-3.6 on the reading side.
func NextCommands(sections []GateResult) []string {
	var out []string
	for _, s := range sections {
		for _, f := range s.Findings {
			if next := f.Fields.Str(FieldNext); next != "" {
				out = append(out, next)
			}
		}
	}
	return out
}

// The governance checklist `prd new` injects and `check` prints
// (gather_prd_governance, `:551-570`). Its records are NOT rendered as findings:
// they compose a markdown fragment that is written into an artifact, so they
// carry their own codes for --json rather than for a line grammar.
const (
	// CodeChecklistComponent is `:559`, the bold component header.
	CodeChecklistComponent = "checklist.component"
	// CodeChecklistItem is `:561` and `:565` — one unchecked requirement.
	// Scope is "company" or a platform id, which is the only difference between
	// the two Python lines.
	CodeChecklistItem = "checklist.item"
)
