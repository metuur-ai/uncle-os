# failing-workspace — the failure-path fixture for `company-os validate`

This workspace is **deliberately broken** and is **expected to exit 1**. It is
not a template — do not copy it. Its only job is to be an oracle: it drives at
least one `[FAIL]` through every failure site in gates `[1/7]`–`[7/7]` of
`cmd_validate`, plus the single `warn()` site, so the failure-rendering path of
the CLI has a byte-frozen snapshot (`examples/failing-golden-validate.txt`).

The two green goldens (`golden-validate.txt`, `federated-golden-validate.txt`)
contain zero `[FAIL]` and zero `[warn]` lines between them. Every prefix shape
below — the component id, the team id, the quoted component id, the compound
`platform/prd-id`, the workspace-relative path, the platform id, and the two
gates that emit no prefix at all — is exercised only here.

Gate `[8/8]` is not reachable from a monorepo workspace (it self-suppresses when
there is no `workspace.yaml`); it is covered by `examples/failing-federated/` and
`examples/failing-federated-nolock/`.

## What is broken, and why

| Gate | Finding | Planted in |
|---|---|---|
| 1 | owns a component with no descriptor | `teams/ghost/ownership/components.yaml` — `no-such-service` |
| 1 | claims `accountable` against a descriptor naming another team | same file — `svc-alpha` vs `platforms/alpha/components/svc-alpha.yaml` (`team://other`) |
| 2 | deviation past its `reviewDate` | `teams/ghost/governance/deviations.yaml` — `2020-01-01` |
| 2 | exception with no `expires` at all | `teams/ghost/governance/exceptions.yaml` — entry 1 |
| 2 | exception past its `expires` | same file — entry 2, `2020-01-01` |
| 3 | active PRD missing contract frontmatter | `platforms/alpha/change-records/active/2035-broken-contract/prd.md` |
| 4 | frontmatter core field missing | `platforms/alpha/reality/components/svc-alpha.md` — no `updated:` |
| 4 | committed tags drifted from derivation | `teams/ghost/product/discovery/2035-drifted-tags/brief.md` |
| 4 | `warn` — malformed `pointers:` | `teams/ghost/product/discovery/2035-bad-pointers/brief.md` |
| 5 | team identity shape error | `teams/ghost/team.yaml` — `roster[1]` has no `role` |
| 5 | generated CLAUDE.md block drifted | `platforms/alpha/CLAUDE.md` |
| 6 | feature-index drifted from derivation | `platforms/alpha/generated/feature-index.yaml` |
| 6 | index references an unresolvable discovery | `platforms/beta/...` — PRD `fromDiscovery: 2035-no-such-brief` |
| 6 | index references an unresolvable PRD | `platforms/beta/archive/prds/2035-old-dirname/` — see below |
| 7 | skill shadows a canonical id | `teams/ghost/skills/creating-prd.SKILL.md` |
| 7 | skill `extends:` a base that does not exist | `teams/ghost/skills/reviewing-prd.SKILL.md` |

## Three details that are load-bearing, not accidental

**The reality doc omits `updated:` specifically.** Gate 4's `[ok]` line is
conditional on the absence of core-field errors, so a document with core errors
emits its `[FAIL]` lines and **no** `[ok]` line. `updated` feeds neither
`derive_tags` nor the CLAUDE.md node, so the tag comparison still passes and
that conditional is isolated in the snapshot with nothing else moving.

**Two platforms, not one.** Gate 6 `continue`s after reporting index drift, so
the drift finding and the unresolved-reference findings can never come from the
same platform. `alpha` carries the drift; `beta` keeps an in-sync index whose
references dangle.

**`platforms/beta/archive/prds/2035-old-dirname/` is named that way on purpose.**
Its `prd.md` declares `id: 2035-renamed-change`, and `outcome.md` keys its
`prd:` edge to that id. The index therefore records an outcome for
`2035-renamed-change`, which resolves to no `archive/prds/2035-renamed-change/`
directory — that is what makes the `kind='prd'` half of the unresolved-reference
message reachable at all, alongside the `kind='discovery'` half.

## Findings that pass here on purpose

Three gate-5 `[ok]` shapes are all rendered: `company-os/CLAUDE.md` is
hand-owned (its generated markers were stripped), `teams/ghost/` has no
`CLAUDE.md` at all (absent → pass), and `platforms/beta/CLAUDE.md` is in sync.
Gates 1, 2, 3 and 4 each also emit at least one `[ok]`, so the snapshot freezes
the interleaving of passing and failing lines within a gate, not just the
failures.

## Maintaining it

`graph build` will "repair" most of this — it rewrites tags, CLAUDE.md nodes and
feature indexes. Never run it against this fixture. If a finding needs to
change, edit the file named in the table above and re-baseline with
`examples/acceptance.sh --update`.
