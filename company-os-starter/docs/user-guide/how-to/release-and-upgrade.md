# Cut a release, and upgrade an existing install

Two audiences share this page. The first half is for whoever publishes
artifacts. The second half — [Upgrading](#upgrading-an-existing-install) and
[Version skew](#version-skew-across-a-shared-workspace) — is for everyone,
because a federated workspace is shared and the person next to you is probably
on a different build.

All `make` commands run from `company-os-starter/`.

---

## Build the release matrix

```bash
make release
```

Three artifacts land in `dist/`, plus `install.sh` and a `SHA256SUMS` covering
all four:

| Artifact | Runs on |
|---|---|
| `company-os-darwin-arm64` | Apple Silicon Macs |
| `company-os-darwin-amd64` | Intel Macs |
| `company-os-linux-amd64` | x86-64 Linux |

The names carry **no version**. `install.sh` resolves
`/releases/latest/download/company-os-<os>-<arch>`, and that URL needs a fixed
filename. The version lives in the release tag and is stamped into the binary —
`company-os --version` reports it, so an artifact is never ambiguous about what
it is.

There is deliberately no Windows build. Go cross-compiles to it for free, but
nothing here is tested there, so nothing is claimed.

`SHA256SUMS` is published next to the artifacts. Verify a download with:

```bash
shasum -a 256 -c SHA256SUMS --ignore-missing    # macOS
sha256sum -c SHA256SUMS --ignore-missing        # Linux
```

### The checksums are reproducible, and that is checked

```bash
make repro
```

builds the whole matrix twice — the second pass against a throwaway build
cache — and fails if any checksum moved. Two independent clean clones of the
same commit, at different paths, produce byte-identical artifacts.

Three things make that true, and all three are load-bearing:

- **`-trimpath`.** Without it the compiler records the absolute path of the
  checkout in the binary's line tables, which `-s -w` does not strip. Two
  clones of one commit at different paths then produce different bytes and the
  checksum stops meaning "this is that commit."
- **No build date.** `internal/model/buildinfo.go` carries `Version`, `Commit`,
  `GoVersion` and `Platform` and no timestamp, on purpose: a timestamp would
  make two builds of identical source differ, and it says nothing `Commit` does
  not already say precisely.
- **A pinned Go toolchain.** The compiler version is part of the input. Two
  different Go releases produce different bytes from the same source. Record
  the toolchain version with the release; `go version -m <artifact>` reads it
  back out of any artifact, on any host.

One caveat worth knowing before you go looking for it: `cmd/go` omits the
`-ldflags` build setting from an artifact's embedded metadata whenever
`-trimpath` is set, so `go version -m` will not show you the version stamp.
Read it from `--version` instead.

### Confirm the artifacts depend on nothing

```bash
make deps-check
```

This rebuilds the matrix and inspects **the release artifacts**, not a local
build. It asserts, per artifact: `CGO_ENABLED=0`, `-trimpath`, that the version
string reached the image, and that the linkage is what it should be. It exits
non-zero if any of that is false.

Be precise about what it can prove from one machine:

- **`go version -m`** reads metadata out of the artifact and is OS- and
  arch-independent, so it covers all three from anywhere.
- **`otool -L`** parses Mach-O and covers both darwin artifacts from either
  darwin host. It cannot read ELF — pointed at the linux artifact it prints
  `is not an object file`, which is not a pass and must not be read as one.
- **`ldd`** exists only on Linux. On a darwin release host the linux artifact's
  linkage is checked structurally with `file(1)` (`statically linked`) and
  confirmed for real only on the clean Linux box in
  [the acceptance procedure](#clean-machine-acceptance-procedure).

**"Statically linked" is literally true on Linux and approximately true on
macOS.** Apple ships no static `libSystem`, so a fully static Mach-O executable
does not exist. Every darwin artifact links `/usr/lib/libSystem.B.dylib` and
`/usr/lib/libresolv.9.dylib` no matter what `CGO_ENABLED` says. Both are part
of macOS itself and are present on every Mac, so the promise users care about —
nothing to install underneath the binary — holds. `deps-check` fails on any
*third* dylib.

---

## macOS: no signing, no Apple Developer ID

**The artifacts are unsigned and un-notarized, deliberately. Nothing in the
release or install path needs an Apple Developer account.**

`com.apple.quarantine` is attached by the *downloading application* — Safari,
Chrome, Mail — not by `curl`, `wget` or `tar`. Gatekeeper only adjudicates files
that carry it. `install.sh` fetches with `curl`, so the binary it places in
`~/.local/bin` has no quarantine attribute and executes with no prompt.

Verified on macOS 15 / arm64: an installed binary carries only
`com.apple.provenance` and runs. That is the same attribute profile as the
`local-search` CLI, which ships unsigned on the same basis.

**Therefore: publish the install line, never a download link.** A browser
download of the same bytes is the quarantined path, and it does not fail
cleanly — it hangs with no output. The `xattr -d com.apple.quarantine` fallback
for anyone who takes that route is documented in the install tutorial.

`spctl -a` reports `rejected` for these artifacts. That is expected, not a defect
to chase: `spctl` answers "would Gatekeeper admit this?", while the question that
matters is "will this execute?" — and Gatekeeper is not consulted for an
unquarantined file. Check `xattr`, not `spctl`.

### If someone reconsiders this later

There are no `sign` / `notarize` targets; they were removed along with the
`CODESIGN_IDENTITY` and `NOTARY_PROFILE` variables once the install path made
them unnecessary. Three findings from when signing *was* investigated, recorded
so the work is not repeated from scratch:

1. **Signing alone is not a partial fix.** A binary signed with a real Developer
   ID Application certificate, hardened runtime and secure timestamp verified
   clean under `codesign` — correct Designated Requirement, `TeamIdentifier`
   set, full Apple chain — and `spctl` still reported `rejected`,
   `source=Unnotarized Developer ID`, with the quarantined binary still hanging.
   Signing without notarization buys nothing a user can perceive.
2. **A bare Mach-O executable cannot be stapled.** There is nowhere in a flat
   file to store the ticket, and `notarytool` will not accept one either — hence
   the `ditto -c -k` zip for submission. Gatekeeper then resolves the ticket
   online on first run; offline first-run requires shipping a `.dmg` or `.pkg`
   and stapling that.
3. **Pass the certificate SHA-1, not the name.** `codesign --sign "<name>"`
   fails with `ambiguous` the moment a renewed and an expiring certificate share
   a name. The 40-char SHA-1 from `security find-identity -v -p codesigning` is
   the only reliable argument.


## Clean-machine acceptance procedure

R-6.4 is satisfied only by a **downloaded release artifact on a clean machine**.
Verification against a locally built binary explicitly does not count, and no
amount of local testing substitutes for this. Run it once per release, on two
machines, and record the result against task 5.3.

**Machine A: macOS arm64. Machine B: Linux amd64.** Both freshly imaged or a
fresh VM. Neither with a Python interpreter, a Go toolchain, or this repository
on it. Do not `scp` the binary from your workstation — installing something
other than what a user would get is how this check passes while the real path
is broken.

1. Confirm the machine is actually clean:

   ```bash
   command -v python python3 || echo "no python: good"
   command -v go            || echo "no go: good"
   ```

2. Install using **only the published one-liner**, exactly as a user would:

   ```bash
   curl -fsSL https://raw.githubusercontent.com/metuur-ai/uncle-os/main/company-os-starter/install.sh | bash
   ```

   On macOS, confirm the binary is not quarantined — this is the property the
   whole install path rests on, so assert it rather than assume it:

   ```bash
   xattr ~/.local/bin/company-os     # nothing, or only com.apple.provenance
   ```

   Ignore `spctl`; it reports `rejected` by design here. See
   [macOS signing and notarization](#macos-signing-and-notarization).

3. **Also test the browser path once**, because some users will take it whatever
   the docs say. Download the artifact for the platform in Safari, verify the
   checksum, and confirm the documented `xattr -d` step is sufficient:

   ```bash
   shasum -a 256 -c SHA256SUMS --ignore-missing   # or sha256sum -c on Linux
   ```

   Follow only the published instructions in
   [tutorials/01-first-day-with-company-os.md](../tutorials/01-first-day-with-company-os.md).
   If a step is missing from the docs, that is the finding.

4. Run the surface. Every command must behave as documented:

   ```bash
   company-os --version
   company-os --help
   mkdir demo && cd demo
   company-os init --company Acme --team core --platform web
   company-os graph build                    # init scaffolds; this derives
   company-os validate                       # exits 0
   company-os add platform api
   company-os add team payments
   company-os add component svc-pay --platform api
   company-os reality new --platform api svc-pay
   company-os governance resolve --team payments
   company-os governance explain svc-pay
   company-os discover new --team payments "First brief"
   company-os check ready  --team payments --components svc-pay
   company-os graph build && company-os graph build   # second run is a no-op
   company-os ids list
   company-os skills list
   company-os today --role developer
   company-os scratchpad init --repo .
   company-os validate --json                # valid JSON, carries a build object
   company-os validate; echo "exit=$?"       # documented exit code
   ```

5. Record: OS and version, artifact filename, checksum verified yes/no, whether
   any step needed a workaround, and the `--version` line the binary printed.

**Pass condition:** every command runs, no step required anything not in the
published docs, and no interpreter or library had to be installed.

### What was verified locally, and what it does not replace

The no-Python property has been pushed as far as a developer workstation
allows, and the result is real evidence — it is simply the wrong kind of
evidence for R-6.4.

Every invocation in the differential corpus that has a real fixture — 227 of
them, spanning all 16 subcommands and including 13 that run in an empty
directory — was executed twice over identical fixture copies: once normally,
and once with `env -i` and a `PATH` containing nothing but the binary and
`git`, with no Python of any version reachable and `HOME` pointed at an empty
directory. **Exit code, stdout and stderr were identical in all 227 cases.**
The binary was then moved to a different directory and re-run, confirming
nothing is read relative to its own location (R-6.7).

That establishes the binary reads nothing from the environment it should not.
It does not establish R-6.4, because the artifact was not downloaded and the
machine was not clean. Only the procedure above does that.

---

## Upgrading an existing install

```bash
make install            # or: cp the new artifact over the old one
```

`make install` writes to a sibling path and renames over the target. The
rename is atomic — there is no window in which a half-written binary sits on
your `PATH` — and it unlinks the old inode rather than truncating it, which is
what an in-place `cp` cannot do on Linux while the old binary is still running
(`ETXTBSY`). macOS permits the in-place write; Linux is the platform this
protects.

Upgrading one binary to the next is a file copy. There is no migration step,
no state outside the workspace, and nothing to un-install first. On macOS, a
**downloaded** replacement carries a fresh quarantine attribute even though the
old one ran fine — see [the workaround](../tutorials/01-first-day-with-company-os.md#macos-the-first-run-will-be-blocked).

Confirm what you are running:

```bash
company-os --version
# company-os 1.4.0 (commit a1b2c3d, go1.25.7, darwin/arm64)
```

That sentence holds for binary-to-binary upgrades. Coming from the **Python
kit**, there is exactly one migration, and it is below.

### Migrating off the Python kit

The old `install.sh` did not install a file — it installed a *tree*. It copied
`bin/`, `templates/`, `skills/`, `schemas/`, `docs/`, `vendor/` and `README.md`
(79 files, ~1.2 MB) into a kit root, then wrote a small bash launcher onto your
`PATH` that `exec`'d the Python entrypoint inside that tree:

```
$COMPANY_OS_PREFIX/bin/company-os            <- generated bash launcher
$COMPANY_OS_PREFIX/share/company-os/         <- kit root: bin, templates, vendor, …
```

`$COMPANY_OS_PREFIX` defaulted to `~/.local`. Nothing reads it any more.

**The kit is self-contained, and that is the problem.** Because everything was
*copied*, deleting the Python implementation from the repository does not
disturb an existing install at all — the launcher keeps working, indefinitely,
running a Python CLI that no longer exists upstream. It will not error. It will
not warn. If `$COMPANY_OS_PREFIX/bin` sits ahead of the new binary on your
`PATH`, the stale launcher silently **shadows** the binary you just installed.
Every real subcommand — `validate`, `graph build`, `governance resolve` — then
runs the old implementation and succeeds, so nothing looks wrong. The Python
CLI had no `--version` flag of its own, so the one command that would expose
the substitution answers with an argparse usage banner and exit 2 instead of a
version string. You can be a year behind and never see a symptom.

So the migration is not optional cleanup. Do it in this order:

```bash
# 1. See every company-os on your PATH, in resolution order.
type -a company-os          # or: which -a company-os

# 2. Remove the generated launcher.
rm -f ~/.local/bin/company-os          # or $COMPANY_OS_PREFIX/bin/company-os

# 3. Remove the kit root.
rm -rf ~/.local/share/company-os       # or $COMPANY_OS_PREFIX/share/company-os

# 4. Install the binary (download the artifact, or from source):
make install                           # PREFIX=/some/where to relocate
```

Order matters: remove the launcher *before* the kit root. A launcher whose kit
root has been deleted is the one state that fails outright, with a Python-level
`No such file or directory` and exit 2 — it is not harmful, but it is a
confusing thing to leave on someone's `PATH` between steps.

**Verify the migration took:**

```bash
type -a company-os
# exactly one path, and it is where you installed the binary

company-os --version
# company-os 1.4.0 (commit a1b2c3d, go1.25.7, darwin/arm64)
```

The version line is the real check. A usage banner instead of a version means
you are still on the Python launcher and step 2 missed — look again at what
`type -a` printed and delete the entry that is not your binary.

> **If you install with `make install` and never changed `PREFIX`, step 2 is
> already handled**: the binary lands on `$PREFIX/bin/company-os`, the exact
> path the launcher occupied, and overwrites it. The kit root in step 3 is
> orphaned either way and still has to go.

There is no automated migration and there cannot be one. The launcher is a
generated file sitting on your machine; nothing shipped in a later release can
reach back and modify or remove it. This procedure is the whole mechanism.

---

## Version skew across a shared workspace

A federated workspace is shared, so two people on different builds is a normal
condition rather than an edge case. The position:

> **Version skew is supported within a workspace-format major version.** Any
> build reads and writes a workspace last written by any other build of the
> same format version. There is no version negotiation, no lockout, and no
> warning, because there is nothing to negotiate: the binary version is not
> recorded in any artifact.

What that rests on, and what it costs:

- **No workspace artifact records which binary wrote it.** `BuildInfo()` feeds
  exactly two things — the `--version` line and the `build` object in `--json`
  output. It is never persisted. Artifacts carry their own `schemaVersion`
  (`"1.0"` for descriptors and governance files, `version: 1` for
  `workspace.lock.yaml`), and *that* is the compatibility boundary. It moves
  when the format moves, not when the binary does.
- **Generated output is a pure function of workspace state.** Verified: two
  builds with different stamped versions, run over the same fixture, produce
  byte-identical trees; re-running `graph build` and `governance resolve` under
  a different build than wrote them is a no-op diff. So a shared workspace does
  not churn, and CI's regenerate-and-diff gate does not fail because a
  teammate is a version behind.
- **The cost, stated plainly: skew is invisible.** Because nothing is recorded,
  nothing can warn you. If a future release *does* change the format, an older
  binary will not announce that it is too old — it will read what it can and
  produce output the newer one does not expect. The mitigation is not in the
  tool, it is in the release: **a workspace-format change is a `schemaVersion`
  bump plus a validator gate, and must never ship as a silent behavior change
  in a generator.** Any such change belongs in a release note with a stated
  minimum version.
- **The lock file is the one place skew is already caught.** Gate `[8/8]` fails
  on a slice-set change made without a re-sync regardless of who ran `sync` or
  with what build, because it compares recorded hashes rather than versions.

Practical guidance: pin one version per repository in CI, put it in the release
notes, and upgrade the team together. Nothing breaks if you do not — but
nothing tells you either.

---

## Break glass: recovering the Python reference implementation

This section is for maintainers, not adopters. Nothing here is part of normal
operation.

The Go binary was built as a port, and its correctness was established
differentially: `examples/differential.py` ran both implementations over the
same fixtures and compared their output byte for byte. The Python CLI was the
oracle. When it was deleted, the ability to *generate* an expected output for a
case nobody thought to write a golden for went with it.

That is the whole reason a tag exists. A defect found later takes the form
"the binary produces X here, and I believe it should produce Y" — and with the
oracle gone there is no way to settle the argument except by reading code. The
tag restores the machine that answers the question.

### The tag

| | |
|---|---|
| **Name** | `python-cli-final` |
| **Points at** | the last commit in which `company-os-starter/bin/company-os` exists — the direct parent of the commit that deletes it |
| **Kind** | annotated (`git tag -a`), so the reason survives in the object itself |

The name deliberately does not start with `v`. Release tags for the binary are
`v<major>.<minor>.<patch>`; this is not a release and must never sort or glob
in with them.

### Recovering it

The whole point of the tag is that recovery is a checkout, not an archaeology
expedition:

```bash
git checkout python-cli-final -- \
  company-os-starter/bin/company-os \
  company-os-starter/vendor \
  company-os-starter/templates
```

**Recover all three paths, not just the first.** `bin/company-os` alone does
not run usefully:

- `vendor/` holds the pure-Python PyYAML the CLI imports. Without it you need
  `pip install pyyaml` in whatever interpreter you use. The vendored copy is
  the reason the original kit needed no network and no pip, and recovering it
  keeps that property.
- `templates/` is resolved at import time as `Path(__file__).resolve().parent.parent / "templates"` — a path *relative to the script*. Recover the script
  without its sibling `templates/` and the CLI imports fine, `validate` works,
  and every scaffolding command (`init`, `add`, `discover new`, `prd new`)
  fails on a missing file. This is the failure mode most likely to waste an
  afternoon, because the CLI looks healthy right up until it doesn't.

Then run it directly — do not install it:

```bash
PYTHONPATH=company-os-starter/vendor \
  python3 company-os-starter/bin/company-os --root examples/workspace validate
```

Nothing needs to go on your `PATH`, and nothing should. `install.sh` at the
tag still works, but installing a second `company-os` alongside the binary
recreates exactly the shadowing hazard described under
[Migrating off the Python kit](#migrating-off-the-python-kit). Invoke the
recovered script by path, use it, delete it.

To recover the differential harness along with the oracle — which is what you
actually want when generating a golden — take the harness from the same tag:

```bash
git checkout python-cli-final -- examples/differential.py examples/declared-divergences.txt
```

`declared-divergences.txt` is the ledger of intentional, reviewed differences.
Without it every declared divergence reads as a failure and the harness output
is noise.

### What the recovered CLI can and cannot do

**It can** validate a workspace, resolve governance, run the product lifecycle
commands, scaffold from templates, and — the reason it is kept — act as the
reference side of `examples/differential.py` to produce an expected output for
a case that has no golden.

**It cannot** stand in for the shipped binary, and must not be offered to
anyone as a fallback:

- **No `--json`.** The structured envelope does not exist in it. Any agent,
  script, or skill written against `--json` fails immediately.
- **No exit-code contract.** It exits `0` or `1`; the differentiated codes the
  binary guarantees are absent, so CI that branches on an exit code will
  misread it.
- **No `--version`.** It cannot identify itself. `--version` yields an argparse
  usage banner and exit 2.
- **No TUI.**
- **It is frozen at the cutover.** It knows nothing of any behavior added to
  the binary afterwards. As the binary moves on, the oracle's answer to "what
  should this produce?" is only authoritative for behavior that existed at the
  tag — which is precisely the behavior parity was proven over, and precisely
  nothing else. Treat an answer from it as evidence about the port, not as a
  specification for current behavior.

### Creating the tag

**The tag does not exist yet, and this is the correct state.** It cannot be
created before the cutover, because the commit it must point at does not exist
yet: at the time of writing, the entire Go port is uncommitted working tree.
Tagging any commit that exists today would point the recovery path at a
snapshot whose `examples/differential.py` predates the port and which contains
no `declared-divergences.txt` at all — a Python CLI with no harness able to
compare it to anything.

The ordering constraint in R-9.2 is *tag before delete*, and it is satisfied by
committing the port, tagging that commit, and deleting in a **separate**
commit. Collapsing the port and the deletion into one commit destroys the
recovery point: there would be no commit containing both a working oracle and
the Go implementation it was proven against.

A human runs this, once, at cutover:

```bash
# 1. Commit the port. bin/company-os must still be present in this commit.
git add -A
git commit -m "feat(cli): port company-os to Go, at parity with the Python CLI"

# 2. Confirm this commit really is a valid recovery point.
git ls-tree --name-only HEAD company-os-starter/bin/company-os   # must print the path
git ls-tree --name-only HEAD examples/declared-divergences.txt   # must print the path

# 3. Tag it. Annotated, so the reason is in the object.
git tag -a python-cli-final -m "Final commit containing the Python reference implementation.

Recovery:
  git checkout python-cli-final -- \\
    company-os-starter/bin/company-os \\
    company-os-starter/vendor \\
    company-os-starter/templates

See docs/user-guide/how-to/release-and-upgrade.md, 'Break glass'."

# 4. Push the tag BEFORE the deletion lands anywhere shared.
git push origin python-cli-final

# 5. Only now delete, as its own commit (task 6.5).
```

Verify the tag before deleting anything, from a throwaway directory so the
check cannot be satisfied by files already in your working tree:

```bash
git archive python-cli-final \
  company-os-starter/bin/company-os \
  company-os-starter/vendor \
  company-os-starter/templates | tar -x -C "$(mktemp -d)"
```

If that extracts and the script runs, the recovery path is real. If step 4 is
skipped, the tag exists on exactly one laptop and the recovery path is a
rumour.
