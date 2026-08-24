---
type: doc
tags: [doc/company-os-starter, kind/explanation]
---

# Obsidian, Local Search, and the workspace

Company OS never exports to Obsidian and never pushes an index to Local
Search. There is no sync step, no plugin, no adapter. All three tools open
the same directory of Markdown files, and each reads the part of those files
it understands.

This page explains what is actually shared, what each tool does with it, and
which pieces are aspiration rather than shipped behaviour.

## Who owns what

| | Reads | Writes | Needs |
|---|---|---|---|
| `company-os` | the whole workspace | frontmatter `tags:`, generated blocks in `CLAUDE.md`, `generated/feature-index.yaml` | nothing |
| Local Search | `.md` / `.mdx` / `.txt` under a registered repo | its own SQLite index, outside the workspace | one `local-search repo add` |
| Obsidian | the vault directory | `.obsidian/` config, plus whatever you type | one "Open folder as vault" |

Only `company-os` writes into the workspace. That is the property that keeps
the other two safe to attach and detach at any time.

## The seam is `company-os derive`

`derive` is the only command that turns your authored frontmatter into
things the read-side tools consume. It produces exactly three artifacts:

**1. Frontmatter `tags:`** — derived from `type:`, `team:`, `platform:`,
`status:`, and the canonical IDs in `ids/registry.yaml`. Flat YAML strings
with slashes:

```yaml
tags: [kind/discovery, status/validated, team/customer-engagement]
```

**2. Generated blocks in each `CLAUDE.md`** — a doc index and a federation
index, written between markers and using **relative Markdown links**:

```markdown
<!-- company-os:generated:start — do not edit; run `company-os derive` -->
## communications — context index

### Doc index
- **prd**
  - [Per-channel quiet hours](archive/prds/2026-per-channel-quiet-hours/prd.md)

### Federation roots
- [company-os](../../company-os/CLAUDE.md)
<!-- company-os:generated:end -->
```

**3. `platforms/<p>/generated/feature-index.yaml`** — the derived component
and pointer rollup. This one is YAML, so neither read-side tool sees it;
it exists for the CLI's own gates.

## What Obsidian actually gives you

Point Obsidian at the workspace root and the vault *is* the workspace — same
files, same Git history, no copy.

- **Tag pane.** Obsidian reads YAML `tags:` and treats `/` as nesting, so
  the derived tags above appear as a browsable `kind/` → `discovery` tree
  without you writing a single `#`.
- **Backlinks and graph view.** These follow the relative Markdown links in
  the generated `CLAUDE.md` blocks. Every platform, team, and ontology root
  links to its documents and to its sibling roots, so `CLAUDE.md` files
  become the hubs of the graph view — which is roughly the shape of the
  federation itself.
- **Aliases.** `aliases:` in frontmatter is honoured by Obsidian's quick
  switcher, and Company OS neither writes nor strips it.

That is the whole integration. There is nothing to install and nothing to
keep in sync.

## What Local Search actually gives you

Register the workspace once and Local Search maintains its own index in its
own directory:

```bash
local-search repo add company-os .
```

It then offers three things the file system cannot:

- **Search that tolerates wording.** SQLite FTS5 for exact terms, plus
  vector re-ranking, so "how do we handle retries" finds a document that
  says "delivery backoff".
- **A typed graph built from your frontmatter.** This is the part people
  miss: Local Search does not need wikilinks to build edges. It reads
  Company OS reference fields directly and gives each its own edge type:

  | Frontmatter field | Edge |
  |---|---|
  | `relationships` | `related_to` |
  | `components` | `has_component` |
  | `dependsOn` | `depends_on` |
  | `upstream` / `downstream` | `upstream` / `downstream` |
  | `implementedBy` | `implements` (reversed) |
  | `fromDiscovery` | `from_discovery` |
  | `boundedContext` | `in_context` |

  Unknown fields never error — they are reported in the scan summary and
  ignored, which is the same open-world rule the frontmatter core states.
- **`@spec` and link edges.** Body Markdown links become link edges, and
  `@spec req://…` markers become spec references, so requirement-to-code
  tracing is greppable *and* queryable.

If you do write `[[wikilinks]]` by hand, Local Search extracts them too
(including `[[target#heading|alias]]`), and it deliberately ignores `[[ … ]]`
inside fenced code blocks so shell examples don't pollute the graph.

## Three things that are not true

Older pages and the original proposal imply more integration than exists.

- **`derive` does not write wikilinks.** It writes `tags:`, relative
  Markdown links inside generated blocks, and `feature-index.yaml`. Nothing
  else. Obsidian's graph works because relative Markdown links work, not
  because anything was converted for it.
- **`COMPANY_OS_VAULT_ROOT` is not read by anything.** It appears in
  `docs/00-original-proposal.md` as a sketch of a vault that lives apart
  from the repo. The shipped CLI resolves one root: `--root`, then
  `$COMPANY_OS_WORKSPACE_ROOT`, then the current directory.
- **`generated/*.yaml` is invisible to both tools.** Local Search indexes
  text files, not YAML; Obsidian shows YAML as an unopenable file. If you
  want a derived rollup to be searchable, it has to be Markdown.

## Editing in Obsidian without breaking the gates

Obsidian is a text editor with no idea that some regions are owned. Two
rules keep it honest:

1. **Never hand-write `tags:`.** Change the source field — `type:`,
   `team:`, `status:` — and run `company-os derive`. Gate `[4/7]` compares
   what `derive` would produce against what is on disk and fails on drift.
2. **Never type inside the `company-os:generated` markers.** Gate `[5/7]`
   detects an edited or missing block. Everything outside the markers in a
   `CLAUDE.md` is yours and is preserved across every rebuild.

`company-os validate --fix` repairs both classes of drift in place rather
than just reporting them, which is the intended recovery after an
enthusiastic editing session.

## Why they stay separate binaries

Merging the search engine into `company-os` would buy a shorter install and
cost the property that makes this work: Local Search indexes *any* Markdown
corpus, and Company OS validates *one* kind of workspace. Keeping them apart
means the workspace stays a plain directory that any tool can read — which
is the same reason Obsidian needs no plugin.

See [how-it-fits-together.md](how-it-fits-together.md) for where these two
sit relative to Company OS and the Claude skills, and
[../how-to/open-your-workspace-in-obsidian.md](../how-to/open-your-workspace-in-obsidian.md)
for the actual setup steps.
