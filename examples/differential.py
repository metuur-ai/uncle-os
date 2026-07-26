#!/usr/bin/env python3
"""Cross-implementation differential harness (task 0.3, R-7.1 / R-7.2 / R-7.3a).

Runs an identical corpus of CLI invocations against TWO binaries — a reference
and a candidate — and diffs, per invocation:

  * stdout   (normalized)
  * stderr   (normalized)
  * exit code
  * the resulting workspace filesystem tree (path + mode + normalized content)

Why this exists: `examples/acceptance.sh` byte-freezes `validate` only, and a
Go test suite written against the Go implementation cannot detect a Go/Python
divergence by construction. This harness is the only behavioral oracle for
`discover`, `deviation`, `exception`, `check`, `governance`, `today`, `graph`,
`ids`, `skills`, and `scratchpad`, which `examples/selftest.py` never drove.

Usage:

    examples/differential.py <reference-binary> <candidate-binary>
    examples/differential.py <cli> <cli>              # self-check: must be clean
    examples/differential.py <a> <b> --only governance --only prd
    examples/differential.py <a> <b> --list
    examples/differential.py <a> <b> -v               # print every PASS line

Exit status: 0 when every invocation is PASS, DECLARED, PARTIAL, or SKIP; 1 on any
DIVERGE or on any STALE declaration.

DECLARED means the divergence is listed in `examples/declared-divergences.txt`, the
registry required by R-7.1a. Every entry is keyed by (invocation id, stream) and
must cite the R-0.7a clause, requirement id, or `.devlocal/go-port/exit-code-map.md`
line that sanctions it; the harness refuses to load an entry with no citation. Every
DECLARED entry prints its citation, so the waiver list is auditable from the report
alone. A declaration whose invocation ran and no longer diverges on that stream is
reported STALE and fails the run — a stale waiver is how a fixed bug silently
un-fixes itself later.

PARTIAL means one stream of that invocation is non-deterministic even under a
SINGLE implementation (today: exactly one — git relays a subprocess's stderr on
its own pipe, so line ORDER varies run to run). The stream is excluded and the
exclusion is printed with its reason in every report. Every PARTIAL is a hole in
the oracle; the summary names them rather than burying them in the PASS count.

PARITY IS NOT EXIT 0. A fixture that fails identically under both binaries is
passing evidence. `examples/banking/bank/workspaces/team-fraud-detection` is
deliberately in the corpus even though it validates non-zero.

Normalization is deliberately minimal — over-normalizing is how a harness
reports false parity. Exactly two substitutions are applied, both to output
streams AND to file contents before hashing:

  1. the per-run temporary workspace path (and its realpath) -> <WS>
  2. a generated UTC timestamp `YYYY-MM-DDTHH:MM:SSZ`        -> <TS>
     (`NOW`, bin/company-os:32; surfaces as `generatedAt` in
     teams/<t>/generated/effective-governance.yaml and in `today` stdout)

`TODAY`-derived dates (bin/company-os:31) are NOT normalized: both binaries run
on the same calendar day, so they are deterministic. Nothing else is touched.

The tree snapshot skips any path containing a `.git` component — git's own
object store inside `.company-os/federation-cache` carries reflog timestamps and
pack nondeterminism that no CLI produced. The materialized slices themselves are
fully compared, including their read-only mode bits.
"""

import argparse
import difflib
import hashlib
import os
import re
import shutil
import stat
import subprocess
import sys
import tempfile
import time
from pathlib import Path

HERE = Path(__file__).resolve().parent

# ------------------------------------------------------------------ fixtures

# Committed fixtures. Never mutated: every invocation runs against a fresh copy.
FIXTURES = {
    "workspace": HERE / "workspace",
    "standalone-team": HERE / "standalone-team",
    "federated": HERE / "federated",
    "banking-small": HERE / "banking" / "small-company",
    "banking-rails": HERE / "banking" / "bank" / "workspaces" / "team-payments-rails",
    "banking-fraud": HERE / "banking" / "bank" / "workspaces" / "team-fraud-detection",
    # Failure-path fixtures captured under task 0.2 / R-0.9. They exist to fail;
    # a fixture that fails identically under both binaries is passing evidence.
    "failing-workspace": HERE / "failing-workspace",
    "failing-federated": HERE / "failing-federated",
    "failing-federated-nolock": HERE / "failing-federated-nolock",
}

# Committed fixtures that are optional (added by a sibling task): a missing one
# SKIPs loudly rather than crashing the whole run.
OPTIONAL_FIXTURES = {"failing-workspace", "failing-federated", "failing-federated-nolock"}

# Fixtures synthesized at startup (bad manifests, empty dirs, the git source
# repo). Populated by build_synthetic_fixtures(); same copy-per-invocation rule.
SYNTHETIC = {}

TS_RE = re.compile(r"\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z")


def normalizer(ws_path):
    """Return a fn normalizing the two genuinely non-deterministic things."""
    paths = {str(ws_path), os.path.realpath(str(ws_path))}
    subs = sorted(paths, key=len, reverse=True)

    def norm(text):
        for p in subs:
            text = text.replace(p, "<WS>")
        return TS_RE.sub("<TS>", text)

    return norm


# ------------------------------------------------------------------ the corpus


class Inv:
    """One corpus entry: a fixture plus an ordered list of argv steps.

    Multiple steps run in the SAME workspace copy, in order, so lifecycle
    sequences (prd new -> validate -> complete) and state-dependent commands
    (workspace sync -> status -> sync --frozen) are reachable. Every step's
    stdout/stderr/exit is compared; the tree is compared once at the end.
    """

    def __init__(self, cid, fixture, steps, group, note="", unstable=""):
        self.id = cid
        self.fixture = fixture
        self.steps = [list(s) for s in steps]
        self.group = group
        self.note = note
        # Non-empty => one stream of this invocation is not deterministic even
        # under a SINGLE implementation, so comparing it would produce false
        # DIVERGE. The stream is excluded and the invocation is reported PARTIAL,
        # never silently. Format: "<stream>: <reason>". Use sparingly — every use
        # is a hole in the oracle and must justify itself.
        self.unstable = unstable

    @property
    def unstable_stream(self):
        return self.unstable.split(":", 1)[0].strip() if self.unstable else ""


CORPUS = []


def inv(cid, fixture, *steps, group=None, note="", unstable=""):
    CORPUS.append(Inv(cid, fixture, steps, group or cid.split("/")[0], note, unstable))


W, S, F = "workspace", "standalone-team", "federated"
BS, BR, BF = "banking-small", "banking-rails", "banking-fraud"
XW, XF, XN = "failing-workspace", "failing-federated", "failing-federated-nolock"
ALL_WS = [W, S, F, BS, BR, BF]
FAILING_WS = [XW, XF, XN]

# --- validate (byte-frozen elsewhere; here for exit-code + tree parity) ------
for fx in ALL_WS:
    inv(f"validate/{fx}", fx, ["validate"])
inv("validate/not-a-root", "empty", ["validate"], note="exit 3 both sides")
inv("validate/after-graph-build", W, ["graph", "build"], ["validate"])
# R-0.9 failure-path fixtures: every gate's [FAIL]/[warn] rendering, under both.
for fx in FAILING_WS:
    inv(f"validate/{fx}", fx, ["validate"])
    inv(f"graph/build-{fx}", fx, ["graph", "build"], group="graph")
    inv(f"today/default-{fx}", fx, ["today"], group="today")
    inv(f"ids/list-{fx}", fx, ["ids", "list"], group="ids")
    inv(f"skills/list-{fx}", fx, ["skills", "list"], group="skills")
    inv(f"workspace/status-{fx}", fx, ["workspace", "status"], group="workspace")
inv("governance/resolve-failing-workspace", XW, ["governance", "resolve", "--team", "ghost"])
inv("governance/explain-failing-workspace", XW, ["governance", "explain", "orphan"])
inv("check/ready-failing-workspace", XW,
    ["check", "ready", "--team", "ghost", "--components", "orphan"])
inv("workspace/frozen-failing-federated", XF, ["workspace", "sync", "--frozen"], group="workspace")
inv("workspace/frozen-failing-nolock", XN, ["workspace", "sync", "--frozen"], group="workspace")

# --- governance (ZERO prior coverage) ---------------------------------------
inv("governance/resolve-workspace", W, ["governance", "resolve", "--team", "customer-engagement"])
inv("governance/resolve-federated", F, ["governance", "resolve", "--team", "customer-engagement"])
inv("governance/resolve-standalone", S, ["governance", "resolve", "--team", "solo"])
inv("governance/resolve-banking", BS, ["governance", "resolve", "--team", "core"])
inv("governance/resolve-fraud", BF, ["governance", "resolve", "--team", "fraud-detection"])
inv("governance/resolve-rails", BR, ["governance", "resolve", "--team", "payments-rails"])
inv("governance/resolve-twice", W,
    ["governance", "resolve", "--team", "customer-engagement"],
    ["governance", "resolve", "--team", "customer-engagement"],
    note="idempotence of generated/effective-governance.yaml")
inv("governance/resolve-unknown-team", W, ["governance", "resolve", "--team", "nope"])
inv("governance/resolve-no-team", W, ["governance", "resolve"])
inv("governance/explain-known", W, ["governance", "explain", "customer-notification-service"])
inv("governance/explain-banking", BS, ["governance", "explain", "banking-app"])
inv("governance/explain-unknown-with-suggestions", W,
    ["governance", "explain", "customer-notification-servic"],
    note="difflib suggestion path (GPF-R-2.3)")
inv("governance/explain-unknown-no-suggestions", W, ["governance", "explain", "zzzzzzzz"])
inv("governance/explain-standalone", S, ["governance", "explain", "anything"])
inv("governance/explain-no-component", W, ["governance", "explain"])
inv("governance/explain-after-resolve", S,
    ["governance", "resolve", "--team", "solo"],
    ["governance", "explain", "anything"])

# --- discover (ZERO prior coverage) -----------------------------------------
inv("discover/new", W, ["discover", "new", "--team", "customer-engagement", "Quiet hours v2"])
inv("discover/new-twice-conflict", W,
    ["discover", "new", "--team", "customer-engagement", "Quiet hours v2"],
    ["discover", "new", "--team", "customer-engagement", "Quiet hours v2"],
    note="second must refuse (exit 8)")
inv("discover/new-then-validate", W,
    ["discover", "new", "--team", "customer-engagement", "Quiet hours v2"],
    ["discover", "validate", "--team", "customer-engagement", "2026-quiet-hours-v2"],
    note="fresh brief is a stub -> validate should fail")
inv("discover/new-standalone", S, ["discover", "new", "--team", "solo", "Solo idea"])
inv("discover/new-banking", BS, ["discover", "new", "--team", "core", "Statements v2"])
inv("discover/new-unknown-team", W, ["discover", "new", "--team", "nope", "X"])
inv("discover/new-slugify", W,
    ["discover", "new", "--team", "customer-engagement", "  Weird // Title!! 2026  "])
inv("discover/new-no-title", W, ["discover", "new", "--team", "customer-engagement"])
inv("discover/validate-passing", W,
    ["discover", "validate", "--team", "customer-engagement", "2026-per-channel-quiet-hours"])
inv("discover/validate-federated", F,
    ["discover", "validate", "--team", "customer-engagement", "2026-per-channel-quiet-hours"])
inv("discover/validate-banking", BS,
    ["discover", "validate", "--team", "core", "2026-instant-statements"])
inv("discover/validate-fraud", BF,
    ["discover", "validate", "--team", "fraud-detection", "2026-alert-triage-queues"])
inv("discover/validate-missing", W,
    ["discover", "validate", "--team", "customer-engagement", "2026-nope"])
inv("discover/validate-unknown-team", W, ["discover", "validate", "--team", "nope", "x"])
inv("discover/missing-team-flag", W, ["discover", "new", "Title"], note="argparse usage error")
inv("discover/bad-action", W, ["discover", "frobnicate", "--team", "customer-engagement", "x"])

# --- prd (partial selftest coverage; lifecycle is invariant #4) -------------
PRD_NEW_W = ["prd", "new", "--team", "customer-engagement", "--platform", "communications",
             "--components", "customer-notification-service",
             "--from-discovery", "2026-per-channel-quiet-hours"]
inv("prd/new-from-discovery", W, PRD_NEW_W)
inv("prd/new-then-validate", W, PRD_NEW_W,
    ["prd", "validate", "--platform", "communications", "2026-per-channel-quiet-hours"])
inv("prd/full-lifecycle", W, PRD_NEW_W,
    ["prd", "validate", "--platform", "communications", "2026-per-channel-quiet-hours"],
    ["prd", "complete", "--platform", "communications", "2026-per-channel-quiet-hours"],
    note="done-check should refuse: stale reality/ (invariant #4)")
inv("prd/full-lifecycle-force", W, PRD_NEW_W,
    ["prd", "complete", "--platform", "communications", "2026-per-channel-quiet-hours", "--force"],
    note="archive + outcome.md + log.md written")
inv("prd/new-twice-conflict", W, PRD_NEW_W, PRD_NEW_W)
inv("prd/new-with-title", W,
    ["prd", "new", "--team", "customer-engagement", "--platform", "communications",
     "--components", "customer-notification-service", "--title", "Ad hoc change"])
inv("prd/new-no-title-no-discovery", W,
    ["prd", "new", "--team", "customer-engagement", "--platform", "communications",
     "--components", "customer-notification-service"])
inv("prd/new-missing-discovery", W,
    ["prd", "new", "--team", "customer-engagement", "--platform", "communications",
     "--components", "customer-notification-service", "--from-discovery", "2026-nope"])
inv("prd/new-draft-discovery", W,
    ["discover", "new", "--team", "customer-engagement", "Draft idea"],
    ["prd", "new", "--team", "customer-engagement", "--platform", "communications",
     "--components", "customer-notification-service", "--from-discovery", "2026-draft-idea"],
    note="brief is draft, not validated -> exit 5")
inv("prd/new-unknown-platform", W,
    ["prd", "new", "--team", "customer-engagement", "--platform", "nope",
     "--components", "x", "--title", "T"])
inv("prd/new-missing-platform-flag", W,
    ["prd", "new", "--team", "customer-engagement", "--title", "T"])
inv("prd/validate-missing", W, ["prd", "validate", "--platform", "communications", "2026-nope"])
inv("prd/complete-missing", W, ["prd", "complete", "--platform", "communications", "2026-nope"])
inv("prd/validate-banking-active", BS,
    ["prd", "validate", "--platform", "product", "2026-instant-statements"])
inv("prd/complete-banking-active", BS,
    ["prd", "complete", "--platform", "product", "2026-instant-statements"])
inv("prd/complete-banking-force", BS,
    ["prd", "complete", "--platform", "product", "2026-instant-statements", "--force"])
inv("prd/new-multi-component", W,
    ["prd", "new", "--team", "customer-engagement", "--platform", "communications",
     "--components", "customer-notification-service,ghost-component",
     "--title", "Multi component"])

# --- check (ZERO prior coverage) --------------------------------------------
inv("check/ready-workspace", W,
    ["check", "ready", "--team", "customer-engagement",
     "--components", "customer-notification-service"])
inv("check/done-workspace", W,
    ["check", "done", "--team", "customer-engagement",
     "--components", "customer-notification-service"])
inv("check/ready-federated", F,
    ["check", "ready", "--team", "customer-engagement",
     "--components", "customer-notification-service"])
inv("check/ready-banking", BS, ["check", "ready", "--team", "core", "--components", "banking-app"])
inv("check/done-banking", BS, ["check", "done", "--team", "core", "--components", "banking-app"])
inv("check/ready-standalone", S, ["check", "ready", "--team", "solo", "--components", "none"])
inv("check/done-standalone", S, ["check", "done", "--team", "solo", "--components", "none"])
inv("check/ready-multi", W,
    ["check", "ready", "--team", "customer-engagement",
     "--components", "customer-notification-service,ghost,another"])
inv("check/ready-unknown-team", W, ["check", "ready", "--team", "nope", "--components", "x"])
inv("check/ready-unknown-component", W,
    ["check", "ready", "--team", "customer-engagement", "--components", "does-not-exist"])
inv("check/missing-components-flag", W, ["check", "ready", "--team", "customer-engagement"])
inv("check/bad-kind", W,
    ["check", "sideways", "--team", "customer-engagement", "--components", "x"])
inv("check/ready-fraud", BF,
    ["check", "ready", "--team", "fraud-detection", "--components", "fraud-scoring-engine"])

# --- deviation (ZERO prior coverage) ----------------------------------------
inv("deviation/declare-default-rule", W,
    ["deviation", "declare", "platform-standard://communications/prd-structure",
     "--team", "customer-engagement"])
inv("deviation/declare-then-resolve", W,
    ["deviation", "declare", "platform-standard://communications/prd-structure",
     "--team", "customer-engagement"],
    ["governance", "resolve", "--team", "customer-engagement"],
    note="deviation must appear in deviationsApplied")
inv("deviation/declare-mandatory-then-resolve", W,
    ["deviation", "declare", "platform-standard://communications/delivery-reliability",
     "--team", "customer-engagement"],
    ["governance", "resolve", "--team", "customer-engagement"],
    note="mandatory rule -> deviationRejected (invariant #1)")
inv("deviation/declare-with-rationale", W,
    ["deviation", "declare", "platform-standard://communications/prd-structure",
     "--team", "customer-engagement", "--rationale", "we ship a different shape"])
inv("deviation/declare-twice", W,
    ["deviation", "declare", "platform-standard://communications/prd-structure",
     "--team", "customer-engagement"],
    ["deviation", "declare", "platform-standard://communications/prd-structure",
     "--team", "customer-engagement"],
    ["governance", "resolve", "--team", "customer-engagement"])
inv("deviation/declare-standalone", S,
    ["deviation", "declare", "company-control://change-log", "--team", "solo"])
inv("deviation/declare-banking", BS,
    ["deviation", "declare", "company-control://change-log", "--team", "core"],
    ["governance", "resolve", "--team", "core"])
inv("deviation/declare-unknown-team", W,
    ["deviation", "declare", "x", "--team", "nope"])
inv("deviation/declare-then-validate", W,
    ["deviation", "declare", "platform-standard://communications/prd-structure",
     "--team", "customer-engagement"],
    ["governance", "resolve", "--team", "customer-engagement"],
    ["validate"],
    note="reviewDate is TODAY+180 -> gate [2/8] must stay green")
inv("deviation/bad-action", W, ["deviation", "revoke", "x", "--team", "customer-engagement"])

# --- exception (ZERO prior coverage) ----------------------------------------
inv("exception/request-future", W,
    ["exception", "request", "platform-standard://communications/delivery-reliability",
     "--team", "customer-engagement", "--component", "customer-notification-service",
     "--expires", "2035-01-01"])
inv("exception/request-then-validate", W,
    ["exception", "request", "platform-standard://communications/delivery-reliability",
     "--team", "customer-engagement", "--component", "customer-notification-service",
     "--expires", "2035-01-01"],
    ["validate"])
inv("exception/request-expired-then-validate", W,
    ["exception", "request", "platform-standard://communications/delivery-reliability",
     "--team", "customer-engagement", "--component", "customer-notification-service",
     "--expires", "2020-01-01"],
    ["validate"],
    note="past expiry -> gate [2/8] must FAIL identically on both sides")
inv("exception/request-with-reason", W,
    ["exception", "request", "platform-standard://communications/message-schema",
     "--team", "customer-engagement", "--component", "customer-notification-service",
     "--expires", "2035-06-30", "--reason", "legacy consumer"])
inv("exception/request-standalone", S,
    ["exception", "request", "company-control://security-service-baseline",
     "--team", "solo", "--component", "none", "--expires", "2035-01-01"])
inv("exception/request-banking", BS,
    ["exception", "request", "platform-standard://product/release-safety",
     "--team", "core", "--component", "banking-app", "--expires", "2035-01-01"],
    ["validate"])
inv("exception/request-unknown-team", W,
    ["exception", "request", "x", "--team", "nope", "--component", "c",
     "--expires", "2035-01-01"])
inv("exception/missing-expires", W,
    ["exception", "request", "x", "--team", "customer-engagement", "--component", "c"])
inv("exception/missing-component", W,
    ["exception", "request", "x", "--team", "customer-engagement", "--expires", "2035-01-01"])
inv("exception/garbage-expires", W,
    ["exception", "request", "platform-standard://communications/message-schema",
     "--team", "customer-engagement", "--component", "customer-notification-service",
     "--expires", "not-a-date"],
    ["validate"])

# --- today (ZERO prior coverage) --------------------------------------------
ROLES = ["developer", "team-lead", "product-owner", "architect",
         "vp-engineering", "director-of-product"]
for fx in ALL_WS:
    inv(f"today/default-{fx}", fx, ["today"])
for role in ROLES:
    inv(f"today/role-{role}", W, ["today", "--role", role])
inv("today/po-banking", BS, ["today", "--role", "product-owner"],
    note="active PRD + archived outcome review")
inv("today/dev-banking", BS, ["today", "--role", "developer"])
inv("today/dev-standalone-no-generated", S, ["today", "--role", "developer"],
    note="warn: no effective-governance.yaml")
inv("today/after-resolve", S,
    ["governance", "resolve", "--team", "solo"], ["today", "--role", "developer"])
inv("today/bad-role", W, ["today", "--role", "wizard"])
inv("today/not-a-root", "empty", ["today"])

# --- graph (ZERO prior coverage) --------------------------------------------
for fx in ALL_WS:
    inv(f"graph/build-{fx}", fx, ["graph", "build"])
inv("graph/build-twice", W, ["graph", "build"], ["graph", "build"],
    note="idempotence (R-0.6)")
inv("graph/build-then-validate", BS, ["graph", "build"], ["validate"])
inv("graph/build-after-discover", W,
    ["discover", "new", "--team", "customer-engagement", "Tagged brief"],
    ["graph", "build"],
    note="derived tags: for a brand-new artifact")
inv("graph/build-not-a-root", "empty", ["graph", "build"])
inv("graph/bad-action", W, ["graph", "rebuild"])

# --- ids (ZERO prior coverage) ----------------------------------------------
for fx in ALL_WS:
    inv(f"ids/list-{fx}", fx, ["ids", "list"])
inv("ids/filter-team", W, ["ids", "list", "--team", "customer-engagement"])
inv("ids/filter-platform", W, ["ids", "list", "--platform", "communications"])
inv("ids/filter-prefix-component", W, ["ids", "list", "--prefix", "component://"])
inv("ids/filter-prefix-team", W, ["ids", "list", "--prefix", "team://"])
inv("ids/filter-prefix-nomatch", W, ["ids", "list", "--prefix", "zzzz"])
inv("ids/filter-team-nomatch", W, ["ids", "list", "--team", "nope"])
inv("ids/filter-combined", W,
    ["ids", "list", "--team", "customer-engagement", "--platform", "communications"])
for role in ROLES:
    inv(f"ids/role-{role}", W, ["ids", "list", "--role", role])
inv("ids/role-unknown", W, ["ids", "list", "--role", "wizard"])
inv("ids/list-banking-filtered", BS, ["ids", "list", "--platform", "product"])
inv("ids/bad-action", W, ["ids", "show"])
inv("ids/not-a-root", "empty", ["ids", "list"])

# --- skills (ZERO prior coverage) -------------------------------------------
for fx in ALL_WS:
    inv(f"skills/list-{fx}", fx, ["skills", "list"])
inv("skills/list-after-scratchpad", W,
    ["scratchpad", "init", "--repo", "teams/customer-engagement"],
    ["skills", "list"],
    note="personal-rules layer becomes discoverable")
inv("skills/bad-action", W, ["skills", "show"])
inv("skills/not-a-root", "empty", ["skills", "list"])

# --- scratchpad (ZERO prior coverage; exempt from require_root) -------------
inv("scratchpad/init-cwd", W, ["scratchpad", "init"])
inv("scratchpad/init-repo-dot", W, ["scratchpad", "init", "--repo", "."])
inv("scratchpad/init-nested-new-dir", W, ["scratchpad", "init", "--repo", "sub/dir"])
inv("scratchpad/init-twice", W, ["scratchpad", "init"], ["scratchpad", "init"],
    note=".gitignore must not be appended twice")
inv("scratchpad/init-standalone", S, ["scratchpad", "init"])
inv("scratchpad/init-outside-workspace", "empty", ["scratchpad", "init"],
    note="exempt from require_root")
inv("scratchpad/init-team-dir", W, ["scratchpad", "init", "--repo", "teams/customer-engagement"])
inv("scratchpad/bad-action", W, ["scratchpad", "reset"])

# --- init -------------------------------------------------------------------
inv("init/full-flags", "empty",
    ["init", "--company", "Acme", "--team", "core", "--platform", "platform-1"])
inv("init/then-validate", "empty",
    ["init", "--company", "Acme", "--team", "core", "--platform", "platform-1"],
    ["validate"])
inv("init/slugify", "empty",
    ["init", "--company", "Acme Inc.", "--team", "Core Team!", "--platform", "My Platform!!"])
inv("init/non-interactive-no-flags", "empty", ["init"], note="no TTY -> exit 7")
inv("init/non-interactive-partial-flags", "empty", ["init", "--company", "Acme"])
inv("init/refuse-reinit", W,
    ["init", "--company", "Acme", "--team", "core", "--platform", "p"],
    note="already a root -> exit 8, mutating nothing")
inv("init/refuse-reinit-manifest-only", BR,
    ["init", "--company", "Acme", "--team", "core", "--platform", "p"],
    note="workspace.yaml alone marks a root (GPF-R-6.2)")
inv("init/then-add-then-validate", "empty",
    ["init", "--company", "Acme", "--team", "core", "--platform", "platform-1"],
    ["add", "component", "widget", "--platform", "platform-1"],
    ["validate"])

# --- add --------------------------------------------------------------------
inv("add/platform", W, ["add", "platform", "newplat"])
inv("add/team", W, ["add", "team", "newteam"])
inv("add/component", W, ["add", "component", "newcomp", "--platform", "communications"])
inv("add/component-no-platform", W, ["add", "component", "newcomp"])
inv("add/component-unknown-platform", W, ["add", "component", "newcomp", "--platform", "nope"])
inv("add/platform-slugify", W, ["add", "platform", "Weird Name!!"])
inv("add/platform-standalone", S, ["add", "platform", "newplat"],
    note="platforms/ does not exist yet")
inv("add/team-standalone", S, ["add", "team", "second"])
inv("add/platform-then-component-then-validate", W,
    ["add", "platform", "newplat"],
    ["add", "component", "newcomp", "--platform", "newplat"],
    ["validate"])
inv("add/duplicate-platform", W, ["add", "platform", "communications"],
    note="existing platform id")
inv("add/bad-kind", W, ["add", "widget", "x"])
inv("add/banking", BS, ["add", "team", "second"], ["validate"])

# --- reality ----------------------------------------------------------------
inv("reality/new-for-new-component", W,
    ["add", "component", "newcomp", "--platform", "communications"],
    ["reality", "new", "--platform", "communications", "newcomp"])
inv("reality/new-conflict", W,
    ["reality", "new", "--platform", "communications", "customer-notification-service"],
    note="reality doc exists -> exit 8")
inv("reality/new-unknown-platform", W, ["reality", "new", "--platform", "nope", "x"])
inv("reality/new-no-platform-flag", W, ["reality", "new", "x"])
inv("reality/new-orphan-component", W,
    ["reality", "new", "--platform", "communications", "never-declared"],
    ["validate"],
    note="reality doc for a component with no descriptor")
inv("reality/bad-action", W, ["reality", "delete", "x", "--platform", "communications"])
inv("reality/new-banking", BS, ["reality", "new", "--platform", "product", "second-app"])

# --- workspace: manifest + monorepo paths (git-free) ------------------------
inv("workspace/sync-no-manifest", W, ["workspace", "sync"])
inv("workspace/status-no-manifest", W, ["workspace", "status"])
inv("workspace/status-no-manifest-standalone", S, ["workspace", "status"])
inv("workspace/status-federated", F, ["workspace", "status"],
    note="committed lock + committed slices -> clean")
inv("workspace/status-rails", BR, ["workspace", "status"], note="never synced")
inv("workspace/status-fraud", BF, ["workspace", "status"], note="never synced")
inv("workspace/frozen-federated-no-cache", F, ["workspace", "sync", "--frozen"],
    note="lock present, cache absent -> exit 6")
inv("workspace/frozen-no-lock", BR, ["workspace", "sync", "--frozen"])
inv("workspace/only-nomatch", F, ["workspace", "sync", "--only", "nope"])
inv("workspace/bad-action", F, ["workspace", "pull"])
inv("workspace/not-a-root", "empty", ["workspace", "status"])

# Bad-manifest fixtures are registered by build_synthetic_fixtures(); one
# `workspace status` invocation each exercises a distinct load_manifest die().
BAD_MANIFESTS = {
    "not-a-mapping": "- just\n- a\n- list\n",
    "no-repos-key": "version: 1\nsomething: else\n",
    "empty-repos": "version: 1\nrepos: []\n",
    "repo-not-mapping": "version: 1\nrepos:\n  - just-a-string\n",
    "missing-url": ("version: 1\nrepos:\n  - name: a\n    pin: {commit: "
                    "'0000000000000000000000000000000000000000'}\n"
                    "    localDirectory: platforms/a\n    paths: [governance/]\n"),
    "missing-pin": ("version: 1\nrepos:\n  - name: a\n    url: file:///nowhere\n"
                    "    localDirectory: platforms/a\n    paths: [governance/]\n"),
    "duplicate-name": ("version: 1\nrepos:\n"
                       "  - name: a\n    url: file:///n\n    pin: {commit: '" + "0" * 40 + "'}\n"
                       "    localDirectory: platforms/a\n    paths: [governance/]\n"
                       "  - name: a\n    url: file:///n\n    pin: {commit: '" + "1" * 40 + "'}\n"
                       "    localDirectory: platforms/b\n    paths: [governance/]\n"),
    "bad-name-chars": ("version: 1\nrepos:\n  - name: 'bad name/slash'\n    url: file:///n\n"
                       "    pin: {commit: '" + "0" * 40 + "'}\n"
                       "    localDirectory: platforms/a\n    paths: [governance/]\n"),
    "renamed-root-key": ("version: 1\nrepos:\n  - name: a\n    url: file:///n\n"
                         "    pin: {commit: '" + "0" * 40 + "'}\n"
                         "    root: platforms/a\n    paths: [governance/]\n"),
    "floating-pin-branch": ("version: 1\nrepos:\n  - name: a\n    url: file:///n\n"
                            "    pin: {branch: main}\n"
                            "    localDirectory: platforms/a\n    paths: [governance/]\n"),
    "pin-both-commit-and-tag": ("version: 1\nrepos:\n  - name: a\n    url: file:///n\n"
                                "    pin: {commit: '" + "0" * 40 + "', tag: v1}\n"
                                "    localDirectory: platforms/a\n    paths: [governance/]\n"),
    "pin-not-a-mapping": ("version: 1\nrepos:\n  - name: a\n    url: file:///n\n"
                          "    pin: main\n"
                          "    localDirectory: platforms/a\n    paths: [governance/]\n"),
    "empty-paths": ("version: 1\nrepos:\n  - name: a\n    url: file:///n\n"
                    "    pin: {commit: '" + "0" * 40 + "'}\n"
                    "    localDirectory: platforms/a\n    paths: []\n"),
    "absolute-localdir": ("version: 1\nrepos:\n  - name: a\n    url: file:///n\n"
                          "    pin: {commit: '" + "0" * 40 + "'}\n"
                          "    localDirectory: /etc/passwd\n    paths: [governance/]\n"),
    "escaping-localdir": ("version: 1\nrepos:\n  - name: a\n    url: file:///n\n"
                          "    pin: {commit: '" + "0" * 40 + "'}\n"
                          "    localDirectory: ../outside\n    paths: [governance/]\n"),
    "non-canonical-localdir": ("version: 1\nrepos:\n  - name: a\n    url: file:///n\n"
                               "    pin: {commit: '" + "0" * 40 + "'}\n"
                               "    localDirectory: elsewhere/a\n    paths: [governance/]\n"),
    "bare-knowledge-root": ("version: 1\nrepos:\n  - name: a\n    url: file:///n\n"
                            "    pin: {commit: '" + "0" * 40 + "'}\n"
                            "    localDirectory: knowledge\n    paths: [docs/]\n"),
    "overlapping-targets": ("version: 1\nrepos:\n"
                            "  - name: a\n    url: file:///n\n    pin: {commit: '" + "0" * 40 + "'}\n"
                            "    localDirectory: platforms/x\n    paths: [governance/]\n"
                            "  - name: b\n    url: file:///n\n    pin: {commit: '" + "1" * 40 + "'}\n"
                            "    localDirectory: platforms/x/inner\n    paths: [governance/]\n"),
    "slices-not-a-list": ("version: 1\nrepos:\n  - name: a\n    url: file:///n\n"
                          "    pin: {commit: '" + "0" * 40 + "'}\n    slices: nope\n"),
    "slices-entry-not-mapping": ("version: 1\nrepos:\n  - name: a\n    url: file:///n\n"
                                 "    pin: {commit: '" + "0" * 40 + "'}\n    slices: [nope]\n"),
    "slices-missing-localdir": ("version: 1\nrepos:\n  - name: a\n    url: file:///n\n"
                                "    pin: {commit: '" + "0" * 40 + "'}\n"
                                "    slices:\n      - paths: [governance/]\n"),
    "slices-uses-root-key": ("version: 1\nrepos:\n  - name: a\n    url: file:///n\n"
                             "    pin: {commit: '" + "0" * 40 + "'}\n"
                             "    slices:\n      - root: platforms/a\n        paths: [g/]\n"),
    "slices-and-localdir": ("version: 1\nrepos:\n  - name: a\n    url: file:///n\n"
                            "    pin: {commit: '" + "0" * 40 + "'}\n"
                            "    localDirectory: platforms/a\n    paths: [governance/]\n"
                            "    slices:\n      - localDirectory: platforms/b\n"
                            "        paths: [governance/]\n"),
    "neither-slices-nor-localdir": ("version: 1\nrepos:\n  - name: a\n    url: file:///n\n"
                                    "    pin: {commit: '" + "0" * 40 + "'}\n"),
    "malformed-yaml": "version: 1\nrepos:\n  - name: [unclosed\n",
}
for _bm in sorted(BAD_MANIFESTS):
    inv(f"workspace/manifest-{_bm}", f"badmanifest-{_bm}", ["workspace", "status"],
        group="workspace")
    inv(f"workspace/manifest-{_bm}-sync", f"badmanifest-{_bm}", ["workspace", "sync"],
        group="workspace")
inv("workspace/manifest-validate-gate", "badmanifest-empty-repos", ["validate"],
    group="workspace", note="a malformed manifest must also break `validate`")

# --- workspace: real sync (git >= 2.27 required; file:// URL, no network) ----
inv("workspace/sync-online", "gitfed", ["workspace", "sync"], group="workspace-git")
inv("workspace/sync-then-status", "gitfed",
    ["workspace", "sync"], ["workspace", "status"], group="workspace-git")
inv("workspace/sync-then-frozen", "gitfed",
    ["workspace", "sync"], ["workspace", "sync", "--frozen"], ["workspace", "status"],
    group="workspace-git", note="--frozen must reproduce slice bytes from the lock")
inv("workspace/sync-twice", "gitfed",
    ["workspace", "sync"], ["workspace", "sync"], group="workspace-git",
    note="re-sync over a read-only nested slice")
inv("workspace/sync-only-known", "gitfed",
    ["workspace", "sync", "--only", "testplat"], group="workspace-git")
inv("workspace/sync-only-unknown", "gitfed",
    ["workspace", "sync", "--only", "nope"], group="workspace-git")
inv("workspace/sync-then-validate", "gitfed",
    ["workspace", "sync"], ["validate"], group="workspace-git")
inv("workspace/sync-then-tamper-then-validate", "gitfed",
    ["workspace", "sync"], ["__tamper__"], ["validate"],
    group="workspace-git", note="gate [8/8] hash-integrity FAIL on a hand-edited slice")
inv("workspace/sync-bad-pin", "gitfed-badpin", ["workspace", "sync"], group="workspace-git",
    note="short commit pin -> exit 4")
inv("workspace/sync-missing-ref", "gitfed-missingref", ["workspace", "sync"],
    group="workspace-git", note="unreachable commit -> git failure, exit 6",
    unstable="stderr: git relays the local upload-pack process's stderr and its "
             "own on separate pipes, so the two `not our ref` lines interleave in "
             "a non-deterministic ORDER between runs of ONE binary. Exit code and "
             "file tree are still compared; the wrapping `error: \\`git ...\\` "
             "failed (exit 128)` line is covered deterministically by "
             "workspace/sync-bad-pin.")

# --- top-level / usage ------------------------------------------------------
inv("usage/no-args", W, [])
inv("usage/unknown-subcommand", W, ["frobnicate"])
inv("usage/help", W, ["--help"])
inv("usage/validate-help", W, ["validate", "--help"])
inv("usage/prd-help", W, ["prd", "--help"])
inv("usage/bad-flag", W, ["validate", "--nope"])
inv("usage/root-flag-nonexistent-dir", W, ["--root", "/nonexistent/xyz", "validate"],
    note="--root is passed through verbatim by the runner")


# ------------------------------------------- declared-divergence registry ---
#
# R-7.1a. The data lives in examples/declared-divergences.txt so the waiver list
# is reviewable in a diff without reading Python; this section is only the
# loader, the matcher, and the validation that makes an uncited waiver
# impossible to write.

REGISTRY_PATH = HERE / "declared-divergences.txt"

STREAMS = ("stdout", "stderr", "exit_code", "file_tree")

# A citation must be SHAPED like an authority, not merely non-empty: an R-0.7a
# sanctioned-difference clause, a requirement id, or a classified die() site in
# the task-0.5 exit-code map. Anything else fails the load.
AUTHORITY_RE = re.compile(
    r"^(?:R-0\.7a\([a-z]\)"
    r"|R-\d+\.\d+[a-z]?"
    r"|\.devlocal/go-port/exit-code-map\.md:\d+)$")

EXPECT_RE = re.compile(r"^(-?\d+)\s*->\s*(-?\d+)$")

# The first line of a stderr stream that is a DIAGNOSTIC rather than usage.
# Everything before it is argparse's wrapped usage block (R-0.7a(i)); the
# diagnostic itself is never waived (R-1.4a).
USAGE_ERR_RE = re.compile(r"^company-os[^\n]*?: error: ")


class Waiver:
    def __init__(self, invocation, stream, authority, reason, expect, waive, step, lineno):
        self.invocation = invocation
        self.stream = stream
        self.authority = authority          # list[str]
        self.reason = reason
        self.expect = expect                # (ref, cand) for exit_code, else None
        self.waive = waive                  # "whole-stream" | "usage-block" | None
        self.step = step                    # int or None (any)
        self.lineno = lineno
        self.matched = 0                    # blocks this waiver was consulted for
        self.resolved = 0                   # blocks it fully waived

    @property
    def key(self):
        return (self.invocation, self.stream, self.step)

    @property
    def cite(self):
        return ", ".join(self.authority)

    def applies(self, inv_id, stream, step):
        if self.invocation != inv_id or self.stream != stream:
            return False
        return self.step is None or self.step == step


class RegistryError(Exception):
    pass


def _parse_records(text):
    """Blank-line-separated `key: value` records; `#` comments; indented
    continuation lines append to the previous value."""
    records, cur, key = [], None, None
    for n, raw in enumerate(text.splitlines(), 1):
        if raw.lstrip().startswith("#"):
            continue
        if not raw.strip():
            if cur:
                records.append(cur)
            cur, key = None, None
            continue
        if raw[:1].isspace() and cur and key:
            cur[key] = (cur[key] + " " + raw.strip()).strip()
            continue
        if ":" not in raw:
            raise RegistryError(f"{REGISTRY_PATH}:{n}: not a `key: value` line: {raw!r}")
        k, v = raw.split(":", 1)
        k, v = k.strip(), v.strip()
        if cur is None:
            cur = {"__line__": n}
        if k in cur:
            raise RegistryError(f"{REGISTRY_PATH}:{n}: duplicate key '{k}' in one record")
        cur[k] = v
        key = k
    if cur:
        records.append(cur)
    return records


def load_registry(corpus_ids, path=REGISTRY_PATH):
    """Parse + validate the declared-divergence registry. Raises RegistryError."""
    if not path.exists():
        raise RegistryError(f"declared-divergence registry not found: {path}")
    waivers, seen = [], {}
    for rec in _parse_records(path.read_text()):
        n = rec.pop("__line__")
        where = f"{path}:{n}"

        for required in ("invocation", "stream", "authority", "reason"):
            if not rec.get(required):
                raise RegistryError(f"{where}: '{required}:' is required and must be non-empty")

        inv_id, stream = rec.pop("invocation"), rec.pop("stream")
        if stream not in STREAMS:
            raise RegistryError(f"{where}: stream '{stream}' not one of {', '.join(STREAMS)}")
        if inv_id not in corpus_ids:
            raise RegistryError(f"{where}: invocation '{inv_id}' is not in the corpus — "
                                f"a waiver for an id that does not exist can never fire")

        authority = [a.strip() for a in rec.pop("authority").split(",") if a.strip()]
        if not authority:
            raise RegistryError(f"{where}: 'authority:' is required and must be non-empty")
        for a in authority:
            if not AUTHORITY_RE.match(a):
                raise RegistryError(
                    f"{where}: authority '{a}' is not a citation — expected an R-0.7a clause "
                    f"(e.g. R-0.7a(i)), a requirement id (e.g. R-1.4a), or "
                    f"`.devlocal/go-port/exit-code-map.md:<line>`")

        reason = rec.pop("reason")
        expect, waive = None, None
        if stream == "exit_code":
            if "waive" in rec:
                raise RegistryError(f"{where}: 'waive:' is not valid for stream exit_code; "
                                    f"use 'expect: <reference> -> <candidate>'")
            m = EXPECT_RE.match(rec.pop("expect", ""))
            if not m:
                raise RegistryError(f"{where}: stream exit_code requires "
                                    f"'expect: <reference> -> <candidate>'")
            expect = (int(m.group(1)), int(m.group(2)))
        else:
            if "expect" in rec:
                raise RegistryError(f"{where}: 'expect:' is only valid for stream exit_code")
            waive = rec.pop("waive", "")
            allowed = ("whole-stream", "usage-block") if stream in ("stdout", "stderr") \
                else ("whole-stream",)
            if waive not in allowed:
                raise RegistryError(f"{where}: stream {stream} requires "
                                    f"'waive:' one of {', '.join(allowed)}")

        step = rec.pop("step", None)
        if step is not None:
            if not step.isdigit() or int(step) < 1:
                raise RegistryError(f"{where}: 'step:' must be a 1-based integer")
            step = int(step)

        if rec:
            raise RegistryError(f"{where}: unknown key(s): {', '.join(sorted(rec))}")

        w = Waiver(inv_id, stream, authority, reason, expect, waive, step, n)
        if w.key in seen:
            raise RegistryError(f"{where}: duplicate declaration for "
                                f"{inv_id} / {stream} / step={step} "
                                f"(first at {path}:{seen[w.key]})")
        seen[w.key] = n
        waivers.append(w)
    return waivers


def strip_usage_block(text):
    """Drop everything before the first `company-os...: error:` line.

    Waives argparse's COLUMNS-wrapped usage block (R-0.7a(i)) while leaving the
    diagnostic line itself intact for comparison (R-1.4a). A stream with no
    diagnostic line is returned unchanged — nothing is silently dropped."""
    lines = text.splitlines(keepends=True)
    for i, line in enumerate(lines):
        if USAGE_ERR_RE.match(line):
            return "".join(lines[i:])
    return text


def apply_waivers(inv_id, blocks, waivers):
    """Split compare()'s blocks into (declared, remaining) and record firing."""
    declared, remaining = [], []
    for b in blocks:
        w = next((w for w in waivers if w.applies(inv_id, b["stream"], b["step"])), None)
        if w is None:
            remaining.append(b)
            continue
        w.matched += 1
        if b["stream"] == "exit_code":
            if (b["ref"], b["cand"]) == w.expect:
                w.resolved += 1
                declared.append((b, w))
            else:
                b["note"] = (f"declared {w.expect[0]} -> {w.expect[1]} "
                             f"({w.cite}), observed {b['ref']} -> {b['cand']}")
                remaining.append(b)
            continue
        if w.waive == "whole-stream":
            w.resolved += 1
            declared.append((b, w))
            continue
        # usage-block: waive the block, still compare the diagnostic (R-1.4a).
        ra, ca = strip_usage_block(b["ref"]), strip_usage_block(b["cand"])
        if ra == ca:
            w.resolved += 1
            declared.append((b, w))
            continue
        b["body"] = indent(udiff(ra, ca, b["label"]))
        b["note"] = (f"usage block waived ({w.cite}); the diagnostic line still "
                     f"differs and R-1.4a does not waive it")
        remaining.append(b)
    return declared, remaining


# ------------------------------------------------------ synthetic fixture set


def build_synthetic_fixtures(base, report):
    """Create the fixtures that are not committed: an empty dir, the bad-manifest
    workspaces, and (git permitting) a real source repo + federated workspace."""
    SYNTHETIC["empty"] = base / "empty"
    (base / "empty").mkdir(parents=True)

    for name, text in BAD_MANIFESTS.items():
        d = base / f"badmanifest-{name}"
        shutil.copytree(FIXTURES["standalone-team"], d)
        (d / "workspace.yaml").write_text(text)
        SYNTHETIC[f"badmanifest-{name}"] = d

    ok, why = git_ok()
    if not ok:
        report.skip_reason = why
        return
    src = base / "gitsrc"
    (src / "governance").mkdir(parents=True)
    (src / "components").mkdir(parents=True)
    (src / "docs" / "sdd").mkdir(parents=True)
    (src / "src").mkdir(parents=True)
    (src / "governance" / "requirements.yaml").write_text(
        "version: 1\nplatform: testplat\nrequirements: []\n")
    (src / "components" / "foo.yaml").write_text(
        "id: foo\ncomponentType: service\nownership:\n  accountableTeam: team://none\n")
    (src / "docs" / "sdd" / "spec.md").write_text("# Spec\n")
    (src / "README.md").write_text("# not governance - must not be sliced\n")
    (src / "src" / "app.py").write_text("print('not governance')\n")
    env = dict(os.environ)
    # Fixed identity + dates so the commit SHA is reproducible run to run; the
    # SHA is baked into the manifest that BOTH binaries then read.
    env.update({
        "GIT_AUTHOR_NAME": "t", "GIT_AUTHOR_EMAIL": "t@e",
        "GIT_COMMITTER_NAME": "t", "GIT_COMMITTER_EMAIL": "t@e",
        "GIT_AUTHOR_DATE": "2020-01-01T00:00:00+0000",
        "GIT_COMMITTER_DATE": "2020-01-01T00:00:00+0000",
    })
    for cmd in (["init", "-q", "-b", "main"], ["add", "-A"], ["commit", "-q", "-m", "init"]):
        r = subprocess.run(["git", "-C", str(src)] + cmd, env=env,
                           capture_output=True, text=True)
        if r.returncode != 0:
            report.skip_reason = f"git setup failed: {r.stderr.strip()[:200]}"
            return
    sha = subprocess.run(["git", "-C", str(src), "rev-parse", "HEAD"],
                         capture_output=True, text=True).stdout.strip()

    def manifest(commit):
        return (
            "version: 1\nrepos:\n"
            "  - name: testplat\n"
            f"    url: file://{src}\n"
            "    pin:\n"
            f"      commit: {commit}\n"
            "    slices:\n"
            "      - localDirectory: platforms/testplat\n"
            "        paths: [governance/, components/]\n"
            "      - localDirectory: knowledge/testplat\n"
            "        paths: [docs/sdd]\n"
        )

    for key, commit in (("gitfed", sha),
                        ("gitfed-badpin", sha[:8]),
                        ("gitfed-missingref", "b" * 40)):
        d = base / key
        d.mkdir()
        (d / "workspace.yaml").write_text(manifest(commit))
        SYNTHETIC[key] = d
    report.git_available = True


def git_ok():
    if not shutil.which("git"):
        return False, "git not found on PATH"
    try:
        out = subprocess.run(["git", "--version"], capture_output=True, text=True).stdout
    except OSError as exc:
        return False, f"git --version failed: {exc}"
    m = re.search(r"(\d+)\.(\d+)", out)
    if not m:
        return False, f"could not parse `git --version`: {out.strip()!r}"
    major, minor = int(m.group(1)), int(m.group(2))
    if (major, minor) < (2, 27):
        return False, f"git {major}.{minor} < 2.27 (cone-mode sparse-checkout required)"
    return True, ""


# ------------------------------------------------------------------- runner


def fixture_path(name):
    if name in SYNTHETIC:
        return SYNTHETIC[name]
    return FIXTURES[name]


def unavailable(name, report):
    """'' when the fixture can be used, else the loud reason it cannot."""
    if name in SYNTHETIC:
        return ""
    if name not in FIXTURES:
        return (f"fixture '{name}' could not be synthesized: "
                f"{report.skip_reason or 'unknown reason'}")
    if not FIXTURES[name].is_dir():
        if name in OPTIONAL_FIXTURES:
            return (f"optional fixture '{name}' is not committed yet "
                    f"(expected at {FIXTURES[name]}) — task 0.2 output")
        return f"fixture '{name}' missing at {FIXTURES[name]}"
    return ""


def chmod_writable(path):
    for root, dirs, files in os.walk(path):
        for d in dirs:
            p = os.path.join(root, d)
            try:
                os.chmod(p, os.stat(p).st_mode | stat.S_IWUSR | stat.S_IXUSR)
            except OSError:
                pass
        for f in files:
            p = os.path.join(root, f)
            try:
                os.chmod(p, os.stat(p).st_mode | stat.S_IWUSR)
            except OSError:
                pass
    try:
        os.chmod(path, os.stat(path).st_mode | stat.S_IWUSR | stat.S_IXUSR)
    except OSError:
        pass


def snapshot_tree(root, norm):
    """{relpath: (kind, mode, digest)} plus {relpath: normalized text} for diffs.

    Skips any path with a `.git` component: git's own object store carries reflog
    timestamps and pack nondeterminism that no CLI produced. Materialized slices
    are fully compared, mode bits included."""
    tree, texts = {}, {}
    for dirpath, dirnames, filenames in os.walk(root):
        rel_dir = os.path.relpath(dirpath, root)
        if rel_dir == ".":
            rel_dir = ""
        if ".git" in Path(rel_dir).parts:
            dirnames[:] = []
            continue
        dirnames[:] = sorted(d for d in dirnames if d != ".git")
        for d in dirnames:
            rel = os.path.join(rel_dir, d) if rel_dir else d
            tree[rel] = ("dir", stat.S_IMODE(os.lstat(os.path.join(dirpath, d)).st_mode), "")
        for f in sorted(filenames):
            full = os.path.join(dirpath, f)
            rel = os.path.join(rel_dir, f) if rel_dir else f
            st = os.lstat(full)
            if stat.S_ISLNK(st.st_mode):
                tree[rel] = ("link", 0, norm(os.readlink(full)))
                continue
            try:
                raw = open(full, "rb").read()
            except OSError as exc:
                tree[rel] = ("unreadable", stat.S_IMODE(st.st_mode), str(exc))
                continue
            try:
                text = norm(raw.decode("utf-8"))
                texts[rel] = text
                payload = text.encode("utf-8")
            except UnicodeDecodeError:
                payload = raw
            tree[rel] = ("file", stat.S_IMODE(st.st_mode),
                         hashlib.sha256(payload).hexdigest())
    return tree, texts


def run_side(binary, invocation, workdir):
    """Copy the fixture, run every step, snapshot. Returns (steps, tree, texts)."""
    ws = workdir / "ws"
    src = fixture_path(invocation.fixture)
    shutil.copytree(src, ws, symlinks=True)
    chmod_writable(ws)
    norm = normalizer(ws)
    env = dict(os.environ)
    env.pop("COMPANY_OS_WORKSPACE_ROOT", None)
    env["LC_ALL"] = "C"
    env["TERM"] = "dumb"
    env["COLUMNS"] = "80"
    results = []
    for step in invocation.steps:
        if step == ["__tamper__"]:
            # deterministic in-tree mutation used to drive gate [8/8]
            target = ws / "platforms" / "testplat" / "governance" / "requirements.yaml"
            try:
                if target.exists():
                    os.chmod(target.parent, 0o755)
                    os.chmod(target, 0o644)
                    target.write_text("version: 1\nplatform: tampered\nrequirements: []\n")
                results.append((step, 0, "", ""))
            except OSError as exc:
                results.append((step, 0, "", f"tamper failed: {exc}"))
            continue
        argv = [str(binary)]
        if "--root" not in step:
            argv += ["--root", str(ws)]
        argv += step
        try:
            proc = subprocess.run(argv, cwd=str(ws), env=env, stdin=subprocess.DEVNULL,
                                  capture_output=True, text=True, timeout=180)
            out, err, code = proc.stdout, proc.stderr, proc.returncode
        except subprocess.TimeoutExpired:
            out, err, code = "", "<TIMEOUT after 180s>", -999
        results.append((step, code, norm(out), norm(err)))
    tree, texts = snapshot_tree(ws, norm)
    return results, tree, texts


def udiff(a, b, label, la="reference", lb="candidate"):
    lines = list(difflib.unified_diff(a.splitlines(keepends=True),
                                      b.splitlines(keepends=True),
                                      fromfile=f"{la}:{label}", tofile=f"{lb}:{label}",
                                      lineterm="\n"))
    if lines and not lines[-1].endswith("\n"):
        lines[-1] += "\n"
    return "".join(lines)


def _block(stream, step, kind, label, body, ref=None, cand=None):
    """One divergence, keyed by (stream, step) so the registry can address it."""
    return {"stream": stream, "step": step, "kind": kind, "label": label,
            "body": body, "ref": ref, "cand": cand, "note": ""}


def render(b):
    head = f"  {b['kind']:9s}  {b['label']}"
    if b["note"]:
        head += f"\n    ! {b['note']}"
    return head + ("\n" + b["body"].rstrip("\n") if b["body"] else "")


def compare(invocation, ref, cand):
    """Return a list of divergence blocks (empty == PASS), each keyed by stream."""
    rsteps, rtree, rtexts = ref
    csteps, ctree, ctexts = cand
    blocks = []
    for i, (rs, cs) in enumerate(zip(rsteps, csteps), 1):
        label = f"step {i}: {' '.join(rs[0]) or '(no args)'}"
        if rs[1] != cs[1]:
            blocks.append(_block("exit_code", i, "EXIT CODE", label,
                                 f"    reference={rs[1]}  candidate={cs[1]}\n",
                                 ref=rs[1], cand=cs[1]))
        if rs[2] != cs[2] and invocation.unstable_stream != "stdout":
            blocks.append(_block("stdout", i, "STDOUT", label,
                                 indent(udiff(rs[2], cs[2], "stdout")),
                                 ref=rs[2], cand=cs[2]))
        if rs[3] != cs[3] and invocation.unstable_stream != "stderr":
            blocks.append(_block("stderr", i, "STDERR", label,
                                 indent(udiff(rs[3], cs[3], "stderr")),
                                 ref=rs[3], cand=cs[3]))

    def tree(kind, label, body):
        blocks.append(_block("file_tree", None, kind, label, body))

    only_ref = sorted(set(rtree) - set(ctree))
    only_cand = sorted(set(ctree) - set(rtree))
    if only_ref:
        tree("FILE TREE", "present only under reference:",
             "".join(f"    - {p}\n" for p in only_ref[:40])
             + ("    ...\n" if len(only_ref) > 40 else ""))
    if only_cand:
        tree("FILE TREE", "present only under candidate:",
             "".join(f"    + {p}\n" for p in only_cand[:40])
             + ("    ...\n" if len(only_cand) > 40 else ""))
    for p in sorted(set(rtree) & set(ctree)):
        rk, rm, rd = rtree[p]
        ck, cm, cd = ctree[p]
        if rk != ck:
            tree("FILE TREE", f"{p}: kind {rk} != {ck}", "")
            continue
        if rm != cm:
            tree("FILE MODE", f"{p}: {oct(rm)} != {oct(cm)}", "")
        if rd != cd:
            if p in rtexts and p in ctexts:
                tree("FILE DIFF", p, indent(udiff(rtexts[p], ctexts[p], p)))
            else:
                tree("FILE DIFF",
                     f"{p}: binary content differs ({rd[:12]} != {cd[:12]})", "")
    return blocks


def indent(text, pad="    "):
    return "".join(pad + line for line in text.splitlines(keepends=True))


class Report:
    def __init__(self):
        self.git_available = False
        self.skip_reason = ""
        self.passed = []
        self.diverged = []
        self.skipped = []
        self.partial = []
        self.declared = []      # (invocation id, [(stream, citation, reason)])
        self.ran = set()        # invocation ids actually executed (not SKIPped)


def main():
    ap = argparse.ArgumentParser(
        description="Cross-implementation differential harness for company-os.")
    ap.add_argument("reference", help="reference binary (e.g. company-os-starter/bin/company-os)")
    ap.add_argument("candidate", help="candidate binary (the Go build)")
    ap.add_argument("--only", action="append", default=[],
                    help="restrict to invocation ids/groups containing this substring "
                         "(repeatable)")
    ap.add_argument("--list", action="store_true", help="list the corpus and exit")
    ap.add_argument("-v", "--verbose", action="store_true", help="print every PASS line")
    ap.add_argument("--keep", action="store_true", help="keep the temp run dirs")
    ap.add_argument("--list-waivers", action="store_true",
                    help="list the declared-divergence registry and exit")
    args = ap.parse_args()

    if args.list:
        for c in CORPUS:
            print(f"{c.id:55s} fixture={c.fixture}  steps={len(c.steps)}")
        print(f"\n{len(CORPUS)} invocations")
        return 0

    try:
        waivers = load_registry({c.id for c in CORPUS})
    except RegistryError as exc:
        print(f"error: declared-divergence registry rejected: {exc}", file=sys.stderr)
        return 2

    if args.list_waivers:
        print(f"declared-divergence registry ({REGISTRY_PATH}) — {len(waivers)} entries "
              f"(R-7.1a)")
        for w in sorted(waivers, key=lambda w: (w.stream, w.invocation)):
            what = (f"exit {w.expect[0]} -> {w.expect[1]}" if w.expect
                    else f"waive {w.waive}")
            step = "" if w.step is None else f" step={w.step}"
            print(f"  {w.invocation:48s} {w.stream:10s}{step} {what}")
            print(f"      authority: {w.cite}")
            print(f"      reason   : {w.reason}")
        return 0

    ref_bin, cand_bin = Path(args.reference).resolve(), Path(args.candidate).resolve()
    for b in (ref_bin, cand_bin):
        if not b.exists():
            print(f"error: binary not found: {b}", file=sys.stderr)
            return 2

    self_check = ref_bin == cand_bin
    report = Report()
    base = Path(tempfile.mkdtemp(prefix="company-os-diff-"))
    started = time.time()
    try:
        build_synthetic_fixtures(base / "fixtures", report)

        selected = [c for c in CORPUS
                    if not args.only or any(o in c.id or o in c.group for o in args.only)]

        print("=" * 78)
        print("company-os cross-implementation differential harness (task 0.3)")
        print(f"  reference : {ref_bin}")
        print(f"  candidate : {cand_bin}")
        print(f"  corpus    : {len(selected)} invocation(s)"
              f"{'' if len(selected) == len(CORPUS) else f' of {len(CORPUS)}'}")
        print(f"  git       : {'available' if report.git_available else 'UNAVAILABLE'}"
              f"{'' if report.git_available else ' -> ' + report.skip_reason}")
        print("  normalized: <WS> (temp workspace path), <TS> (generated UTC timestamp)")
        print("  tree scope: every path except `.git` internals; mode bits compared")
        print(f"  waivers   : {len(waivers)} declared divergence(s) from "
              f"{REGISTRY_PATH.name} (R-7.1a)"
              + ("  [SELF-CHECK: reference IS candidate, so no waiver can fire "
                 "and the stale check is not evaluated]" if self_check else ""))
        print("=" * 78)

        runs = base / "runs"
        for idx, c in enumerate(selected, 1):
            why = unavailable(c.fixture, report)
            if why:
                report.skipped.append((c.id, why))
                print(f"[{idx:3d}/{len(selected)}] SKIP     {c.id}\n"
                      f"            reason: {why}")
                continue
            rdir, cdir = runs / f"{idx:04d}-ref", runs / f"{idx:04d}-cand"
            rdir.mkdir(parents=True)
            cdir.mkdir(parents=True)
            report.ran.add(c.id)
            ref = run_side(ref_bin, c, rdir)
            cand = run_side(cand_bin, c, cdir)
            declared, blocks = apply_waivers(c.id, compare(c, ref, cand), waivers)
            if declared:
                report.declared.append(
                    (c.id, [(b["stream"], w.cite, w.reason) for b, w in declared]))
            if blocks:
                report.diverged.append(c.id)
                print(f"[{idx:3d}/{len(selected)}] DIVERGE  {c.id}"
                      f"{'  — ' + c.note if c.note else ''}")
                for b in blocks:
                    print(render(b))
                for b, w in declared:
                    print(f"  (declared {b['stream']} — {w.cite})")
            elif declared:
                print(f"[{idx:3d}/{len(selected)}] DECLARED {c.id}")
                for b, w in declared:
                    print(f"  {b['stream']:10s} {w.cite}")
            elif c.unstable:
                report.partial.append((c.id, c.unstable))
                codes = ",".join(str(s[1]) for s in ref[0])
                print(f"[{idx:3d}/{len(selected)}] PARTIAL  {c.id}  (exit {codes}) "
                      f"— {c.unstable_stream} NOT compared")
            else:
                report.passed.append(c.id)
                if args.verbose:
                    codes = ",".join(str(s[1]) for s in ref[0])
                    print(f"[{idx:3d}/{len(selected)}] PASS     {c.id}  (exit {codes})")
            if not args.keep:
                chmod_writable(rdir)
                chmod_writable(cdir)
                shutil.rmtree(rdir, ignore_errors=True)
                shutil.rmtree(cdir, ignore_errors=True)

        elapsed = time.time() - started
        # A declaration is STALE when its invocation ran and no longer diverges
        # on that stream — the waiver never fired and is now hiding nothing.
        #
        # Not evaluated in self-check mode (reference IS candidate): the registry
        # declares cross-implementation differences, so under one binary NO entry
        # can fire and every one of them would report stale. Suppressed loudly,
        # never silently — self_check is printed in the summary either way.
        stale = ([] if self_check else
                 [w for w in waivers if w.invocation in report.ran and w.matched == 0])

        print("=" * 78)
        declared_ids = {d[0] for d in report.declared}
        by_group = {}
        for c in selected:
            by_group.setdefault(c.group, [0, 0, 0, 0])
            if c.id in report.diverged:
                by_group[c.group][2] += 1
            elif any(c.id == s[0] for s in report.skipped):
                by_group[c.group][3] += 1
            elif c.id in declared_ids:
                by_group[c.group][1] += 1
            else:
                by_group[c.group][0] += 1
        print("per-command breakdown (pass+partial / declared / diverge / skip):")
        for g in sorted(by_group):
            p, dc, d, s = by_group[g]
            print(f"  {g:20s} {p:4d} / {dc:4d} / {d:4d} / {s:4d}")
        print("-" * 78)
        print(f"invocations : {len(selected)}")
        print(f"PASS        : {len(report.passed)}")
        print(f"DECLARED    : {len(report.declared)}")
        print(f"PARTIAL     : {len(report.partial)}")
        print(f"DIVERGE     : {len(report.diverged)}")
        print(f"STALE       : {len(stale)}"
              + ("  (not evaluated: self-check, reference IS candidate)"
                 if self_check else ""))
        print(f"SKIP        : {len(report.skipped)}")
        if report.declared:
            print("DECLARED (sanctioned divergence — each entry cites its authority):")
            for did, items in report.declared:
                for stream, cite, reason in items:
                    print(f"  - {did} [{stream}]")
                    print(f"      authority: {cite}")
                    print(f"      reason   : {reason}")
        if stale:
            print("STALE (declared but the invocation no longer diverges on that stream —")
            print("       delete the entry; a stale waiver re-hides the bug when it returns):")
            for w in stale:
                what = (f"exit {w.expect[0]} -> {w.expect[1]}" if w.expect
                        else f"waive {w.waive}")
                print(f"  - {w.invocation} [{w.stream}] {what}")
                print(f"      declared at {REGISTRY_PATH}:{w.lineno} ({w.cite})")
        if report.partial:
            print("PARTIAL (one stream excluded — declared, not silently dropped):")
            for pid, why in report.partial:
                print(f"  - {pid}: {why}")
        if report.skipped:
            print("SKIPPED (not silently omitted):")
            for sid, why in report.skipped:
                print(f"  - {sid}: {why}")
        print(f"runtime     : {elapsed:.1f}s")
        print("=" * 78)
        if report.diverged:
            print("RESULT: DIVERGENCE DETECTED — the two implementations are NOT at parity.")
            return 1
        if stale:
            print("RESULT: STALE DECLARATION(S) — no undeclared divergence, but the registry "
                  "claims a difference that no longer exists.")
            return 1
        print("RESULT: ZERO UNDECLARED DIVERGENCE across the corpus"
              f"{f' ({len(report.declared)} declared)' if report.declared else ''}.")
        return 0
    finally:
        if args.keep:
            print(f"(temp dirs kept at {base})")
        else:
            chmod_writable(base)
            shutil.rmtree(base, ignore_errors=True)


if __name__ == "__main__":
    sys.exit(main())
