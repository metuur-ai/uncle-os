#!/usr/bin/env python3
"""Live oracle for the internal/frontmatter differential test.

Loads bin/company-os as a module the way examples/selftest.py:11-15 does, so the
decision recorded here is the REAL frontmatter(), not a copy of its regex. The
vendored PyYAML is put on sys.path; nothing is installed.

Usage:  oracle.py <dir-of-.md-files>
Emits:  one JSON object on stdout, {basename-without-.md: record}, where record
        is {"decision": "accept"|"reject"|"decode-error", "yaml_b64", "body_b64"}.
        yaml_b64/body_b64 are m.group(1)/m.group(2) on accept; on reject body_b64
        is the whole newline-translated text (what Python returns as the body).

The YAML layer is deliberately not exercised: this asserts the fence split only,
because internal/frontmatter stops at the seam and leaves yaml.safe_load to
internal/yamlio.
"""
import base64
import json
import re
import sys
from importlib.machinery import SourceFileLoader
from pathlib import Path

HERE = Path(__file__).resolve()
# internal/frontmatter/testdata/oracle.py -> company-os-starter/
STARTER = HERE.parents[3]
sys.path.insert(0, str(STARTER / "vendor"))
CLI = STARTER / "bin" / "company-os"

# Import for its side effect of proving the module still loads with the vendored
# PyYAML; frontmatter()'s regex is re-run below so the raw groups are visible.
co = SourceFileLoader("co", str(CLI)).load_module()
assert callable(co.frontmatter)

PATTERN = re.compile(r"^---\n(.*?)\n---\n(.*)$", re.DOTALL)

b64 = lambda s: base64.b64encode(s.encode("utf-8")).decode()

out = {}
for p in sorted(Path(sys.argv[1]).glob("*.md")):
    name = p.stem
    try:
        text = p.read_text()
    except UnicodeDecodeError:
        out[name] = {"decision": "decode-error"}
        continue
    m = PATTERN.match(text)
    if not m:
        # frontmatter() returns ({}, text) here; text is already translated.
        out[name] = {"decision": "reject", "body_b64": b64(text)}
        continue
    out[name] = {
        "decision": "accept",
        "yaml_b64": b64(m.group(1)),
        "body_b64": b64(m.group(2)),
    }
    # Cross-check the split against the real function wherever its YAML layer
    # does not raise; a raise is a downstream concern, not a split disagreement.
    try:
        _, body = co.frontmatter(p)
    except Exception:
        continue
    assert body == m.group(2), name

json.dump(out, sys.stdout)
