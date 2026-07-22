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

print(f"\nselftest: {len(fails)} failure(s)" if fails else "\nselftest: PASS")
sys.exit(1 if fails else 0)
