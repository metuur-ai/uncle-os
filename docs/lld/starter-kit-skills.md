# Starter-Kit Skills Repair — Low-Level Design

## Architecture

Four file moves and eight frontmatter lines. No Go changes except one new test.

### The moves

```
skills/completing-a-change/SKILL.md      -> skills/completing-a-change.SKILL.md
skills/creating-prd/SKILL.md             -> skills/creating-prd.SKILL.md
skills/requesting-an-exception/SKILL.md  -> skills/requesting-an-exception.SKILL.md
skills/running-discovery/SKILL.md        -> skills/running-discovery.SKILL.md
```

Each emptied directory is removed. `Suffix` is the constant `.SKILL.md`
(`internal/skills/skills.go:22`), and `globSorted` matches `*.SKILL.md` against
entries of a single directory (`skills.go:249-269`), so only the flat form is
reachable.

### The frontmatter

Two edits per file. Add `type: skill`, which `CoreFieldErrors`
(`internal/product/contract.go:65-68`) requires of every artifact. Then replace
the hand-written tag list with what derivation produces.

All four currently read `tags: [authority/canonical, kind/skill, process/<x>]`.
`DeriveTags` (`internal/graph/tags.go:69-117`) reads `type`, `platform`, `team`,
`components`, `boundedContext`, `status`, `authority`, `fromDiscovery`, `prd`,
and `role`. Of the four files' frontmatter, only `authority: canonical` is an
input. `kindTag` (`tags.go:22-29`) has no `"skill"` entry, so `type: skill`
contributes no kind tag; `process/*` is not in `curatedFacets` (`tags.go:33-35`),
so it is dropped. Derivation therefore produces `[authority/canonical]` — which
is exactly what the one committed company-layer skill in the corpus carries,
`examples/workspace/company-os/skills/syncing-knowledge.SKILL.md:9`.

### The test

For each of the four: copy into a synthesized workspace's `company-os/skills/`,
run discovery and the frontmatter gate, assert the file is found and no
core-field or tag-drift finding is raised. This is the only new code, and it is
the only thing that will notice if these files rot again.

## Constraints

**Derived tags are layer-specific, so the guarantee is scoped.** `IterGraphDocs`
adds a location facet from the graph root (`internal/graph/tags.go:203-209,
232-235`). `graphRoots` (`tags.go:268-272`) gives `company-os` no facet, but a
platform root yields `platform/<p>` — visible in the corpus at
`examples/workspace/platforms/communications/skills/creating-prd.SKILL.md:9`,
which reads `tags: [authority/canonical, platform/communications]`.

So `[authority/canonical]` is in sync at the **company layer only**. Copy a
starter-kit skill to `platforms/<p>/skills/` and the tags gate raises drift until
`graph build` runs. That is correct behavior, not a bug — tags are derived and
never hand-written, per standing invariant — but the promise this change makes
has to be stated at the layer it holds, and the files should say so.

**These files are read by nothing, so nothing will catch a regression.** No
golden covers them, no existing test names them, and the CLI never opens them.
The new test is the entire safety net; without it this change is unverifiable and
the same drift recurs.

**Skills are deliberately kind-tagless, and that is worth noticing.**
`kindTag` maps `prd`, `discovery-brief`, `component-reality`, `outcome-review`
and others to `kind/*`, but has no `skill` entry, so a skill artifact gets no
kind facet at all. The hand-written `kind/skill` in these four files was
someone's reasonable guess at the ontology's faceted scheme. Adding
`"skill": "kind/skill"` to the map would make the guess correct — and would also
change the derived tags of the two committed workspace skills, their generated
context nodes, and two goldens. That is a different change with a different blast
radius, and it is out of scope here.

## Key Decisions

**Flatten rather than teach the nested form.** The alternative is to widen the
discovery glob to match `<name>/SKILL.md` as well. Rejected: the glob is a
transcription of the reference implementation's own behavior, the flat form is
what `templates/SKILL-template.md` already documents as required, and every
committed workspace skill already uses it. Four files are wrong, not the rule.

**Match derivation exactly rather than adding the missing `kindTag` entry.**
Adding `kind/skill` to the map is one line and arguably more correct against the
ontology guide. It also regenerates tags on two committed skills, four context
nodes, and the goldens that render them — turning a four-file repair into a
corpus-wide one. Keep this change unverifiable-by-nothing small; raise the
ontology question separately.

**No enforcement.** The obvious follow-on is a gate that would have caught this.
That gate is specified elsewhere and carries a gate insertion, an ordinal shift,
and a counter sweep across the documentation. None of that is needed to make four
broken files work, and coupling them would delay a repair that is ready now.

**Ship this before anything else touching these files.** Two other changes are
specified against the same four files. This one has no dependencies, no golden
churn, and no migration, so it goes first and the others rebase onto it.

## Out of Scope

- Any prefix, rename, or namespace decision.
- Any new validation gate or enforcement mechanism.
- Adding `skill` to the kind-tag map, and the corpus-wide tag regeneration that
  would follow.
- The reference copies under `templates/`, which share the no-reader status.
- Packaging any of these skills for an agent runtime.
