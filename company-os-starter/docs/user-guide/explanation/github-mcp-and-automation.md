---
title: GitHub MCP and automation
---

> **Company OS ships no MCP server and no MCP client.** Nothing on this page
> asks you to install one, and no `company-os` command talks to the GitHub API.
> If you came here to pull a repo's docs into your workspace, the shipped answer
> is [how-to/sync-a-knowledge-catalog.md](../how-to/sync-a-knowledge-catalog.md).

# GitHub MCP and automation

The recurring proposal is reasonable enough that it's worth answering
precisely: *let an agent drive the knowledge catalog through the GitHub MCP
server — pull a specific directory out of a component repo, push changes back,
notify consumers when docs move, and have GitHub Actions keep it all in sync.*

Most of what that proposal wants already exists, implemented as plain git rather
than as an API integration. The rest is deliberately absent. This page separates
the two so you can tell what to rely on from what to build yourself.

## What already does the job

The three-part mapping — *source repo → a directory inside it → a local
destination* — is three keys on a `workspace.yaml` entry:

| What you want | The key |
|---|---|
| the repo holding the documentation | `url:` |
| the directories inside it, and nothing else | `paths:` (an allowlist) |
| where it lands in your workspace | `localDirectory:` |

One repo contributing several areas uses a `slices:` list instead — one clone,
one cache, N disjoint destinations. `company-os workspace sync` then
materializes them, and `workspace.lock.yaml` records the resolved commit plus a
per-file hash of everything written.

The whole of `sync` is eight git invocations, and the list is the point:

```text
git --version                                  # floor is 2.27
git init --quiet / remote add|set-url origin
git fetch --filter=blob:none --depth 1 <ref>   # blobless, shallow
git rev-parse FETCH_HEAD^{commit}              # pin resolved to a SHA
git sparse-checkout init --cone / set <dirs>
git checkout --quiet --detach <sha>
```

No `push`, no `commit`, no `pull`, no `merge`, and no HTTP API call anywhere.
Everything materialized is then frozen — files `0444`, directories `0555` — and
validate gate `[8/8]` re-hashes the tree offline on every run.

That design is what a GitHub MCP integration would have to replace, and it buys
three properties an API-driven sync would have to re-earn:

- **Reproducible.** A pin is a commit or a tag; a branch is rejected as
  floating. The same manifest produces the same bytes next year.
- **Offline and CI-cheap.** Gate `[8/8]` needs no network, no credentials, and
  no git cache — the slices and the lock are committed, so a clean runner just
  checks hashes.
- **Auditable in one diff.** The manifest, the slices, and the lock move
  together in one commit. "Which content changed, from which upstream SHA" is a
  `git diff`, not a query against a service.

## What is deliberately absent

| The proposal wants | Ships? | Reality |
|---|---|---|
| repo → source subdirectory mapping | **Yes** | `paths:` allowlist, enforced twice (sparse-checkout, then the copy) |
| repo → local destination mapping | **Yes** | `localDirectory:`; `knowledge/` is a canonical root, so a catalog target is legal |
| several destinations from one repo | **Yes** | `slices:` — one clone, one cache; targets must be disjoint |
| pulling docs without the source code | **Yes** | blobless + shallow + cone sparse-checkout + copy-time allowlist |
| glob or exclude patterns | No | allowlist only; entries are literal path prefixes |
| renaming a path on the way in | No | the path below `localDirectory:` mirrors the source exactly |
| tracking a branch | No | `pin:` takes `commit:` or `tag:`; a branch is refused as non-reproducible |
| pushing local edits back upstream | No | zero write-side git verbs; slices are `0444` and a re-sync overwrites them |
| detecting an upstream change | No | a pin bump is a hand edit; there is no polling, webhook, or watcher |
| notifying consumers when docs move | No | no bot, cron, daemon, or notification path exists |
| conflict detection or merge guidance | No | only hash mismatch at gate `[8/8]`; the remedy is always re-sync |
| a GitHub MCP integration | No | no MCP server, no MCP client, no `.mcp.json` anywhere in the repo |
| a committed Actions workflow that syncs | No | the only committed workflow publishes release artifacts; every sync/validate workflow in the docs is an example to copy |

Three recorded decisions stand behind the "No" column, and they were not
oversights:

- **Local Search shipped without an MCP server on purpose** — a CLI plus a
  Claude skill instead ([tutorials/03](../tutorials/03-search-your-workspace.md)).
- **Live external-system integration is a stated non-goal** of the federation
  work — "MCP pulls, Slack posting, Figma reads" are named together
  (`docs/lld/federation-enrichment.md`).
- **Repository directory sync via GitHub MCP / Actions is deferred**, tracked as
  non-goal N8 (`docs/hld/okf-v02-conformance.md`). An optional MCP shim appears
  in the [Observer roadmap](observer-roadmap.md) phase 4, not started.

The through-line: the CLI stays offline-fast, and only `workspace sync` touches
the network at all.

One more thing worth not misreading — the `CLAUDE.md` files in a workspace are
**generated context nodes**, produced by `graph build` and drift-checked by gate
`[5/N]`. They are indexes of what a root contains, not agent configuration and
not an MCP manifest.

## Using GitHub MCP anyway

Nothing stops you, and for some work it's the right tool. Just be clear about
whose tool it is: the GitHub MCP server is configured in *your agent*, and
Company OS neither knows nor cares that it exists. That makes the division of
labor simple.

Reasonable, with no Company OS invariant at risk:

- Reading a source repo's tree to decide what belongs in `paths:` before you
  write the manifest entry.
- Finding the tag or commit you want to pin to, and reading the upstream diff
  between two pins to see what a bump would bring in.
- Opening the pull request that bumps `pin:` on the workspace repo, or the PR
  that fixes the docs *in the source repo* — which is the only correct place to
  fix a slice's content.
- Ordinary repo chores around a change record: branch, PR, review comments.

Not reasonable, because the CLI is the only writer:

- **Writing into a materialized slice.** It is `0444` derived content; gate
  `[8/8]` fails on any hand-edit and names the file. Change the source, bump the
  pin, re-sync.
- **Writing anything under `generated/`.** `governance resolve` owns
  `effective-governance.yaml`; CI regenerates and diffs it.
- **Fetching content into the workspace itself.** `workspace sync` writes the
  lock; content that arrives another way has no recorded provenance and will
  either fail the hash check or, worse, pass one that means nothing.

The rule that keeps this coherent: **an agent may read anything, but the CLI
writes the workspace and the lock is the oracle.** An MCP server that respects
that boundary is a convenience. One that crosses it produces a workspace whose
validation results are no longer evidence of anything.

## If you want the automation anyway

The pieces are all documented, just not assembled for you:

- `company-os validate` exits non-zero on any problem, which is the whole CI
  contract — see [how-to/run-the-validation-gate.md](../how-to/run-the-validation-gate.md).
- `company-os workspace status` reports pin drift, slice-set drift, and
  hand-edits, so a scheduled job that runs it and opens an issue is a small
  script, not an integration.
- `FEDERATION-RUNBOOK.md` §5 carries two full workflow examples, including the
  source-repo variant that overlays a PR's content into a composition checkout
  before validating.
- Every command speaks `--json` with stable codes and a `guidance` chain, so an
  agent or script branches on structure rather than parsing prose:
  [reference/company-os-cli.md § `--json`](../reference/company-os-cli.md#--json).

Build it in your own repo and the invariants hold. What is not on offer is a
Company OS feature that reaches out to GitHub on its own.

## Related

- [how-to/sync-a-knowledge-catalog.md](../how-to/sync-a-knowledge-catalog.md) — the shipped way to pull a repo's docs in
- [how-to/use-the-agent-skills.md](../how-to/use-the-agent-skills.md) — what a skill can and cannot ask an agent to do
- [explanation/how-it-fits-together.md](how-it-fits-together.md) — the tools that do exist, and how they meet
- [explanation/observer-roadmap.md](observer-roadmap.md) — where the optional MCP shim sits on the roadmap
- [FEDERATION-RUNBOOK.md](../../FEDERATION-RUNBOOK.md) — the full federation model and CI patterns
