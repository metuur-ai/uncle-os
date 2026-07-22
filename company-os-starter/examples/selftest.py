#!/usr/bin/env python3
"""Unit-level checks for the derived/validated helpers — part of the Unit 0
characterization harness. Run by acceptance.sh; exits non-zero on any failure.

Covers the invariants that have no other oracle in a repo with no test suite:
canonical compare (R-0.1), preserve-unknown-fields (R-1.5), and the fail-safe
generated-block rewrite across every marker state (R-3.3/3.4/3.5/3.6)."""
import sys
import tempfile
from importlib.machinery import SourceFileLoader
from pathlib import Path

CLI = Path(__file__).resolve().parent.parent / "bin" / "company-os"
co = SourceFileLoader("co", str(CLI)).load_module()

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

print(f"\nselftest: {len(fails)} failure(s)" if fails else "\nselftest: PASS")
sys.exit(1 if fails else 0)
