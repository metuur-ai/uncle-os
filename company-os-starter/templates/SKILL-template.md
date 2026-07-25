---
id: skill://<scope>/<name>
type: skill
version: '1.0'
authority: canonical
appliesTo: ['<platform:// or team:// id>']
inputs: [<what must exist before this skill runs>]
outputs: [<artifact that must pass which validation command>]
tags: [authority/canonical]
---

# <Skill title>

<!-- File name must be <name>.SKILL.md, directly inside a skills/ directory —
     discovery globs *.SKILL.md one level deep, so <name>/SKILL.md is invisible.

     `type: skill` is required: validate's frontmatter gate fails without it.
     `tags:` are DERIVED — run `company-os graph build` and let it write them.
     Hand-written facets like kind/skill are not derivable and fail the gate.

     Numbered steps. Mark each step (mandatory) / (default) / (guidance).
     Put the action on the numbered line itself: only that head line is parsed,
     so continuation lines never appear in `company-os skills list`.

     Agents apply personal rules from scratchpad/personal-rules/ ON TOP of
     this skill; mandatory steps always win over personal rules. -->
