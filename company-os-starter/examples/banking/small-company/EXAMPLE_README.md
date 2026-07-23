# Example: `small-company/` — pattern P8, the ~10-person startup

**What it demonstrates:** the minimum honest deployment. One monorepo
workspace, **no `workspace.yaml`** (federation is manifest-optional: no
manifest, zero federation overhead), one platform (`product`), one team
(`core`), a four-entry ID registry, and a two-control company baseline.

**People wear multiple hats:** the roster is two named people; Ana is
team-lead *and* acts as product-owner — same human, different `today --role`
views. The team declares one deviation (`company-standard://change-log`,
default tier): "the PR description is the change log" — comply-or-explain at
its smallest.

**The growth path (P9):** every directory here keeps its meaning at bank
scale. To federate later: move `platforms/product/` into its own repo, add a
`workspace.yaml` entry with a pin, run `company-os workspace sync`. Compare
side-by-side with `../bank/workspaces/team-fraud-detection/` — same canonical
roots, platform dirs become pinned read-only slices.

**Layout:** `company-os/standards/` · `company-ontology/ids/` ·
`platforms/product/{platform.yaml,components,governance,reality}` ·
`teams/core/{team.yaml,ownership,governance}`.
