#!/usr/bin/env bash
# Characterization harness (Unit 0) — the safety net for a repo with no test
# suite. Freezes `validate` behavior against a golden snapshot, proves the
# generated artifacts are idempotent (double graph build is a no-op), and runs
# the helper-level selftest. Pass --update to refresh the golden snapshot after
# an intentional change.
#
#   examples/acceptance.sh            # verify (CI gate)
#   examples/acceptance.sh --update   # re-baseline the golden snapshot
set -uo pipefail
HERE="$(cd "$(dirname "$0")" && pwd)"
CLI="$HERE/../bin/company-os"
GOLDEN="$HERE/golden-validate.txt"
TMP="$(mktemp)"
fail=0

# validate prints the absolute workspace path on line 1 — normalize it so the
# golden snapshot is portable across checkouts.
normalize() { sed "s#^validating workspace .*#validating workspace <WORKSPACE>#"; }

# 1. helper selftest (canonical compare, preserve-unknown, fail-safe rewrite)
echo "== selftest =="
python3 "$HERE/selftest.py" || fail=1

# 2. golden validate snapshot on examples/workspace (R-0.7)
echo "== golden validate snapshot (examples/workspace) =="
"$CLI" --root "$HERE/workspace" validate 2>&1 | normalize > "$TMP"
if [ "${1:-}" = "--update" ]; then
  cp "$TMP" "$GOLDEN"; echo "golden snapshot updated -> ${GOLDEN}"
fi
if [ ! -f "$GOLDEN" ]; then
  echo "FAIL: no golden snapshot; run: examples/acceptance.sh --update"; fail=1
elif diff -u "$GOLDEN" "$TMP"; then
  echo "ok: validate matches golden snapshot"
else
  echo "FAIL: validate output drifted from golden snapshot"; fail=1
fi

# 3. validate exits 0 on the example workspace AND the standalone-team fixture
echo "== validate exit code (both workspaces) =="
for W in workspace standalone-team; do
  if "$CLI" --root "$HERE/$W" validate >/dev/null 2>&1; then
    echo "ok: validate exit 0 ($W)"
  else
    echo "FAIL: validate non-zero ($W)"; fail=1
  fi
done

# 4. double graph build is a no-op — committed generated state is already fully
#    derived AND the builder is idempotent (R-0.6). Git-free: compare a content
#    checksum of the workspace before/after two builds (s0 == s1 == s2).
echo "== double-build no-op (both workspaces) =="
snapshot() { find "$1" -type f ! -name '*.pyc' -print0 | sort -z | xargs -0 shasum; }
for W in workspace standalone-team; do
  s0="$(snapshot "$HERE/$W")"
  "$CLI" --root "$HERE/$W" graph build >/dev/null 2>&1
  s1="$(snapshot "$HERE/$W")"
  "$CLI" --root "$HERE/$W" graph build >/dev/null 2>&1
  s2="$(snapshot "$HERE/$W")"
  if [ "$s0" = "$s1" ] && [ "$s1" = "$s2" ]; then
    echo "ok: committed state fully derived + idempotent ($W)"
  else
    echo "FAIL: graph build changed committed state or is not idempotent ($W)"; fail=1
  fi
done

rm -f "$TMP"
if [ "$fail" = 0 ]; then echo "ACCEPTANCE: PASS"; else echo "ACCEPTANCE: FAIL"; exit 1; fi
