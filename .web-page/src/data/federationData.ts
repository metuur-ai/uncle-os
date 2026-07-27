export interface FederatedSource {
  id: string;
  name: string;
  repoUrl: string;
  pin: string;
  paths: string[];
  targets: string[];
  layer: 'company-os' | 'platforms' | 'company-ontology' | 'knowledge' | 'teams';
  owner: string;
  note: string;
}

/**
 * One workspace root, five different origins. Every directory below is the SAME
 * tree the Folder Explorer shows — the difference is only where the bytes came
 * from. `teams/web/` is deliberately absent: it is native, hand-edited content.
 */
export const FEDERATED_SOURCES: FederatedSource[] = [
  {
    id: 'src-company-os',
    name: 'company-os',
    repoUrl: 'github.com/moonbeam/company-os',
    pin: 'tag v2026.2',
    paths: ['standards/', 'onboarding/', 'skills/', 'templates/'],
    targets: ['company-os/'],
    layer: 'company-os',
    owner: 'Governance / VP Engineering',
    note: 'The company baseline lives in its own repo so governance can release it on its own cadence. Teams adopt a new baseline by bumping one tag.'
  },
  {
    id: 'src-ontology',
    name: 'company-ontology',
    repoUrl: 'github.com/moonbeam/company-ontology',
    pin: 'tag v2026.2',
    paths: ['ids/', 'contexts/', 'context-maps/', 'concepts/'],
    targets: ['company-ontology/'],
    layer: 'company-ontology',
    owner: 'Architecture Board',
    note: 'Canonical IDs must be identical in every workspace, so they are pulled from one repo rather than copied per team.'
  },
  {
    id: 'src-platform-ordering',
    name: 'platform-ordering',
    repoUrl: 'github.com/moonbeam/platform-ordering',
    pin: 'commit 6622076972b9…',
    paths: ['governance/', 'components/', 'reality/', 'change-records/', 'skills/'],
    targets: ['platforms/ordering/'],
    layer: 'platforms',
    owner: 'Ordering platform team',
    note: 'Each platform is its own repo with its own release train. Two platforms means two entries and two different destinations under platforms/.'
  },
  {
    id: 'src-platform-identity',
    name: 'platform-identity',
    repoUrl: 'github.com/moonbeam/platform-identity',
    pin: 'tag v4.3.1',
    paths: ['governance/', 'components/'],
    targets: ['platforms/identity/'],
    layer: 'platforms',
    owner: 'Identity platform team',
    note: 'A narrower allowlist is fine. You only pull the areas you actually need to reason about — this workspace never reads identity reality.'
  },
  {
    id: 'src-component-library',
    name: 'component-library',
    repoUrl: 'github.com/moonbeam/component-library',
    pin: 'tag v1.2.0',
    paths: ['docs/sdd', 'architecture', '.claude/skills'],
    targets: [
      'knowledge/components/component-library/',
      'knowledge/architecture/component-library/',
      'knowledge/skills/component-library/'
    ],
    layer: 'knowledge',
    owner: 'Component library maintainers',
    note: 'ONE repo, THREE destinations — this is what slices: is for. An ordinary product repo (not a Company OS workspace) contributes documentation only; its source code never lands here.'
  }
];

export const FEDERATION_MANIFEST_SAMPLE = `# workspace.yaml — at the workspace root. Its PRESENCE is what switches the
# workspace into federated mode. No workspace.yaml => plain monorepo, 7 gates.
#
# Read every entry as one sentence:
#   "take THESE paths, from THIS repo, at THIS exact pin,
#    and materialize them READ-ONLY at THIS localDirectory."
version: 1
repos:
  # 1. The company baseline: its own repo, its own release cadence.
  - name: company-os
    url: https://github.com/moonbeam/company-os.git
    localDirectory: company-os
    pin: {tag: v2026.2}
    paths: [standards/, onboarding/, skills/, templates/]

  # 2. Canonical IDs and bounded contexts: a second repo.
  - name: company-ontology
    url: https://github.com/moonbeam/company-ontology.git
    localDirectory: company-ontology
    pin: {tag: v2026.2}
    paths: [ids/, contexts/, context-maps/, concepts/]

  # 3. Each platform is a separate repo -> a separate destination
  #    under platforms/. Pin by commit when there is no release tag yet.
  - name: platform-ordering
    url: https://github.com/moonbeam/platform-ordering.git
    localDirectory: platforms/ordering
    pin: {commit: 6622076972b9f82b10e03ba584aa9199f77e9eb7}
    paths: [governance/, components/, reality/, change-records/, skills/]

  - name: platform-identity
    url: https://github.com/moonbeam/platform-identity.git
    localDirectory: platforms/identity
    pin: {tag: v4.3.1}
    paths: [governance/, components/]

  # 4. ONE repo -> MANY local directories. Use slices: instead of the
  #    top-level localDirectory:/paths: pair (setting both is rejected).
  #    The repo is cloned once, checked out once, materialized three times.
  - name: component-library
    url: https://github.com/moonbeam/component-library.git
    pin: {tag: v1.2.0}
    slices:
      - {paths: [docs/sdd],       localDirectory: knowledge/components/component-library}
      - {paths: [architecture],   localDirectory: knowledge/architecture/component-library}
      - {paths: [.claude/skills], localDirectory: knowledge/skills/component-library}

# teams/web/ is NOT in this file. It is native content: this repo owns it,
# humans edit it, and \`governance resolve\` writes its generated/ folder.`;

export const FEDERATION_SYNC_SAMPLE = `$ company-os workspace sync
workspace sync (5 repo(s))

  synced company-os @ 71ab0661805d (tag v2026.2) -> company-os (14 file(s))
  synced company-ontology @ 8f3c2d19ab04 (tag v2026.2) -> company-ontology (9 file(s))
  synced platform-ordering @ 6622076972b9 (commit 6622076972b9f82b10e03ba584aa9199f77e9eb7) -> platforms/ordering (23 file(s))
  synced platform-identity @ c7241463de65 (tag v4.3.1) -> platforms/identity (11 file(s))
  synced component-library @ 5d257bf542fa (tag v1.2.0) -> knowledge/components/component-library, knowledge/architecture/component-library, knowledge/skills/component-library (18 file(s))

wrote workspace.lock.yaml (5 repo(s))
next: company-os workspace status   # then: company-os validate

$ company-os workspace status
workspace federation status (5 repo(s))

  company-os: tag:v2026.2 @ 71ab0661805d -> company-os — clean
  company-ontology: tag:v2026.2 @ 8f3c2d19ab04 -> company-ontology — clean
  platform-ordering: commit:6622076972b9f82b10e03ba584aa9199f77e9eb7 @ 6622076972b9 -> platforms/ordering — clean
  platform-identity: tag:v4.3.1 @ c7241463de65 -> platforms/identity — clean
  component-library: tag:v1.2.0 @ 5d257bf542fa -> knowledge/components/component-library, knowledge/architecture/component-library, knowledge/skills/component-library — clean

next: company-os validate

# The slices are read-only on disk. This is a feature, not a permissions bug:
$ ls -l platforms/ordering/governance/requirements.yaml
-r--r--r--  1 dev  staff  4210 requirements.yaml

# Bump one line in workspace.yaml (v4.3.1 -> v4.4.0) and status says so
# BEFORE you sync — the manifest pin no longer matches the lock:
$ company-os workspace status
  platform-identity: tag:v4.4.0 @ c7241463de65 -> platforms/identity — drifted (manifest pin tag:v4.4.0 != lock {'tag': 'v4.3.1'})

next: company-os workspace sync`;

export const FEDERATION_LOCK_SAMPLE = `# workspace.lock.yaml — machine-owned, written by \`workspace sync\`, committed
# alongside the slices. It is the oracle for validate gate [8/8]: the recorded
# slice set catches a MOVED target, the per-file hashes catch a HAND-EDIT.
repos:
- name: platform-ordering
  url: https://github.com/moonbeam/platform-ordering.git
  slices:
  - localDirectory: platforms/ordering
    paths: [governance/, components/, reality/, change-records/, skills/]
  pin:
    commit: 6622076972b9f82b10e03ba584aa9199f77e9eb7
  resolvedCommit: 6622076972b9f82b10e03ba584aa9199f77e9eb7
  sliceHash: a2a6445fe15e0b81ed0b2554466b9299a5ed895e236df0cb86cad0b660751b08
  files:
    platforms/ordering/governance/requirements.yaml: 7481ac4c3b661dec…
    platforms/ordering/components/checkout-api.yaml: 9047752c1f0ab33e…

# A tag: pin is resolved to a SHA here — that is how a moving tag can never
# silently change what your workspace contains.
- name: component-library
  url: https://github.com/moonbeam/component-library.git
  slices:
  - localDirectory: knowledge/components/component-library
    paths: [docs/sdd]
  - localDirectory: knowledge/architecture/component-library
    paths: [architecture]
  - localDirectory: knowledge/skills/component-library
    paths: [.claude/skills]
  pin:
    tag: v1.2.0
  resolvedCommit: 5d257bf542fae1c0d2a9b8e77c4f3a1b6e0d9c82
  sliceHash: 1c9f0b2e7a5d84361fbb0e5c9a7d2f48e3610cb5d47a9e2f80b6c1d3a5e79f04
  files:
    knowledge/components/component-library/docs/sdd/overview.md: b31f7ae90c…

# Committed together, workspace.yaml + the slices + this lock let a fresh
# checkout run \`company-os validate\` with NO network and NO source repos.
# In CI use \`workspace sync --frozen\`: materialize from the lock, never resolve.`;

/* ------------------------------------------------------------------ */
/* GitHub MCP: an agent-side tool, reachable only through the skills.  */
/* ------------------------------------------------------------------ */

/**
 * Company OS ships no MCP server and no MCP client — there is no .mcp.json
 * anywhere in the repo, and no company-os command talks to the GitHub API.
 * `workspace sync` is these eight git invocations and nothing else.
 */
export const MCP_SYNC_GIT_VERBS = [
  'git --version                                  # floor is 2.27',
  'git init --quiet / remote add|set-url origin',
  'git fetch --filter=blob:none --depth 1 <ref>   # blobless, shallow',
  'git rev-parse FETCH_HEAD^{commit}              # pin resolved to a SHA',
  'git sparse-checkout init --cone / set <dirs>',
  'git checkout --quiet --detach <sha>'
];

export interface McpBoundaryColumn {
  id: 'reads' | 'writes' | 'never';
  title: string;
  subtitle: string;
  items: string[];
}

export const MCP_BOUNDARY: McpBoundaryColumn[] = [
  {
    id: 'reads',
    title: 'GitHub MCP may read',
    subtitle: 'Your agent\'s tool. No Company OS invariant is at risk.',
    items: [
      'Browse a source repo\'s tree to decide what belongs in paths:',
      'List tags and commits to choose the pin',
      'Diff two pins to see what a bump would bring in',
      'Open the PR that bumps pin: on the workspace repo',
      'Open the PR that fixes the docs in the SOURCE repo — the only correct place to fix a slice\'s content'
    ]
  },
  {
    id: 'writes',
    title: 'Only the CLI writes',
    subtitle: 'The workspace and the lock have exactly one author.',
    items: [
      'company-os workspace sync — the only thing that may fetch content in',
      'company-os governance resolve — owns generated/effective-governance.yaml',
      'company-os graph build — owns derived tags and CLAUDE.md context nodes',
      'company-os validate — re-hashes every slice offline at gate [8/8]'
    ]
  },
  {
    id: 'never',
    title: 'Never, by any agent',
    subtitle: 'Cross this line and validation stops being evidence of anything.',
    items: [
      'Write into a materialized slice — it is 0444 derived content; gate [8/8] fails and names the file',
      'Write anything under generated/ — CI regenerates and diffs it',
      'Fetch content into the workspace another way — it has no recorded provenance, so the hash check either fails or, worse, passes meaninglessly'
    ]
  }
];

export interface McpSkillStep {
  id: string;
  moment: string;
  mcp: string;
  cli: string;
}

/**
 * The syncing-knowledge skill has three mandatory steps. MCP helps around
 * them — never inside them.
 */
export const MCP_SKILL_STEPS: McpSkillStep[] = [
  {
    id: 'mcp-step-0',
    moment: 'Before step 1 — decide what to pin',
    mcp: 'Read the source repo: where do the docs live, which tags exist, what changed between v1.1.0 and v1.2.0?',
    cli: '—'
  },
  {
    id: 'mcp-step-1',
    moment: '1. Declare the entry in workspace.yaml',
    mcp: '—',
    cli: 'You (or the agent, locally) author url:, paths: and localDirectory:. Nothing is fetched yet.'
  },
  {
    id: 'mcp-step-2',
    moment: '2. company-os workspace sync',
    mcp: '—',
    cli: 'Eight git commands materialize the slice 0444 and write the per-file hashes into workspace.lock.yaml.'
  },
  {
    id: 'mcp-step-3',
    moment: '3. graph build → validate',
    mcp: '—',
    cli: 'Regenerates the knowledge/CLAUDE.md context node, then gate [8/8] re-hashes the tree offline.'
  },
  {
    id: 'mcp-step-4',
    moment: 'After — land it',
    mcp: 'Open the PR carrying the manifest, the slice and the lock. Upstream doc fixes go as a separate PR to the source repo.',
    cli: '—'
  }
];

export const MCP_ASSISTED_PROMPT = `Use the GitHub MCP to read github.com/acme/component-library: show me where the SDD docs live and the last three tags, and diff v1.1.0..v1.2.0 so I can see what a bump brings in. Then follow skill://governance/syncing-knowledge — write the workspace.yaml entry for docs/sdd -> knowledge/components/component-library pinned to the tag we pick, run workspace status, then sync, then graph build, then validate. Use MCP for reading only: do not fetch any file into the workspace and do not touch anything under knowledge/ — workspace sync is the only writer, and the lock is the oracle.`;

export interface FederationRule {
  id: string;
  rule: string;
  why: string;
  breaksWith: string;
}

export const FEDERATION_RULES: FederationRule[] = [
  {
    id: 'rule-pin',
    rule: 'Exactly one of commit: or tag: — never a branch',
    why: 'Governance you cannot reproduce is not governance. A branch would let upstream change your rules while you sleep.',
    breaksWith: 'A branch name, both keys at once, or a bare ref: is rejected at load — sync, status and validate all refuse the manifest.'
  },
  {
    id: 'rule-readonly',
    rule: 'Slices are materialized 0444 / 0555 — never edit one',
    why: 'A slice is derived content, governed exactly like generated/. The authoritative file lives in the source repo.',
    breaksWith: 'Gate [8/8] compares every file against the lock hash map and fails on any hand-edit. Fix upstream, bump the pin, re-sync.'
  },
  {
    id: 'rule-disjoint',
    rule: 'Slice targets must be disjoint across the whole manifest',
    why: 'An outer slice\'s read-only pass would freeze a nested slice\'s parent directory and break the next sync.',
    breaksWith: 'Equal or nested localDirectory targets are refused, even across two different repos.'
  },
  {
    id: 'rule-roots',
    rule: 'A destination must land under a canonical root',
    why: 'The directory layout is the contract; the repo boundary is not. Every tool still finds artifacts in the same place.',
    breaksWith: 'localDirectory must sit under company-os/, platforms/, teams/, company-ontology/ or knowledge/. Bare knowledge/ is refused — the catalog root is CLI-owned.'
  },
  {
    id: 'rule-resolve',
    rule: 'sync does not run governance resolve',
    why: 'sync writes machine-owned slices; resolve writes team-owned generated/ that a human reviews. Fusing them would smuggle a governance decision into a materialization step.',
    breaksWith: 'New platform requirements do not reach a team until you run company-os governance resolve --team <t> yourself.'
  },
  {
    id: 'rule-knowledge',
    rule: 'knowledge/ is indexed, not governed',
    why: 'Foreign docs from ordinary product repos carry no type: frontmatter, and the slice is read-only so tags cannot be rewritten.',
    breaksWith: 'graph build and gates [1/8]–[7/8] skip knowledge/; it still gets a CLAUDE.md index and gate [8/8] hash integrity.'
  }
];
