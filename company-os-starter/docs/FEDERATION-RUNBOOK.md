---
type: doc
tags: [doc/company-os-starter, kind/runbook]
---

# Federation Runbook: Running Company OS Across Many Repos

This is the day-2 operations manual for **federated mode** — one governance view
composed from many Git repositories, without submodules and without copying files
by hand. It satisfies `GPF-R-7.8` and covers the four flows that federation adds
to the golden path: **pin bump → sync → resolve → commit-back**, the **two-PR
ownership transfer**, and a **reusable per-repo CI recipe**.

Every command block below was executed against a throwaway fixture (two source
repos + a thin workspace repo, pinned with `file://` URLs) and the output is
quoted verbatim. Reproduce it with the same technique the tests use: `git init`
a couple of repos, point `workspace.yaml` at them, `workspace sync`.

**The principle is unchanged:** strict on artifacts, flexible on process. Federated
slices are *derived, read-only content* — governed exactly like `generated/`. The
directory layout is the contract; the repo boundary is not. See
[TUTORIAL.md](TUTORIAL.md) for the monorepo walkthrough this builds on.

---

## 0. When federation is on, and what it materializes

Presence of a `workspace.yaml` at the workspace root is the **only** switch into
federated mode (`GPF-R-6.1/6.2`). No `workspace.yaml` ⇒ monorepo mode, byte-for-byte
unchanged — every command behaves as in the tutorial, git is never required, and
`validate` runs its 7 gates. Add the manifest and a single eighth gate appears.

The mental model has four moving parts:

| Part | Who owns it | Committed? | Editable? |
|---|---|---|---|
| `workspace.yaml` (manifest) | you, by hand | yes | yes — it *is* the composition |
| Materialized slices (`company-os/`, `platforms/<p>/…`) | the CLI | yes | **no** — read-only, `0444`, like `generated/` |
| `workspace.lock.yaml` (resolved SHAs + per-file hashes) | the CLI | yes | no — machine-owned |
| `.company-os/` (blobless git caches) | the CLI | **no** — git-ignored | n/a |

A federated workspace repo is a **composition repo**: it versions *which* repos at
*which* pins, plus the read-only slices those pins produced and the lock that proves
them. The authoritative, editable artifacts live in their **source repos** (the
company-os repo, each platform repo, each team repo). You never edit a slice; you
change the source, cut a new pin, and re-sync.

### The manifest shape (as the CLI actually parses it)

```yaml
# workspace.yaml — a flat list of repos. Presence of this file = federated mode.
version: 1
repos:
  - name: company-os                     # cache key + status/lock label (unique)
    url: https://github.com/acme/company-os.git
    root: company-os                     # where the slice lands in the canonical layout
    pin: {tag: v1.0.0}                   # EXACTLY ONE of commit:/tag: — no branches
    paths: [standards/, onboarding/]     # governance allowlist — DIRECTORIES ONLY
  - name: platform-communications
    url: https://github.com/acme/platform-communications.git
    root: platforms/communications
    pin: {commit: c7241463de650417298d402409ad56d00fcbebd5}
    paths: [governance/, components/, reality/, change-records/, archive/, generated/, skills/]
```

Hard rules the CLI enforces at load (they fail `sync`, `status`, and `validate`
identically):

- **`name`, `url`, `root` are required.** `root` must be a relative path landing
  under a canonical root (`company-os/`, `platforms/`, `teams/`, `company-ontology/`).
- **`pin:` accepts exactly one of `commit:` or `tag:`.** A branch name, a bare
  `ref:`, or *both* `commit:` and `tag:` together are rejected as floating or
  ambiguous (`GPF-R-6.3`). A `tag:` is resolved to a SHA at sync and recorded in
  the lock (`GPF-R-6.4`).
- **`paths:` is a directory allowlist.** List directories, not files (see
  [§7 gotchas](#7-git-version-offline-and-gotchas)). Omit it to accept the default
  governance allowlist (`governance/`, `components/`, `governance/requirements.yaml`,
  `reality/`, `skills/`, `templates/`).

> These field names differ from the earlier design sketch (which nested
> `company-os:`/`platforms:`/`teams:` and allowed `branch:`). The flat `repos:`
> list above is what `bin/company-os` parses today — treat the CLI as ground truth.

---

## 1. Bootstrap a federated workspace

Author the manifest, run the first sync, commit the slices and the lock.

```bash
# 1. Create the thin workspace repo: workspace.yaml + your native team dir(s).
#    (Native = artifacts THIS repo owns and edits; slices come from sync.)
cat > workspace.yaml <<'YAML'
version: 1
repos:
  - name: company-os
    url: https://github.com/acme/company-os.git
    root: company-os
    pin: {tag: v1.0.0}
    paths: [standards/, onboarding/]
  - name: platform-communications
    url: https://github.com/acme/platform-communications.git
    root: platforms/communications
    pin: {tag: v2.1.0}
    paths: [governance/, components/, reality/, change-records/, archive/, generated/, skills/]
YAML
```

First sync materializes the read-only slices and writes the lock:

```console
$ company-os workspace sync
workspace sync (2 repo(s))

  synced company-os @ 71ab0661805d (tag v1.0.0) -> company-os (3 file(s))
  synced platform-communications @ 6622076972b9 (tag v2.1.0) -> platforms/communications (10 file(s))

wrote workspace.lock.yaml (2 repo(s))
next: company-os workspace status   # then: company-os validate
```

Confirm the state, then validate the composed tree:

```console
$ company-os workspace status
workspace federation status (2 repo(s))

  company-os: tag:v1.0.0 @ 71ab0661805d -> company-os — clean
  platform-communications: tag:v2.1.0 @ 6622076972b9 -> platforms/communications — clean

next: company-os validate
```

`sync` also wrote a `.gitignore` entry for the machine-owned cache. Commit the
manifest, the slices, and the lock — **but not** `.company-os/`:

```console
$ git status --porcelain
?? .gitignore            # contains ".company-os/"
?? company-os/           # slice — committed
?? platforms/            # slice — committed
?? workspace.lock.yaml   # lock  — committed
$ git check-ignore .company-os
.company-os              # cache — NOT committed
```

The lock records, per repo, the original pin, the resolved commit, an aggregate
`sliceHash`, and a `{path: sha256}` map. That map is the hand-edit oracle for
gate `[8/8]`:

```yaml
# workspace.lock.yaml (excerpt)
repos:
- name: platform-communications
  url: https://github.com/acme/platform-communications.git
  root: platforms/communications
  pin:
    tag: v2.1.0
  resolvedCommit: 6622076972b9f82b10e03ba584aa9199f77e9eb7
  sliceHash: a2a6445fe15e0b81ed0b2554466b9299a5ed895e236df0cb86cad0b660751b08
  files:
    platforms/communications/governance/requirements.yaml: 7481ac4c3b661dec…
    platforms/communications/components/customer-notification-service.yaml: 9047752c…
    …
```

A federated `validate` adds gate `[8/8]`; gates `[1/8]`–`[7/8]` are the monorepo
gates, unrenumbered:

```console
$ company-os validate
validating workspace /…/workspace

[1/8] ownership reconciliation
  [ok] customer-notification-service: registry and descriptor agree (communications)
…
[7/8] custom skills layering (shadowing + extends resolution)
  [ok] skills layered cleanly (1 canonical, 0 team, 1 personal; no shadowing or dangling extends)

[8/8] federated slice integrity (read-only derived content)
  [ok] federated slices match workspace.lock.yaml (13 file(s) across 2 repo(s); no hand-edits)

PASS
```

> **Derived artifacts must be built upstream.** The slice carries whatever the
> source repo committed at the pinned SHA — including `graph build` outputs
> (`generated/feature-index.yaml`, and the generated block in each `CLAUDE.md`).
> You cannot regenerate them in the workspace: the slice is read-only, and a
> source repo on its own is not a workspace root. If those artifacts are stale at
> the pin, gates `[5/8]`/`[6/8]` fail and the *only* fix is upstream — run
> `graph build` in a full workspace that contains the source repo, commit it back
> to the source repo, cut a new pin, re-sync. Keep `graph build && git diff
> --exit-code` in each source repo's own CI so a pin can never carry drift.

---

## 2. Pin bump → sync → resolve → commit-back

This is the core day-2 loop: a platform (or the company baseline) ships new
governance, and consuming teams adopt it deliberately.

**Step 1 — upstream cuts a release.** In the platform's *source* repo, the change
lands and is tagged (governance is pinned by `tag`+SHA, not tracked live):

```bash
# in the platform-communications source repo
git commit -am "quiet-hours delivery clause"
git tag v2.2.0 && git push --tags
```

**Step 2 — bump the pin.** In the workspace repo, edit one line of `workspace.yaml`
(`tag: v2.1.0` → `tag: v2.2.0`). `status` now reports drift *before* you sync —
the manifest pin no longer matches the lock:

```console
$ company-os workspace status
workspace federation status (2 repo(s))

  company-os: tag:v1.0.0 @ 71ab0661805d -> company-os — clean
  platform-communications: tag:v2.2.0 @ 6622076972b9 -> platforms/communications — drifted (manifest pin tag:v2.2.0 != lock {'tag': 'v2.1.0'})

next: company-os workspace sync
```

**Step 3 — sync.** Re-materialize the changed slice; the lock's `resolvedCommit`
advances. `--only <name>` limits work to one repo (a single name, not a list):

```console
$ company-os workspace sync --only platform-communications
  synced platform-communications @ 5d257bf542fa (tag v2.2.0) -> platforms/communications (7 file(s))
wrote workspace.lock.yaml (2 repo(s))

$ company-os workspace status
  platform-communications: tag:v2.2.0 @ 5d257bf542fa -> platforms/communications — clean
```

**Step 4 — resolve, per affected team.** New platform requirements do not reach a
team's effective governance until you regenerate it. `governance resolve` reads
the read-only company-os and platform slices and writes the team's **own** file —
`teams/<t>/generated/effective-governance.yaml` is native, editable team content,
not a slice:

```console
$ company-os governance resolve --team customer-engagement
resolved governance for team 'customer-engagement' (1 component(s))
wrote teams/customer-engagement/generated/effective-governance.yaml
  customer-notification-service: platforms [communications], 3 company + 3 platform requirement(s)
```

**Step 5 — commit back.** Commit the regenerated `generated/` to whichever repo
owns `teams/<t>/`. In a team-centric workspace the team dir is native, so this is
one PR against the workspace/team repo carrying: the bumped `workspace.yaml`, the
re-synced slice, the new `workspace.lock.yaml`, and the regenerated
`generated/effective-governance.yaml`. The lock diff *is* the audit trail of which
governance SHAs moved.

### Why `resolve` is a separate step from `sync`

`sync` is a pure **materializer** — network (or the cache) and the lock, nothing
more. It deliberately does *not* run `governance resolve` (this was the LLD's open
question; the CLI resolves it as "separate", and `sync`'s printed next-command
points at `status`/`validate`, never `resolve`). Three reasons the two stay
decoupled:

1. **Different ownership.** `sync` writes machine-owned slices + lock; `resolve`
   writes *team-owned* `generated/` that a human reviews and commits. Fusing them
   would smuggle a governance decision into a materialization step.
2. **`sync` doesn't know which teams are "yours."** A workspace may federate team
   repos it only *reads* (an enterprise-wide validate view). `sync` can't assume
   every team should be resolved and committed.
3. **Single-purpose commands, explicit chain.** Keeping them apart preserves the
   "every mutating command prints the next command" convention — you can see, and
   gate in CI, exactly when governance was recomputed.

---

## 3. Drift and recovery

`workspace status` classifies every repo as **clean**, **drifted** (manifest pin ≠
lock — you edited the pin, run `sync`), or **missing** (never synced / slice files
absent — run `sync`). A clean board ends with `next: company-os validate`; anything
else ends with `next: company-os workspace sync`.

### A hand-edited slice fails gate `[8/8]`

Slices are read-only derived content. If someone forces a slice writable and edits
it, its content hash diverges from the lock and `validate` blocks the merge
(`GPF-R-7.5`):

```console
$ chmod u+w platforms/communications/governance/requirements.yaml
$ echo '# unauthorized local tweak' >> platforms/communications/governance/requirements.yaml
$ company-os validate
…
[8/8] federated slice integrity (read-only derived content)
  [FAIL] federated slice hand-edited: platforms/communications/governance/requirements.yaml — content hash differs from workspace.lock.yaml; slices are read-only derived content — re-sync: company-os workspace sync

FAIL — 1 problem(s)
```

**Remedy — re-sync overwrites the slice from the pinned source and restores integrity:**

```console
$ company-os workspace sync
  synced company-os @ 71ab0661805d (tag v1.0.0) -> company-os (2 file(s))
  synced platform-communications @ 6622076972b9 (tag v2.1.0) -> platforms/communications (7 file(s))
wrote workspace.lock.yaml (2 repo(s))

$ company-os validate
…
[8/8] federated slice integrity (read-only derived content)
  [ok] federated slices match workspace.lock.yaml (9 file(s) across 2 repo(s); no hand-edits)

PASS
```

The lesson to encode in a real change: don't edit slices. Change the source repo,
bump the pin, re-sync. A slice edit is never the answer.

### Missing lock coverage

Gate `[8/8]` also fires when `workspace.lock.yaml` is absent, malformed, or does
not cover a manifest repo ("lock does not cover the manifest") — run `sync`. And a
`--frozen` run refuses outright when the lock is missing:

```console
$ company-os workspace sync --frozen
error: --frozen requires workspace.lock.yaml, but it is missing or malformed at /…/ws. run online first: company-os workspace sync
```

### Rejected pins fail fast (at load, before any git)

```console
$ company-os workspace status      # pin: {branch: main}
error: repo 'platform-communications': pin key(s) ['branch'] are floating and non-reproducible; use commit:/tag: only, then: company-os workspace sync (GPF-R-6.3)

$ company-os workspace status      # pin: {tag: v2.1.0, commit: 6622076…}
error: repo 'platform-communications': pin must set EXACTLY ONE of commit:/tag: (got ['commit', 'tag']). use commit:/tag: only
```

---

## 4. Ownership transfer is two PRs

The component **descriptor** (`platforms/<p>/components/<id>.yaml`,
`ownership.accountableTeam`) is the single source of truth for the accountable
team; a team's `ownership/components.yaml` registry only *claims* what the
descriptor must confirm. Validate gate `[1/8]` fails if a team claims `accountable`
but the descriptor names a different team. Moving a component between teams
therefore touches two repos and is two PRs:

| PR | Repo | Change |
|---|---|---|
| **A** | platform repo | Descriptor `ownership.accountableTeam: team://new-team` |
| **B** | team repos | New team **adds** the component to its `ownership/components.yaml` (`relationship: accountable`); old team **removes** it |

**Pinned refs make the interim benign.** Every consumer stays on the *old* platform
pin until PR A merges and the workspace bumps that pin — so there is no window where
the descriptor and the registries disagree in a live workspace:

1. Merge **PR A** (platform) and **PR B** (teams) in either order — they touch
   different repos and neither is visible to a workspace until its pin advances.
2. In the workspace repo, bump the platform pin to the release containing PR A and
   `workspace sync`. The new descriptor arrives in the slice, and both team
   registries already agree (PR B). Gate `[1/8]` reconciles atomically from the
   workspace's point of view — one manifest bump flips ownership cleanly.

No tooling is required for v1; this ordering *is* the runbook. Do not bump the
platform pin before PR B has merged in both team repos, or the first `validate`
after sync will fail `[1/8]` until the registries catch up.

---

## 5. Per-repo CI recipe

A platform or team *source* repo, on its own, cannot run `company-os validate` — it
is not a workspace root (it holds only governance directories at its root). CI for
these repos composes a workspace around the PR content. Two patterns, both fully
offline, both verified against the current CLI.

### Pattern A — the workspace (composition) repo's own CI

The committed slices and lock are already on disk after `git checkout`. `validate`
verifies them against the lock with **no network and no git cache** — gate `[8/8]`
is a pure hash check. This is the whole job:

```yaml
# .github/workflows/os-validate.yml  (in the workspace/composition repo)
name: os-validate
on: [pull_request]
jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - run: pip install pyyaml
      - run: bin/company-os --root . validate      # committed slices vs committed lock; offline
      # optional: prove generated/ is truly derived (same pattern as monorepo CI)
      - run: |
          bin/company-os governance resolve --team customer-engagement
          git diff --exit-code teams/*/generated/
```

### Pattern B — a downstream platform/team source-repo PR

The source repo's CI needs the *rest* of the governance to validate its proposed
change. Check out the composition repo (for the committed sibling slices, lock, and
native team dirs), overlay the PR checkout into its own slice path **as native
content**, and — the key step — run against a manifest that **omits the PR's own
repo**, so gate `[8/8]` only hash-checks the untouched siblings while gates
`[1/8]`–`[7/8]` fully validate the proposed content:

```yaml
# .github/workflows/os-validate.yml  (in a platform/team SOURCE repo)
name: os-validate
on: [pull_request]
jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4          # the PR: this repo's governance content
        with: {path: pr}
      - uses: actions/checkout@v4          # the composition repo: siblings + lock + native dirs
        with: {repository: acme/company-workspace, path: ws}
      - run: pip install pyyaml

      # Build a CI manifest that drops THIS repo (so its content is validated
      # natively, not hash-checked against the lock as a slice).
      - name: compose
        working-directory: ws
        run: |
          python - <<'PY'
          import yaml
          m = yaml.safe_load(open("workspace.yaml"))
          m["repos"] = [r for r in m["repos"] if r["name"] != "platform-communications"]
          yaml.safe_dump(m, open("workspace.yaml", "w"))
          PY
          # overlay the PR content at this repo's canonical root, as native (writable) files
          rm -rf platforms/communications
          mkdir -p platforms/communications
          cp -R ../pr/. platforms/communications/
          bin/company-os graph build        # keep feature-index / CLAUDE.md nodes current

      - name: validate
        working-directory: ws
        run: bin/company-os --root . validate
```

Verified result of Pattern B — the PR's content passes gates `[1/8]`–`[7/8]`, and
`[8/8]` checks only the remaining sibling slice:

```console
[8/8] federated slice integrity (read-only derived content)
  [ok] federated slices match workspace.lock.yaml (2 file(s) across 1 repo(s); no hand-edits)

PASS
```

> **On `sync --frozen` in CI.** `--frozen` (see §6) rebuilds slices from the local
> `.company-os/` git cache and verifies them against the lock — it needs the cache,
> which is git-ignored and therefore *absent* in a clean runner. It is not a
> from-the-lock-alone reconstruction. So for a fresh checkout the committed slices
> already suffice (Pattern A) and you do not need `sync` at all. If you *want*
> `sync --frozen` in CI (e.g. to re-prove materialization determinism), cache
> `.company-os/` between runs, keyed on `workspace.lock.yaml`, and run one online
> `sync` to seed it first.

---

## 6. Git version, offline mode, and gotchas

### git ≥ 2.27, federation commands only

Federation shells out to git for cone-mode sparse-checkout and blobless partial
clones, which stabilized in **git 2.27**. The CLI guards this at the federation
entry points only — monorepo commands never call git (`GPF-R-7.7`):

```
git is required for federation (workspace sync/status) but was not found on PATH.
monorepo commands do not need git. install git, then: company-os workspace sync
```
```
git <x.y.z> is too old for federation; >= 2.27 is required (cone-mode
sparse-checkout + partial clone). upgrade git, then: company-os workspace sync
```

### `--frozen` for network-restricted environments

`--frozen` performs no network access: it re-checks-out the pinned SHA from the
local `.company-os/` cache, re-materializes the slices, and asserts each slice hash
against the lock. Use it on a machine that has already synced online at least once
(the cache is populated):

```console
$ company-os workspace sync --frozen
workspace sync --frozen (2 repo(s))

  restored company-os @ 71ab0661805d (from lock) -> company-os (2 file(s))
  restored platform-communications @ 6622076972b9 (from lock) -> platforms/communications (7 file(s))

materialized strictly from workspace.lock.yaml (no network)
next: company-os workspace status   # then: company-os validate
```

It fails clearly if the lock is missing, if the cache for a repo is absent, or if
the re-materialized bytes don't match the lock hash. For an offline runner with no
cache, prefer validating the committed slices directly (§5, Pattern A).

### Gotchas worth pinning to a wall

- **`paths:` must list directories, not files.** Cone-mode sparse-checkout treats
  every allowlist entry as a directory; a top-level file entry (e.g. `CLAUDE.md`,
  `platform.yaml`) fails with `fatal: '<file>' is not a directory` (intermittently
  — a cold cache may slip through, a warm one won't). Slice whole directories.
  `generated/` brings `feature-index.yaml`; the platform's `CLAUDE.md` context node
  is absence-tolerant at the gate, so leaving it out of the slice is fine.
- **Never edit a slice.** It is `0444`/`0555` on purpose. Change the source repo,
  bump the pin, re-sync.
- **Derived artifacts are pinned, not computed locally.** `feature-index.yaml` and
  each `CLAUDE.md` generated block ride along in the slice at the pinned SHA and
  cannot be regenerated in the read-only workspace — keep them current in the
  source repo (`graph build` + `git diff --exit-code` in that repo's CI) before you
  cut the tag you pin.
- **`sync` and `resolve` are separate on purpose** (§2). Bumping a pin never
  regenerates a team's governance for you — run `governance resolve` and commit it.
- **Monorepo stays the default and the test baseline.** With no `workspace.yaml`,
  none of the above runs and `validate` output is byte-identical to the monorepo
  golden snapshot (`GPF-R-6.1`). Migrating monorepo → federated is additive: split
  each top-level directory into its own repo (no path changes — the layout is
  already repo-shaped), add `workspace.yaml`, `sync`. Reversible by inverting.
```
signal ─▶ upstream tags a release ─▶ bump pin (workspace.yaml) ─▶ workspace sync
                                                                        │
                                                              workspace status (clean?)
                                                                        │
                                                     governance resolve --team <t>  (per team)
                                                                        │
                                              commit: manifest + slices + lock + generated/
                                                                        │
                                                            company-os validate  [1/8…8/8]
```
