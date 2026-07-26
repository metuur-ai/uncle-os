#!/usr/bin/env bash
# Task 6.5 — delete the Python reference implementation (R-9.3).
#
# THIS IS THE IRREVERSIBLE STEP. Run it yourself; an agent should not.
# Recovery is `git checkout python-cli-final -- <three paths>` (see below), but
# the cross-implementation oracle is gone the moment this runs: examples/
# differential.py needs two binaries and after this there is one.
#
#   company-os-starter/scripts/cutover.sh --dry-run   # show what would go
#   company-os-starter/scripts/cutover.sh             # do it, then verify
#
# Preconditions, all re-checked below rather than assumed:
#   - the parity gate (6.4) is green
#   - the tag python-cli-final exists and contains a runnable Python CLI
set -uo pipefail
HERE="$(cd "$(dirname "$0")" && pwd)"
ROOT="$(cd "$HERE/../.." && pwd)"
KIT="$ROOT/company-os-starter"
DRY=0
[ "${1:-}" = "--dry-run" ] && DRY=1

TARGETS=(
  "company-os-starter/bin"            # the 2781-line Python CLI
  "company-os-starter/install.sh"     # installer, and the documented uninstaller
  "company-os-starter/vendor"         # 536K vendored PyYAML
  "company-os-starter/go.work"        # only existed to stop Go reading vendor/
  "examples/selftest.py"              # 85 assertions, all ported to go test
)

cd "$ROOT" || exit 1
fail=0

# --- 1. the recovery point must exist and must actually work -----------------
echo "== recovery point =="
if ! git rev-parse -q --verify refs/tags/python-cli-final >/dev/null; then
  echo "FAIL: tag python-cli-final does not exist. Task 6.3 must run first —"
  echo "      without it there is no way back to a runnable reference."
  exit 1
fi
for p in company-os-starter/bin/company-os company-os-starter/vendor \
         company-os-starter/templates examples/differential.py \
         examples/declared-divergences.txt; do
  git cat-file -e "python-cli-final:$p" 2>/dev/null \
    || { echo "FAIL: $p missing from the tag"; fail=1; }
done
[ $fail -eq 0 ] && echo "ok: tag carries the CLI, its vendored deps, its templates, and the harness"

# --- 2. the parity gate must be green ---------------------------------------
echo
echo "== parity gate (R-7.9) =="
( cd "$KIT" && go test ./... >/dev/null 2>&1 ) \
  && echo "ok: go test" || { echo "FAIL: go test"; fail=1; }
if [ -x "$KIT/bin/company-os" ]; then
  ( cd "$ROOT" && PYTHONPATH="$KIT/vendor" python3 examples/differential.py \
      "$KIT/bin/company-os" "$KIT/company-os" >/dev/null 2>&1 ) \
    && echo "ok: differential, zero undeclared divergence" \
    || { echo "FAIL: differential — do NOT cut over"; fail=1; }
else
  echo "skip: differential (Python CLI already gone)"
fi

if [ $fail -ne 0 ]; then
  echo
  echo "ABORTED: preconditions failed. Nothing was deleted."
  exit 1
fi

# --- 3. delete ---------------------------------------------------------------
echo
if [ $DRY -eq 1 ]; then
  echo "== would delete (dry run) =="
  for t in "${TARGETS[@]}"; do printf '  %s\n' "$t"; done
  echo
  echo "Re-run without --dry-run to proceed."
  exit 0
fi

echo "== deleting =="
for t in "${TARGETS[@]}"; do
  git rm -r -q --ignore-unmatch "$t" && printf '  removed %s\n' "$t"
done

# --- 4. verify the tree still stands -----------------------------------------
echo
echo "== post-cutover verification =="
( cd "$KIT" && make --no-print-directory build >/dev/null ) \
  && echo "ok: build" || { echo "FAIL: build"; fail=1; }
( cd "$KIT" && go test ./... >/dev/null 2>&1 ) \
  && echo "ok: go test" || { echo "FAIL: go test"; fail=1; }
"$ROOT/examples/acceptance.sh" >/dev/null 2>&1 \
  && echo "ok: acceptance.sh" || { echo "FAIL: acceptance.sh"; fail=1; }

echo
if [ $fail -eq 0 ]; then
  cat <<'DONE'
CUTOVER COMPLETE. Nothing is committed — review `git status`, then commit.

examples/differential.py is now inert: it compares two binaries and there is one.
That is by design, and it is also the end of the oracle that caught the YAML
1.1/1.2 scalar divergence, the insertion-order lock bug, and the prd complete
archive defect. From here, `go test` and the golden snapshots are the safety net.

To recover the Python CLI you need three paths, not one — bin/company-os:36
resolves TEMPLATES_DIR at import, so a bin-only checkout passes validate while
every scaffolding command fails:

  git checkout python-cli-final -- company-os-starter/bin/company-os \
                                   company-os-starter/vendor \
                                   company-os-starter/templates
DONE
else
  echo "CUTOVER LEFT THE TREE RED. Nothing is committed — 'git checkout -- .'"
  echo "and 'git checkout python-cli-final -- <paths>' restore it."
  exit 1
fi
