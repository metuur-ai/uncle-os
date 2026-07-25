#!/usr/bin/env python3
"""Unit-level checks for the derived/validated helpers — part of the Unit 0
characterization harness. Run by acceptance.sh; exits non-zero on any failure.

Covers the invariants that have no other oracle in a repo with no test suite:
canonical compare (R-0.1), preserve-unknown-fields (R-1.5), and the fail-safe
generated-block rewrite across every marker state (R-3.3/3.4/3.5/3.6)."""
import subprocess
import sys
import tempfile
from importlib.machinery import SourceFileLoader
from pathlib import Path

CLI = Path(__file__).resolve().parent.parent / "company-os-starter" / "bin" / "company-os"
co = SourceFileLoader("co", str(CLI)).load_module()
HERE = Path(__file__).resolve().parent

fails = []


def check(name, cond):
    print(("ok   " if cond else "FAIL ") + name)
    if not cond:
        fails.append(name)


def rewrite(initial, block="B1"):
    d = Path(tempfile.mkdtemp())
    p = d / "CLAUDE.md"
    if initial is not None:
        p.write_text(initial)
    changed = co.rewrite_generated_block(p, block)
    return changed, (p.read_text() if p.exists() else None)


# R-0.1 canonical compare
check("R-0.1 canonical_yaml order-immune",
      co.canonical_yaml({"b": 1, "a": 2}) == co.canonical_yaml({"a": 2, "b": 1}))
check("R-0.1 blocks_equal ws/newline immune",
      co.blocks_equal("x \ny\n\n", "x\r\ny"))
check("R-0.1 blocks_equal detects difference",
      not co.blocks_equal("x\ny", "x\nz"))

# R-1.5 preserve unknown frontmatter keys through a tag rewrite
with tempfile.TemporaryDirectory() as d:
    p = Path(d) / "doc.md"
    p.write_text("---\ntype: prd\nid: x\nvendorField: keep-me\n---\n\nbody\n")
    co.rewrite_frontmatter_tags(p, co.derive_tags({"type": "prd", "id": "x"}))
    meta, _ = co.frontmatter(p)
    check("R-1.5 preserve-unknown frontmatter key", meta.get("vendorField") == "keep-me")

# R-3.4 — 0 markers + absent -> create with header + block
ch, txt = rewrite(None)
check("R-3.4 create when absent",
      ch and txt and "company-os:generated:start" in txt and "B1" in txt)

# R-3.5 — 0 markers + prose -> append, prose untouched, no failure
ch, txt = rewrite("# Hand-written\n\nkeep this prose\n")
check("R-3.5 append preserves prose",
      ch and "keep this prose" in txt and "company-os:generated:start" in txt)

# R-3.3 — one balanced pair -> replace interior only, verbatim outside
before = "PRE\n" + co.render_generated_region("OLD") + "\nPOST\n"
ch, txt = rewrite(before, "NEW")
check("R-3.3 replace interior only",
      "NEW" in txt and "OLD" not in txt
      and txt.startswith("PRE") and txt.rstrip().endswith("POST"))

# R-3.3 — rewriting an identical block yields an identical file (byte-preserve)
same = "PRE\n" + co.render_generated_region("NEW") + "\nPOST\n"
ch, txt = rewrite(same, "NEW")
check("R-3.3 identical block is a no-op", ch is False and txt == same)

# R-3.6 — start without end -> warn, mutate nothing
bad = "x\n" + co.GEN_START + "\nonly start\n"
ch, txt = rewrite(bad, "NEW")
check("R-3.6 unbalanced markers -> no mutation", ch is False and txt == bad)

# R-3.6 — duplicated pair -> warn, mutate nothing
dup = co.render_generated_region("A") + "\n" + co.render_generated_region("B") + "\n"
ch, txt = rewrite(dup, "NEW")
check("R-3.6 duplicated markers -> no mutation", ch is False and txt == dup)

# GPF-R-8.1 — fail-fast workspace-root resolution. The is_root() predicate:
# an empty dir is NOT a root; a full workspace and a teams-only standalone both
# ARE (any one canonical root suffices).
with tempfile.TemporaryDirectory() as d:
    check("R-8.1 empty dir is not a workspace root", co.Workspace(d).is_root() is False)
check("R-8.1 full workspace is a root",
      co.Workspace(HERE / "workspace").is_root() is True)
check("R-8.1 standalone teams-only dir is a root",
      co.Workspace(HERE / "standalone-team").is_root() is True)

# GPF-R-8.1 — a workspace-scoped command outside a root exits non-zero and names
# the resolution order in its message.
with tempfile.TemporaryDirectory() as d:
    r = subprocess.run([sys.executable, str(CLI), "--root", d, "validate"],
                       capture_output=True, text=True)
    check("R-8.1 validate outside a root exits non-zero", r.returncode != 0)
    check("R-8.1 message names resolution order",
          "--root -> $COMPANY_OS_WORKSPACE_ROOT -> current directory" in r.stderr)

# GPF-R-4.3 — *_SECTIONS is the single source: every required heading it names
# renders into the template AS a `## ` heading (the same list `validate` greps).
_disc = co.DISCOVERY_TEMPLATE.format(bid="b", title="t", team="tm",
                                     date="2026-01-01", ds=co.DISCOVERY_SECTIONS)
check("R-4.3 DISCOVERY_SECTIONS drives the template",
      all(f"## {s}" in _disc for s in co.DISCOVERY_SECTIONS))
_prd = co.PRD_TEMPLATE.format(pid="p", title="t", team="tm", platform="pf",
                              components="c", date="2026-01-01", discovery="none",
                              problem="P", metrics="M", ps=co.PRD_SECTIONS,
                              component_list="- c", governance_checklist="- [ ] none")
check("R-4.3 PRD_SECTIONS drives the template",
      all(f"## {s}" in _prd for s in co.PRD_SECTIONS))

# GPF-R-1.x — init/add/reality scaffolding: a fresh scaffold validates green,
# and every scaffolder refuses to overwrite. Driven through the process boundary
# (subprocess) with flags only, so no TTY is needed (GPF-R-1.3). Temp dirs must
# live outside any path containing "scratchpad" (iter_graph_docs filters those).
def _cli(root, *a):
    return subprocess.run([sys.executable, str(CLI), "--root", str(root), *a],
                          capture_output=True, text=True)


with tempfile.TemporaryDirectory(dir="/tmp") as d:
    r = _cli(d, "init", "--company", "Acme", "--team", "core", "--platform", "payments")
    check("R-1.1 init exits 0 (headless flags)", r.returncode == 0)
    check("R-1.7 init workspace validates green",
          _cli(d, "validate").returncode == 0)
    # GPF-R-1.2 — re-init inside an existing workspace refuses, mutates nothing
    check("R-1.2 re-init refuses (non-zero)",
          _cli(d, "init", "--company", "X", "--team", "y", "--platform", "z").returncode != 0)
    # GPF-R-1.9 — add grows the workspace and refuses to overwrite
    check("R-1.9 add component exits 0",
          _cli(d, "add", "component", "checkout-service", "--platform", "payments").returncode == 0)
    check("R-1.9 add refuses to overwrite (non-zero)",
          _cli(d, "add", "component", "checkout-service", "--platform", "payments").returncode != 0)
    # Unit 2 — reality new scaffolds a green reality doc and refuses to overwrite
    check("R-2.x reality new exits 0",
          _cli(d, "reality", "new", "checkout-service", "--platform", "payments").returncode == 0)
    check("R-2.x reality new refuses to overwrite (non-zero)",
          _cli(d, "reality", "new", "checkout-service", "--platform", "payments").returncode != 0)
    check("R-1.7 workspace still validates green after add + reality",
          _cli(d, "validate").returncode == 0)

# GPF-R-2.1 — ids list reads canonical IDs from the ontology registry (never a
# parallel index). load_registry is the single reader.
_reg = co.load_registry(co.Workspace(HERE / "workspace"))
_reg_ids = [e.get("id") for e in _reg]
check("R-2.1 load_registry reads registry entries",
      "component://customer-notification-service" in _reg_ids
      and "team://customer-engagement" in _reg_ids)
# absent/empty registry -> [] (helpful empty result, not a crash)
with tempfile.TemporaryDirectory() as d:
    check("R-2.1 missing registry -> empty list, no crash",
          co.load_registry(co.Workspace(d)) == [])

# GPF-R-2.3 — closest-match suggestion for an unknown component id (difflib).
_sug = co.suggest_ids(co.Workspace(HERE / "workspace"),
                      "customer-notifcation-servce", scheme="component")
check("R-2.3 closest-match suggests the real component id",
      _sug[:1] == ["component://customer-notification-service"])
check("R-2.3 suggestions scoped by scheme (<=3)", len(_sug) <= 3)

# GPF-R-4.1/4.2 — resolve_template first-found-wins, then byte-identical builtin.
with tempfile.TemporaryDirectory(dir="/tmp") as d:
    root = Path(d)
    ws = co.Workspace(root)
    txt, src = co.resolve_template(ws, "prd")
    check("R-4.2 builtin fallback is the *_TEMPLATE string, verbatim",
          txt == co.PRD_TEMPLATE and "PRD_TEMPLATE" in src)
    (root / "company-os" / "templates").mkdir(parents=True)
    (root / "company-os" / "templates" / "prd.md").write_text("COMPANY")
    txt, src = co.resolve_template(ws, "prd", team="t", platform="p")
    check("R-4.1 company override beats builtin",
          txt == "COMPANY" and src == "company-os/templates/prd.md")
    (root / "platforms" / "p" / "templates").mkdir(parents=True)
    (root / "platforms" / "p" / "templates" / "prd.md").write_text("PLATFORM")
    txt, src = co.resolve_template(ws, "prd", team="t", platform="p")
    check("R-4.1 platform override beats company",
          txt == "PLATFORM" and src == "platforms/p/templates/prd.md")
    (root / "teams" / "t" / "templates").mkdir(parents=True)
    (root / "teams" / "t" / "templates" / "prd.md").write_text("TEAM")
    txt, src = co.resolve_template(ws, "prd", team="t", platform="p")
    check("R-4.1 team override wins the whole chain",
          txt == "TEAM" and src == "teams/t/templates/prd.md")

# GPF-R-4.4 — an artifact scaffolded from a CUSTOM override template that omits a
# required heading FAILS validate naming exactly that heading (outputs validated,
# templates not). Driven through the process boundary with flags only.
with tempfile.TemporaryDirectory(dir="/tmp") as d:
    _cli(d, "init", "--company", "Acme", "--team", "core", "--platform", "payments")
    _cli(d, "add", "component", "checkout-service", "--platform", "payments")
    tdir = Path(d) / "teams" / "core" / "templates"
    tdir.mkdir(parents=True, exist_ok=True)
    # omits `## {ps[1]}` (Success metrics); keeps ps[0] and ps[2]
    (tdir / "prd.md").write_text(
        "---\ntype: prd\nid: {pid}\ntitle: {title}\nstatus: proposed\n"
        "team: {team}\nplatform: {platform}\ncomponents: [{components}]\n"
        "governanceSnapshot: {date}\ndecisionOwner: {title}-owner\ncreated: {date}\n"
        "fromDiscovery: {discovery}\n"
        "tags: [kind/prd, platform/{platform}, team/{team}, status/proposed]\n---\n\n"
        "# PRD: {title}\n\n## {ps[0]}\nP\n\n## {ps[2]}\nC\n\n"
        "## Affected components\n{component_list}\n\n"
        "## Applicable governance (snapshot {date})\n{governance_checklist}\n")
    _cli(d, "prd", "new", "--team", "core", "--platform", "payments",
         "--components", "checkout-service", "--title", "Broken")
    r = _cli(d, "prd", "validate", "--platform", "payments", "2026-broken")
    check("R-4.4 custom-template PRD fails validate (non-zero)", r.returncode != 0)
    check("R-4.4 failure names the missing heading",
          "Success metrics" in (r.stdout + r.stderr))

# GPF-R-3.1/3.2/3.3 — role translation is display-only: a mapped role yields a
# legend containing the canonical term; an unmapped role yields nothing (unchanged
# display, no error). role_glossary_lines is a pure function (touches no files).
_po = co.role_glossary_lines("product-owner")
check("R-3.1 mapped role shows plain-language alongside canonical term",
      any("exception" in ln for ln in _po) and len(_po) >= 2)
check("R-3.3 unmapped role -> no glossary, no error",
      co.role_glossary_lines("nobody-role") == [])

# GPF-R-5.1/5.2/5.3/5.4 — custom skills layering. Build a four-layer fixture in a
# temp workspace (outside any "scratchpad" path so graph gates ignore it) and
# exercise discovery+labels, merged ordering, shadowing, and extends resolution.
SKILL_FM = ("---\nid: {id}\ntype: skill\nversion: '1.0'\nauthority: {auth}\n"
            "tags: [authority/{auth}]\n{ext}---\n\n# {id}\n\n{steps}\n")


def _mkskill(path, sid, auth="canonical", extends=None, steps="1. (mandatory) do it"):
    path.parent.mkdir(parents=True, exist_ok=True)
    ext = f"extends: {extends}\n" if extends else ""
    path.write_text(SKILL_FM.format(id=sid, auth=auth, ext=ext, steps=steps))


with tempfile.TemporaryDirectory(dir="/tmp") as d:
    root = Path(d)
    (root / "company-os").mkdir()
    ws = co.Workspace(root)
    # one skill per shared layer + one personal rule = all four layers
    _mkskill(root / "company-os" / "skills" / "governing.SKILL.md",
             "skill://company/governing")
    _mkskill(root / "platforms" / "communications" / "skills" / "creating-prd.SKILL.md",
             "skill://product/creating-prd")
    _mkskill(root / "teams" / "core" / "skills" / "team-extra.SKILL.md",
             "skill://team/team-extra", auth="team")
    (root / "teams" / "core" / "scratchpad" / "personal-rules").mkdir(parents=True)
    (root / "teams" / "core" / "scratchpad" / "personal-rules" / "maria.md").write_text(
        "---\ntype: personal-rule\ntags: [authority/personal]\n---\n\n# Maria\n- my rule\n")

    sk = co.discover_skills(ws)
    layers = {s["layer"] for s in sk}
    # GPF-R-5.1 — all four layers discovered, each origin-labeled
    check("R-5.1 four layers discovered",
          layers == {"company", "platform", "team", "personal"} and len(sk) == 4)
    _plat = [s for s in sk if s["layer"] == "platform"][0]
    check("R-5.1 platform skill labeled with its platform origin",
          _plat["platform"] == "communications" and _plat["name"] == "creating-prd")
    _pers = [s for s in sk if s["layer"] == "personal"][0]
    check("R-5.1 personal rule labeled personal (no canonical authority)",
          _pers["team"] == "core" and _pers["authority"] != "canonical")

    # GPF-R-5.4 — merged view orders canonical (company/platform) skills, whose
    # mandatory steps rank above personal rules (rendered last, non-overriding).
    canon = [s for s in sk if s["layer"] in ("company", "platform", "team")]
    check("R-5.4 canonical skills precede personal rules in the merged view",
          len(canon) == 3 and _pers["layer"] == "personal")

    # GPF-R-5.2 — no shadowing yet -> no conflicts (absence/clean tolerant)
    check("R-5.2 clean layering has no conflicts", co.skill_conflicts(ws) == [])

    # GPF-R-5.2 — a team skill reusing the canonical name -> shadowing names BOTH
    _mkskill(root / "teams" / "core" / "skills" / "creating-prd.SKILL.md",
             "skill://team/creating-prd-copy", auth="team")
    probs = co.skill_conflicts(ws)
    check("R-5.2 shadowing detected", any("shadowing" in p for p in probs))
    check("R-5.2 shadowing names both files",
          any("teams/core/skills/creating-prd.SKILL.md" in p
              and "platforms/communications/skills/creating-prd.SKILL.md" in p
              for p in probs))

with tempfile.TemporaryDirectory(dir="/tmp") as d:
    root = Path(d)
    (root / "company-os").mkdir()
    ws = co.Workspace(root)
    _mkskill(root / "platforms" / "communications" / "skills" / "creating-prd.SKILL.md",
             "skill://product/creating-prd")
    # GPF-R-5.3 — extends resolves to the platform-layer base skill FILE
    _base = co.resolve_extends(ws, "platform-skill://communications/creating-prd")
    check("R-5.3 extends resolves to the platform base skill",
          _base is not None and _base["name"] == "creating-prd")
    check("R-5.3 unknown extends target resolves to None",
          co.resolve_extends(ws, "platform-skill://communications/ghost") is None)
    # a distinct-named team skill extending the base is NOT shadowing, and valid
    _mkskill(root / "teams" / "core" / "skills" / "creating-prd-mobile.SKILL.md",
             "skill://team/creating-prd-mobile", auth="team",
             extends="platform-skill://communications/creating-prd",
             steps="1. (default) add mobile mock")
    check("R-5.3 valid extends is not a conflict", co.skill_conflicts(ws) == [])
    # GPF-R-5.3 — a dangling extends target fails, naming the URI
    _mkskill(root / "teams" / "core" / "skills" / "broken.SKILL.md",
             "skill://team/broken", auth="team",
             extends="platform-skill://communications/nonexistent")
    dang = co.skill_conflicts(ws)
    check("R-5.3 dangling extends detected and names the URI",
          any("dangling extends" in p and "platform-skill://communications/nonexistent" in p
              for p in dang))

# ---------------------------------------------------------- Phase 4 federation
# Pure-function coverage for the pin-validation contract (GPF-R-6.3) and the
# hand-edit content-hash oracle (GPF-R-7.5); the git parts are a guarded
# integration check that skips when git is unavailable/too old.
import contextlib
import hashlib
import io
import os
import shutil


def _dies(fn):
    """True iff fn() calls die() (SystemExit), stderr suppressed."""
    try:
        with contextlib.redirect_stderr(io.StringIO()):
            fn()
        return False
    except SystemExit:
        return True


# GPF-R-6.3 — exactly one of commit:/tag:; floating refs rejected
check("R-6.3 commit pin accepted",
      co.repo_pin({"name": "r", "pin": {"commit": "abc123"}}) == ("commit", "abc123"))
check("R-6.3 tag pin accepted",
      co.repo_pin({"name": "r", "pin": {"tag": "v1.0.0"}}) == ("tag", "v1.0.0"))
check("R-6.3 branch pin rejected (floating)",
      _dies(lambda: co.repo_pin({"name": "r", "pin": {"branch": "main"}})))
check("R-6.3 bare ref pin rejected (floating)",
      _dies(lambda: co.repo_pin({"name": "r", "pin": {"ref": "HEAD"}})))
check("R-6.3 both commit+tag rejected (ambiguous)",
      _dies(lambda: co.repo_pin({"name": "r", "pin": {"commit": "a", "tag": "v1"}})))
check("R-6.3 empty pin rejected",
      _dies(lambda: co.repo_pin({"name": "r", "pin": {}})))

# Multi-slice manifests — one repo, one clone, N destinations. A bare
# localDirectory:/paths: pair must keep normalizing to a single slice.
_bare = {"name": "r", "url": "file:///x", "localDirectory": "knowledge/a",
         "paths": ["docs/sdd"]}
check("slices: bare pair normalizes to one slice",
      co._repo_slices(_bare) == [{"localDirectory": "knowledge/a",
                                  "paths": ["docs/sdd"]}])
check("slices: omitted paths default to DEFAULT_SLICE_PATHS",
      co._repo_slices({"name": "r", "localDirectory": "knowledge/a"})[0]["paths"]
      == list(co.DEFAULT_SLICE_PATHS))
check("slices: list yields one entry per slice",
      len(co._repo_slices({"name": "r", "slices": [
          {"paths": ["docs/sdd"], "localDirectory": "knowledge/a"},
          {"paths": ["arch"], "localDirectory": "knowledge/b"}]})) == 2)
check("slices: + top-level localDirectory rejected",
      _dies(lambda: co._repo_slices({"name": "r", "localDirectory": "knowledge/a",
                                     "slices": [{"localDirectory": "knowledge/b"}]})))
check("slices: + top-level paths rejected",
      _dies(lambda: co._repo_slices({"name": "r", "paths": ["d"],
                                     "slices": [{"localDirectory": "knowledge/b"}]})))
check("slices: neither localDirectory nor slices rejected",
      _dies(lambda: co._repo_slices({"name": "r"})))
check("slices: root: inside a slice names the rename",
      _dies(lambda: co._repo_slices({"name": "r",
                                     "slices": [{"root": "knowledge/a"}]})))

# sliceHash must not depend on manifest slice order: the aggregate is computed
# ONCE over the union, so merging in either order gives the same digest.
_f1, _f2 = {"b/2.md": "bb", "a/1.md": "aa"}, {"c/3.md": "cc"}
check("aggregate_hash is merge-order independent",
      co.aggregate_hash({**_f1, **_f2}) == co.aggregate_hash({**_f2, **_f1}))

# knowledge/ is the catalog root: bare targeting would freeze its generated node
check("knowledge/ bare target rejected (depth >= 2 required)",
      _dies(lambda: co.slice_rel("r", "knowledge")))
check("knowledge/<area> accepted",
      co.slice_rel("r", "knowledge/components/x") == "knowledge/components/x")
check("non-canonical root still rejected",
      _dies(lambda: co.slice_rel("r", "elsewhere/x")))
check("unsafe repo name rejected (it keys the cache dir)",
      _dies(lambda: co._validate_repo_entry(
          {"name": "../evil", "url": "file:///x", "localDirectory": "knowledge/a",
           "pin": {"commit": "abc"}}, 0)))


def _manifest_dies(root, body):
    (root / "workspace.yaml").write_text("version: 1\nrepos:\n" + body)
    return _dies(lambda: co.load_manifest(co.Workspace(root)))


# Overlapping targets corrupt the lock and break re-sync (the outer slice's
# _make_readonly freezes the inner one's parent). Rejected with a global view.
with tempfile.TemporaryDirectory(dir="/tmp") as d:
    r = Path(d)
    (r / "teams").mkdir()
    check("nested targets in one repo rejected", _manifest_dies(r,
          "  - name: a\n    url: file:///x\n    pin: {commit: dead}\n    slices:\n"
          "      - {paths: [d], localDirectory: knowledge/a}\n"
          "      - {paths: [e], localDirectory: knowledge/a/b}\n"))
    check("two repos sharing a target rejected", _manifest_dies(r,
          "  - name: a\n    url: file:///x\n    pin: {commit: dead}\n"
          "    localDirectory: knowledge/x\n"
          "  - name: b\n    url: file:///y\n    pin: {commit: beef}\n"
          "    localDirectory: knowledge/x\n"))
    check("sibling targets sharing a prefix are NOT nested", not _manifest_dies(r,
          "  - name: a\n    url: file:///x\n    pin: {commit: dead}\n    slices:\n"
          "      - {paths: [d], localDirectory: platforms/comms}\n"
          "      - {paths: [e], localDirectory: platforms/comms-v2}\n"))

# GPF-R-6.1 — absent manifest => monorepo mode (None), no behavior change
with tempfile.TemporaryDirectory(dir="/tmp") as d:
    check("R-6.1 no workspace.yaml => load_manifest None",
          co.load_manifest(co.Workspace(Path(d))) is None)

# GPF-R-7.5 — slice_state / federated_slice_problems detect a hand-edit by hash
with tempfile.TemporaryDirectory(dir="/tmp") as d:
    root = Path(d)
    ws = co.Workspace(root)
    slice_file = root / "platforms" / "p" / "governance" / "requirements.yaml"
    slice_file.parent.mkdir(parents=True)
    slice_file.write_text("schemaVersion: '1.0'\n")
    good = hashlib.sha256(slice_file.read_bytes()).hexdigest()
    # files-only entry: slice_state must NEVER grow a dependency on any other
    # lock key — that contract is what keeps --only per-repo and gate 8 cheap.
    lock_repo = {"name": "p", "resolvedCommit": "deadbeef", "sliceHash": "x",
                 "files": {"platforms/p/governance/requirements.yaml": good}}
    check("R-7.5 clean slice => state 'clean'",
          co.slice_state(ws, lock_repo) == "clean")
    (root / "workspace.yaml").write_text(
        "version: 1\nrepos:\n  - name: p\n    url: file:///x\n"
        "    localDirectory: platforms/p\n    pin: {commit: deadbeef}\n    paths: [governance/]\n")
    # federated_slice_problems DOES read slices: it compares the manifest's slice
    # set to the lock's, so a moved target without a re-sync cannot report clean.
    lock_entry = dict(lock_repo,
                      slices=[{"localDirectory": "platforms/p",
                               "paths": ["governance/"]}])
    (root / "workspace.lock.yaml").write_text(
        co.yaml.safe_dump({"version": 1, "repos": [lock_entry]}, sort_keys=False))
    manifest = co.load_manifest(ws)
    probs0, n0 = co.federated_slice_problems(ws, manifest)
    check("R-7.5 in-sync slice => no problems", probs0 == [] and n0 == 1)
    # slice-set drift: same bytes on disk, different target in the manifest
    moved = dict(lock_entry, slices=[{"localDirectory": "platforms/other",
                                      "paths": ["governance/"]}])
    (root / "workspace.lock.yaml").write_text(
        co.yaml.safe_dump({"version": 1, "repos": [moved]}, sort_keys=False))
    probsd, _ = co.federated_slice_problems(ws, co.load_manifest(ws))
    check("R-7.5 slice-set drift detected though bytes still match",
          any("slice set" in p and "workspace sync" in p for p in probsd))
    (root / "workspace.lock.yaml").write_text(
        co.yaml.safe_dump({"version": 1, "repos": [lock_entry]}, sort_keys=False))
    # hand-edit the read-derived slice
    slice_file.write_text("schemaVersion: '1.0'\n# sneaky\n")
    check("R-7.5 hand-edit => state 'drifted'",
          co.slice_state(ws, lock_repo) == "drifted")
    probs1, _ = co.federated_slice_problems(ws, manifest)
    check("R-7.5 hand-edit => problem names path + re-sync remedy",
          any("platforms/p/governance/requirements.yaml" in p
              and "workspace sync" in p for p in probs1))

# Guarded git integration — sparse governance-only materialization + frozen
# reproducibility. Skips cleanly when git is missing/too old (GPF-R-7.7).
def _git_ok():
    if not co._git_available():
        return False
    try:
        return co._git_version() >= co.MIN_GIT
    except SystemExit:
        return False


if not _git_ok():
    check("R-7.1/7.3 git integration SKIPPED (git unavailable/too old)", True)
else:
    with tempfile.TemporaryDirectory(dir="/tmp") as d:
        d = Path(d)
        src = d / "src"
        src.mkdir()

        def _g(*a):
            subprocess.run(["git", "-C", str(src), *a], check=True,
                           capture_output=True, text=True)
        subprocess.run(["git", "init", "-q", str(src)], check=True)
        _g("config", "user.email", "a@b.c")
        _g("config", "user.name", "t")
        (src / "governance").mkdir()
        (src / "governance" / "requirements.yaml").write_text("r: 1\n")
        (src / "README.md").write_text("NOT governance\n")  # must NOT materialize
        # NESTED allowlist entry — the knowledge-catalog shape. Depth-1 entries
        # like governance/ never exercised the parent-permission path in
        # _force_remove/copytree, so a re-sync used to raise PermissionError.
        (src / "docs" / "sdd").mkdir(parents=True)
        (src / "docs" / "sdd" / "spec.md").write_text("# Spec\n")
        _g("add", "-A")
        _g("commit", "-q", "-m", "c1")
        sha = subprocess.run(["git", "-C", str(src), "rev-parse", "HEAD"],
                             check=True, capture_output=True, text=True).stdout.strip()
        ws_dir = d / "ws"
        (ws_dir / "teams").mkdir(parents=True)
        # ONE repo, TWO destinations: a governance slice and a knowledge slice.
        (ws_dir / "workspace.yaml").write_text(
            "version: 1\nrepos:\n  - name: p\n    url: file://%s\n"
            "    pin: {commit: %s}\n    slices:\n"
            "      - {paths: [governance/], localDirectory: platforms/p}\n"
            "      - {paths: [docs/sdd], localDirectory: knowledge/p}\n" % (src, sha))
        cli = str(CLI)
        r = subprocess.run([sys.executable, cli, "--root", str(ws_dir),
                            "workspace", "sync"], capture_output=True, text=True)
        mat = list((ws_dir / "platforms" / "p").rglob("*"))
        files = [p.relative_to(ws_dir).as_posix() for p in mat if p.is_file()]
        check("R-7.1 sync materializes only allowlisted governance paths",
              r.returncode == 0 and files == ["platforms/p/governance/requirements.yaml"])
        check("R-7.1 non-governance file (README.md) NOT materialized",
              not any("README" in f for f in files))
        kfiles = [p.relative_to(ws_dir).as_posix()
                  for p in (ws_dir / "knowledge" / "p").rglob("*") if p.is_file()]
        check("multi-slice: second slice lands in its own knowledge target",
              kfiles == ["knowledge/p/docs/sdd/spec.md"])
        check("multi-slice: targets stay isolated (no cross-contamination)",
              not any("docs" in f for f in files))
        check("multi-slice: ONE cache dir for the whole repo",
              [p.name for p in (ws_dir / ".company-os" / "federation-cache").iterdir()]
              == ["p"])
        lock = co.load_yaml(ws_dir / "workspace.lock.yaml", {})
        check("R-7.2 lock records resolved SHA + per-file hash",
              lock["repos"][0]["resolvedCommit"] == sha
              and "platforms/p/governance/requirements.yaml" in lock["repos"][0]["files"])
        check("multi-slice: ONE flat lock entry whose files map is the union",
              len(lock["repos"]) == 1
              and sorted(lock["repos"][0]["files"]) == [
                  "knowledge/p/docs/sdd/spec.md",
                  "platforms/p/governance/requirements.yaml"]
              and len(lock["repos"][0]["slices"]) == 2)
        # Re-sync over the read-only tree: the regression guard for the nested
        # allowlist entry (docs/sdd) whose parent was frozen 0555 by the first run.
        r2 = subprocess.run([sys.executable, cli, "--root", str(ws_dir),
                             "workspace", "sync"], capture_output=True, text=True)
        check("multi-slice: re-sync over a read-only nested slice succeeds",
              r2.returncode == 0)

        def _tree():
            items = []
            for base in ("platforms", "knowledge"):
                for p in sorted((ws_dir / base).rglob("*")):
                    if p.is_file():
                        items.append(p.relative_to(ws_dir).as_posix() + hashlib.sha256(
                            p.read_bytes()).hexdigest())
            return hashlib.sha256("\n".join(items).encode()).hexdigest()
        h_before = _tree()
        # wipe slices, make source unreachable, materialize --frozen from lock
        for base in ("platforms", "knowledge"):
            for p in sorted((ws_dir / base).rglob("*"), reverse=True):
                os.chmod(p, 0o755)
            shutil.rmtree(ws_dir / base)
        os.rename(src, d / "src.gone")
        rf = subprocess.run([sys.executable, cli, "--root", str(ws_dir),
                             "workspace", "sync", "--frozen"],
                            capture_output=True, text=True)
        check("R-7.3 --frozen materializes offline from lock (source removed)",
              rf.returncode == 0)
        check("R-7.3 frozen tree hash == online sync tree hash (reproducible)",
              _tree() == h_before)


print(f"\nselftest: {len(fails)} failure(s)" if fails else "\nselftest: PASS")
sys.exit(1 if fails else 0)
