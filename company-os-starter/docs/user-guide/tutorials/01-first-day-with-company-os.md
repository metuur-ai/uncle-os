---
title: First day with Company OS
---

# First day with Company OS

You're the first person at **Moonbeam Bakery** to bring Company OS in. By the
end of this tutorial you'll have a real workspace, a real change moving
through discovery → PRD → done, and a green `company-os validate`. Budget
about 25 minutes.

> **Tip:** Company OS has no server, no database, and no account to create.
> Everything you're about to do is files in a folder plus a single-binary CLI
> that guides you through them.

## 1. Install the CLI

`company-os` is a single static binary. There is nothing to install
underneath it — no interpreter, no runtime, no package manager. Download the
one for your platform, make it executable, and put it somewhere on your
`PATH`:

```bash
# pick the artifact matching your machine, e.g.
#   company-os_<version>_darwin_arm64   (Apple Silicon Mac)
#   company-os_<version>_darwin_amd64   (Intel Mac)
#   company-os_<version>_linux_amd64
curl -fsSLo company-os <release-url>/company-os_<version>_darwin_arm64
chmod +x company-os
mkdir -p ~/.local/bin && mv company-os ~/.local/bin/
```

<!-- INSTALL-CAVEATS: macOS Gatekeeper / quarantine guidance is owned by task
     5.2 and lands here. Do not duplicate it elsewhere. -->

### macOS: the first run will be blocked

**This applies on macOS only, and only to a binary you downloaded.** Releases
are not yet Apple-notarized. macOS tags anything a browser downloads with a
`com.apple.quarantine` attribute, and Gatekeeper refuses to execute a
quarantined binary it cannot verify. Run it and it does not fail — it hangs,
and nothing is printed, which is a worse first impression than an error would
be. Launched from Finder you get "cannot be opened because the developer cannot
be verified" instead.

Clear the attribute once, right after moving the binary into place:

```bash
xattr -d com.apple.quarantine ~/.local/bin/company-os
```

That is the whole workaround. It is per-download, so you will do it again the
next time you upgrade by downloading. It is not needed at all if you built from
source, if you fetched the binary with `curl` rather than a browser, or on
Linux.

Check the attribute is gone before you blame anything else:

```bash
xattr ~/.local/bin/company-os      # prints nothing (or only com.apple.provenance)
```

Notarization is the real fix and is
[documented for maintainers](../how-to/release-and-upgrade.md#macos-signing-and-notarization).
Until it happens, this step is an accepted cost, recorded as such in the
project's design docs.

If `~/.local/bin` isn't already on your `PATH`, add it to your shell profile
now and restart your shell:

```bash
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.zshrc
```

> **Note:** Building from source instead? From `company-os-starter/`, run
> `make install` — it builds the binary and copies it to `~/.local/bin`
> (override with `PREFIX=/some/where`). Requires the Go toolchain, which is a
> build-time requirement only; the binary it produces has no dependencies. The
> rest of this tutorial works identically either way.

Confirm it's working:

```bash
$ company-os --help
```

You should see the full subcommand list (`init`, `add`, `discover`, `prd`,
`governance`, `validate`, and more — the full reference is in
[reference/company-os-cli.md](../reference/company-os-cli.md)).

## 2. Scaffold the workspace

Pick an empty directory and initialize it. Passing all three flags skips the
interactive prompts:

```bash
$ mkdir moonbeam-os && cd moonbeam-os
$ company-os init --company "Moonbeam Bakery" --team web --platform ordering
initialized workspace at /Users/you/moonbeam-os
  company: Moonbeam Bakery | first team: web | first platform: ordering
next: cd /Users/you/moonbeam-os && company-os discover new --team web "<discovery title>"
```

Look around — `init` scaffolded the four peer roots the whole system is
built on: `company-os/` (baseline standards), `platforms/ordering/` (the
platform you just named), `teams/web/` (the team you just named), and
`company-ontology/` (canonical IDs). This is the shape every Company OS
workspace starts from, whether it's one team or fifty.

## 3. Register the online ordering app

Moonbeam's `ordering` platform needs a component to own — the online
ordering web app:

```bash
$ company-os add component online-ordering-app --platform ordering
added component 'online-ordering-app' to platform 'ordering'
next: company-os reality new --platform ordering online-ordering-app

$ company-os reality new online-ordering-app --platform ordering
```

`reality new` scaffolds `platforms/ordering/reality/components/online-ordering-app.md`
— the file that will describe *current, true-today* behavior of the app. You'll
come back to it in step 7.

> **Tip:** `add` also grows the workspace sideways — `company-os add platform
> loyalty` or `company-os add team kitchen-ops` — when Moonbeam outgrows one
> platform and one team. See [how-to/grow-a-workspace.md](../how-to/grow-a-workspace.md).

## 4. Resolve the web team's governance

Before any work starts, see what rules actually apply to the `web` team
given the components it owns:

```bash
$ company-os governance resolve --team web
resolved governance for team 'web' (1 component(s))
wrote teams/web/generated/effective-governance.yaml
  online-ordering-app: platforms [ordering], N company + N platform requirement(s)
```

That file under `generated/` is derived — never hand-edit it. Re-run
`governance resolve` any time ownership or requirements change, and see
[how-to/check-your-work-against-governance.md](../how-to/check-your-work-against-governance.md)
for `governance explain`, which tells you *why* a specific rule applies.

## 5. Start discovery

Someone on the team wants same-day pickup slots for online orders. Capture
that as a discovery brief before writing any PRD:

```bash
$ company-os discover new "Same-day pickup slots" --team web
created teams/web/product/discovery/2026-same-day-pickup-slots/brief.md
next: fill Problem signal, Hypothesis, Success criteria, then run:
  company-os discover validate 2026-same-day-pickup-slots --team web
```

Try validating it empty first — the contract pushes back:

```bash
$ company-os discover validate 2026-same-day-pickup-slots --team web
  [FAIL] section 'Problem signal' is empty
  [FAIL] section 'Hypothesis' is empty
  [FAIL] section 'Success criteria' is empty
```

Open `brief.md`, fill the three required sections (*how* you research them —
customer interviews, order data, a prototype — is entirely up to you, that's
`guidance`-tier), and validate again:

```bash
$ company-os discover validate 2026-same-day-pickup-slots --team web
  [ok] brief '2026-same-day-pickup-slots' validated (status: validated)
```

## 6. Turn the brief into a PRD

A validated discovery brief is team-private. A PRD is the platform-visible
change record — it's what actually proposes changing `ordering`'s reality:

```bash
$ company-os prd new --team web --platform ordering \
    --components online-ordering-app \
    --from-discovery 2026-same-day-pickup-slots
created platforms/ordering/change-records/active/2026-same-day-pickup-slots/prd.md
```

Three things happened for free: the Problem statement and Success metrics
were copied over from the brief (no re-typing, no drift), a governance
snapshot was stamped so this PRD is judged against today's rules even if they
change next month, and the applicable governance checklist for
`online-ordering-app` was injected straight into the PRD.

```bash
$ company-os prd validate 2026-same-day-pickup-slots --platform ordering
  [FAIL] frontmatter field 'decisionOwner' missing or TODO
  [FAIL] section 'Proposed change' is empty

# ...fill decisionOwner and Proposed change in prd.md...

$ company-os prd validate 2026-same-day-pickup-slots --platform ordering
  [ok] PRD '2026-same-day-pickup-slots' passes the artifact contract
```

## 7. Check readiness, then close the loop

Before pulling this into a sprint, run the composable Definition of Ready —
team baseline plus resolved governance, generated on demand:

```bash
$ company-os check ready --team web --components online-ordering-app
== Team baseline (definition-of-ready.md) ==
...
== Applicable governance (online-ordering-app) ==
- [ ] ordering: <requirement> (mandatory) — evidence:
...
```

Build the feature, then try to close it out immediately:

```bash
$ company-os prd complete 2026-same-day-pickup-slots --platform ordering
done-check failed — a change is not done until reality is updated:
  [FAIL] reality doc for 'online-ordering-app' not updated since PRD created
```

> **Warning:** This is by design — the single most important rule in Company
> OS is that a change isn't done until the Representation of Reality is
> updated. Skipping it isn't a shortcut you can take.

Edit `platforms/ordering/reality/components/online-ordering-app.md` to
describe the new same-day pickup behavior, bump its `updated:` date, check
off the governance items in the PRD with evidence links, then complete again:

```bash
$ company-os prd complete 2026-same-day-pickup-slots --platform ordering
archived -> platforms/ordering/archive/prds/2026-same-day-pickup-slots
outcome review scheduled (due in 90 days)
appended platforms/ordering/log.md
```

The PRD is now history, reality reflects the new behavior, and an
`outcome.md` is waiting for real metrics in 90 days.

## 8. Confirm the whole workspace is green

```bash
$ company-os validate
[1/7] ownership reconciliation
  [ok] online-ordering-app: registry and descriptor agree (ordering)
[2/7] deviation and exception expiry
[3/7] active PRD contracts
[4/7] frontmatter core and tag derivation (interop contract)
[5/7] CLAUDE.md context node drift (fail-safe, absence-tolerant)
[6/7] feature-index drift (derived component->artifact map)
[7/7] custom skills layering (shadowing + extends resolution)
PASS
```

Seven gates today because there's no `workspace.yaml` federation manifest —
see [reference/company-os-cli.md](../reference/company-os-cli.md) for what
each gate checks and when an eighth appears.

## Where to next

- Make `company-os today --role developer` a daily habit — it's your
  role-aware view of what needs attention.
- [how-to/take-a-change-from-discovery-to-done.md](../how-to/take-a-change-from-discovery-to-done.md)
  is this same loop written as a repeatable recipe, without the walkthrough
  narration.
- Growing past one team and one platform?
  [tutorials/02-running-a-standalone-team.md](02-running-a-standalone-team.md)
  and [how-to/grow-a-workspace.md](../how-to/grow-a-workspace.md).
- Want your new workspace searchable from the terminal?
  [tutorials/03-search-your-workspace.md](03-search-your-workspace.md).
