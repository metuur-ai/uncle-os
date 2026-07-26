package scaffold

// The three markdown scaffolds `add team` and `init` write for a new team.
//
// Unlike DISCOVERY_TEMPLATE and PRD_TEMPLATE (internal/scaffold/template.go),
// these are NOT reachable through resolve_template: scaffold_team formats them
// directly (bin/company-os:1868-1876), so no workspace override applies and no
// provenance label is printed for them. They are therefore plain constants with
// no entry in builtinTemplate.
//
// Each carries its `tags:` pre-derived to match derive_tags(), so the
// `graph build` inside a scaffold is a no-op on them and only the CLAUDE.md
// nodes and feature-index are newly written (bin/company-os:1896-1898).
//
// Python formats them with str.format on a single {tid}/{tname} field, which is
// why the port substitutes rather than templates: nothing else in the text is
// braced.

// dorTemplate is DOR_TEMPLATE (bin/company-os:1901-1913) verbatim.
const dorTemplate = `---
id: team-standard://{tid}/definition-of-ready
type: team-standard
tags: [team/{tid}]
---

## Team Definition of Ready
- The problem and expected outcome are clear.
- The affected components are identified.
- Dependencies are known.
- Acceptance criteria are testable.
- The work is small enough to execute.
`

// dodTemplate is DOD_TEMPLATE (bin/company-os:1915-1926) verbatim.
const dodTemplate = `---
id: team-standard://{tid}/definition-of-done
type: team-standard
tags: [team/{tid}]
---

## Team Definition of Done
- Implementation is complete and tests pass.
- Peer review is complete.
- Operational documentation is updated.
- Relevant Representation of Reality documents are updated.
`

// onboardingTemplate is ONBOARDING_TEMPLATE (bin/company-os:1928-1943) verbatim.
const onboardingTemplate = `---
type: onboarding-guide
id: onboarding-{tid}-developer
role: developer
team: {tid}
tags: [kind/onboarding, role/developer, team/{tid}]
---

# Developer onboarding ({tname})

Start here. Your team's effective governance is composed from the company
baseline, the platforms your components belong to, and this team's standards.

Daily view:  company-os today --role developer
`
