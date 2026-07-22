# Templates are examples, not contracts

These files show *one* way to write each artifact. The shared operating-system
contract is the **frontmatter core** (`docs/FRONTMATTER-CORE.md`) plus the
lifecycle gates — not the section headings below the frontmatter.

Teams may keep their own document styles (ADR vs RFC vs free-form) as long as
each doc carries the core: identity (`type`, `id`), lifecycle (`status`, or
`updated`/`authority` for reality docs), and reference fields (`team`,
`platform`, `components`, …) that `graph build` turns into Obsidian tags and
wikilinks.

Section-structure checks in `discover validate` and `prd validate` are
**warnings by default**. A team can make them blocking for itself with
`teams/<team>/standards/doc-formats.yaml`:

```yaml
schemaVersion: "1.0"
enforce: true
```
