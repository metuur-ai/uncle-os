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

CLI = Path(__file__).resolve().parent.parent / "bin" / "company-os"
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

print(f"\nselftest: {len(fails)} failure(s)" if fails else "\nselftest: PASS")
sys.exit(1 if fails else 0)
