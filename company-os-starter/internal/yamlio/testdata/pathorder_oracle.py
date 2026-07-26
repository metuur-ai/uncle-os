#!/usr/bin/env python3
"""Live oracle for internal/yamlio's PathLess differential test (task 1.4).

hash_tree (bin/company-os:2422) orders a slice's files with
`sorted(src.rglob("*"))`, i.e. by CPython's PurePath ordering, which compares
component-wise rather than as raw strings. That order reaches
workspace.lock.yaml and, through it, gate 8's [FAIL] line order. This oracle
returns the REAL sorted(Path) answer so the Go comparator is measured against
CPython instead of against a reading of CPython.

Usage:  pathorder_oracle.py            # reads JSON [[path, ...], ...] on stdin
Emits:  JSON {"python": "x.y.z", "sorted": [[path, ...], ...]} on stdout, each
        inner list sorted as PurePosixPath, in the same order as the input.
"""
import json
import platform
import sys
from pathlib import PurePosixPath

# Sorted with PurePosixPath as the KEY, not by converting to PurePosixPath and
# back: str(PurePosixPath("a//b")) is "a/b", which would silently rewrite the
# inputs the Go side is being compared against.
groups = json.load(sys.stdin)
out = [sorted(group, key=PurePosixPath) for group in groups]
json.dump({"python": platform.python_version(), "sorted": out}, sys.stdout)
