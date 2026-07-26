#!/usr/bin/env bash
# Characterization harness (Unit 0) — the safety net for a repo with no test
# suite. Freezes `validate` behavior against golden snapshots — on the passing
# path AND on the failure path (R-0.9) — and proves the generated artifacts are
# idempotent (double graph build is a no-op). Pass --update to refresh every
# golden snapshot after an intentional change.
#
#   examples/acceptance.sh            # verify (CI gate)
#   examples/acceptance.sh --update   # re-baseline the golden snapshot
#
# The helper-level `selftest.py` step is gone: R-9.3 deleted it at cutover, and
# all 85 of its assertions live in `go test ./...` (see
# .devlocal/go-port/selftest-inventory.md for the per-assertion mapping).
# `make check` runs both halves.
set -uo pipefail
HERE="$(cd "$(dirname "$0")" && pwd)"
# The Go binary. `make build` produces it; build it first if absent, so this
# script keeps working as a standalone CI gate.
CLI="$HERE/../company-os-starter/company-os"
if [ ! -x "$CLI" ]; then
  echo "== building the CLI =="
  ( cd "$HERE/../company-os-starter" && make --no-print-directory build ) || exit 1
fi
GOLDEN="$HERE/golden-validate.txt"
FED_GOLDEN="$HERE/federated-golden-validate.txt"
TMP="$(mktemp)"
fail=0

# validate prints the absolute workspace path on line 1 — normalize it so the
# golden snapshot is portable across checkouts.
normalize() { sed "s#^validating workspace .*#validating workspace <WORKSPACE>#"; }

# 1. golden validate snapshot on examples/workspace (R-0.7)
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

# 2b. federated golden snapshot on examples/federated (Phase 5) — the committed
#     self-contained federated fixture (manifest + lock + read-only slices) must
#     validate green [8/8] on a fresh checkout with NO network and NO source repo
#     (gate 8 compares committed slice bytes to committed lock hashes). Kept in a
#     SEPARATE snapshot so the monorepo golden above stays byte-for-byte frozen.
echo "== federated golden snapshot (examples/federated) =="
"$CLI" --root "$HERE/federated" validate 2>&1 | normalize > "$TMP"
if [ "${1:-}" = "--update" ]; then
  cp "$TMP" "$FED_GOLDEN"; echo "federated snapshot updated -> ${FED_GOLDEN}"
fi
if [ ! -f "$FED_GOLDEN" ]; then
  echo "FAIL: no federated snapshot; run: examples/acceptance.sh --update"; fail=1
elif diff -u "$FED_GOLDEN" "$TMP"; then
  echo "ok: federated validate matches golden snapshot"
else
  echo "FAIL: federated validate output drifted from golden snapshot"; fail=1
fi

# 2c. failure-path golden snapshots (R-0.9). The two snapshots above cover only
#     the all-pass path — between them they contain zero `[FAIL]` and zero
#     `[warn]` lines, against 15 failure sites in cmd_validate. These three
#     fixtures are DELIBERATELY BROKEN and exit 1; together they drive at least
#     one [FAIL] through every gate, the single warn() site, and gate 4's
#     conditional `[ok]` (a doc with core-field errors emits its failures and NO
#     ok line). Captured from the Python CLI while it still exists, because the
#     per-gate line prefix is not uniform and a records refactor would otherwise
#     break it silently.
echo "== failure-path golden snapshots =="
for W in failing-workspace failing-federated failing-federated-nolock; do
  G="$HERE/$W-golden-validate.txt"
  "$CLI" --root "$HERE/$W" validate 2>&1 | normalize > "$TMP"
  rc="${PIPESTATUS[0]}"
  if [ "${1:-}" = "--update" ]; then
    cp "$TMP" "$G"; echo "failure snapshot updated -> ${G}"
  fi
  if [ ! -f "$G" ]; then
    echo "FAIL: no failure snapshot ($W); run: examples/acceptance.sh --update"; fail=1
  elif diff -u "$G" "$TMP"; then
    echo "ok: validate matches failure snapshot ($W)"
  else
    echo "FAIL: validate output drifted from failure snapshot ($W)"; fail=1
  fi
  # These fixtures must keep FAILING, and with the documented exit code — a
  # fixture that silently starts passing would still diff clean against a
  # re-baselined golden, so the exit code is asserted separately from the diff.
  if [ "$rc" = 1 ]; then
    echo "ok: validate exit 1 ($W)"
  else
    echo "FAIL: validate exit $rc, expected 1 ($W)"; fail=1
  fi
done

# 3. validate exits 0 on every committed fixture: the example workspace, the
#    standalone-team fixture, AND the federated fixture (offline, from lock).
#    The failing-* fixtures are excluded on purpose — they are asserted to exit
#    1 in section 2c above.
echo "== validate exit code (all fixtures) =="
for W in workspace standalone-team federated; do
  if "$CLI" --root "$HERE/$W" validate >/dev/null 2>&1; then
    echo "ok: validate exit 0 ($W)"
  else
    echo "FAIL: validate non-zero ($W)"; fail=1
  fi
done

# 4. double graph build is a no-op — committed generated state is already fully
#    derived AND the builder is idempotent (R-0.6). Git-free: compare a content
#    checksum of the workspace before/after two builds (s0 == s1 == s2).
#    NOTE: `federated` is excluded on purpose — its governance-only slices are
#    read-only derived content (no generated/ or CLAUDE nodes materialized), so
#    running graph build there would create derived artifacts and mutate the
#    committed slice; slice integrity is covered by gate [8/8] + section 5 below.
echo "== double-build no-op (monorepo fixtures) =="
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

# 5. federation reproducibility (build-at-test-time integration check). Proves
#    the sync guarantee end to end: a sparse governance-only fetch from a real
#    git repo, materialized read-only slices that contain ONLY allowlisted paths,
#    and a `--frozen` offline re-materialization that reproduces the exact same
#    slice bytes from the lock. Skips cleanly when git is absent or < 2.27 so the
#    gate never blocks a machine that only runs monorepo mode (GPF-R-7.7).
echo "== federation reproducibility (sync + --frozen) =="
git_ok=0
if command -v git >/dev/null 2>&1; then
  gv="$(git --version 2>/dev/null | grep -oE '[0-9]+\.[0-9]+' | head -1)"
  gmaj="${gv%%.*}"; gmin="${gv##*.}"
  if [ "${gmaj:-0}" -gt 2 ] || { [ "${gmaj:-0}" -eq 2 ] && [ "${gmin:-0}" -ge 27 ]; }; then
    git_ok=1
  fi
fi
if [ "$git_ok" != 1 ]; then
  echo "skip: git absent or < 2.27 (federation sync needs cone-mode sparse-checkout)"
else
  IWORK="$(mktemp -d)"
  SRC="$IWORK/src-testplat"; WS="$IWORK/ws"
  # Source repo: governance content (allowlisted) PLUS non-governance files that
  # the sparse allowlist must NOT pull (README.md at root, src/ subtree).
  mkdir -p "$SRC/governance" "$SRC/components" "$SRC/src" "$SRC/docs/sdd"
  echo "# Spec" > "$SRC/docs/sdd/spec.md"
  cat > "$SRC/governance/requirements.yaml" <<'YAML'
version: 1
platform: testplat
requirements: []
YAML
  cat > "$SRC/components/foo.yaml" <<'YAML'
id: foo
componentType: service
ownership:
  accountableTeam: team://none
YAML
  echo "# not governance — must not be sliced" > "$SRC/README.md"
  echo "print('not governance')" > "$SRC/src/app.py"
  ( cd "$SRC" && git init -q && git config user.email t@e && git config user.name t \
      && git add -A && git commit -q -m init ) || fail=1
  ISHA="$( cd "$SRC" && git rev-parse HEAD )"
  # Workspace: just a manifest (marks the root). ONE repo, TWO destinations —
  # a governance slice under platforms/ and a knowledge slice under knowledge/.
  # The knowledge entry is deliberately NESTED (docs/sdd): depth-1 entries never
  # exercised the parent-permission path that a re-sync needs.
  mkdir -p "$WS"
  cat > "$WS/workspace.yaml" <<YAML
version: 1
repos:
  - name: testplat
    url: file://$SRC
    pin:
      commit: $ISHA
    slices:
      - localDirectory: platforms/testplat
        paths: [governance/, components/]
      - localDirectory: knowledge/testplat
        paths: [docs/sdd]
YAML
  SLICE="$WS/platforms/testplat"
  KSLICE="$WS/knowledge/testplat"
  tree_hash() { find "$1" -type f ! -name '*.pyc' -print0 | sort -z | xargs -0 shasum | shasum | awk '{print $1}'; }

  if "$CLI" --root "$WS" workspace sync >/dev/null 2>&1; then
    # governance-only: allowlisted present, non-allowlisted absent
    if [ -f "$SLICE/governance/requirements.yaml" ] && [ -f "$SLICE/components/foo.yaml" ] \
       && [ ! -e "$SLICE/README.md" ] && [ ! -e "$SLICE/src" ]; then
      echo "ok: slice is governance-only (README.md and src/ not pulled)"
    else
      echo "FAIL: slice leaked non-allowlisted paths"; fail=1
    fi
    # multi-slice: the second target holds its own content and only its own
    if [ -f "$KSLICE/docs/sdd/spec.md" ] && [ ! -e "$KSLICE/governance" ] \
       && [ ! -e "$SLICE/docs" ]; then
      echo "ok: multi-slice targets are populated and isolated"
    else
      echo "FAIL: multi-slice targets leaked into each other"; fail=1
    fi
    # one repo => exactly one cache dir, one fetch
    if [ "$(ls "$WS/.company-os/federation-cache" | wc -l | tr -d ' ')" = "1" ]; then
      echo "ok: multi-slice repo uses a single git cache"
    else
      echo "FAIL: multi-slice repo produced more than one cache dir"; fail=1
    fi
    # the lock must carry slices: (and no top-level localDirectory:) — nothing
    # READS that key at runtime, so an un-migrated lock would otherwise stay green
    if grep -q "slices:" "$WS/workspace.lock.yaml" \
       && ! grep -qE "^  localDirectory:" "$WS/workspace.lock.yaml"; then
      echo "ok: lock records the slice list"
    else
      echo "FAIL: lock is missing slices: or still has a top-level localDirectory:"; fail=1
    fi
    # re-sync over the now read-only tree — regression guard for the nested entry
    if "$CLI" --root "$WS" workspace sync >/dev/null 2>&1; then
      echo "ok: re-sync over a read-only nested slice succeeds"
    else
      echo "FAIL: re-sync over a read-only nested slice failed"; fail=1
    fi
    # status must render every target and report clean (nothing exercises it otherwise)
    if "$CLI" --root "$WS" workspace status 2>&1 | grep -q "platforms/testplat, knowledge/testplat — clean"; then
      echo "ok: status lists every slice target and reports clean"
    else
      echo "FAIL: workspace status did not report all targets clean"; fail=1
    fi
    h_online="$(tree_hash "$WS/platforms")$(tree_hash "$WS/knowledge")"
    # --frozen re-materializes strictly from the lock (offline contract).
    if "$CLI" --root "$WS" workspace sync --frozen >/dev/null 2>&1; then
      h_frozen="$(tree_hash "$WS/platforms")$(tree_hash "$WS/knowledge")"
      if [ "$h_online" = "$h_frozen" ]; then
        echo "ok: --frozen reproduces slice bytes from lock ($h_frozen)"
      else
        echo "FAIL: --frozen slice differs from online sync ($h_online != $h_frozen)"; fail=1
      fi
    else
      echo "FAIL: workspace sync --frozen failed"; fail=1
    fi
  else
    echo "FAIL: workspace sync failed"; fail=1
  fi
  # Slices are read-only (0555 dirs) — make writable before removing.
  chmod -R u+w "$IWORK" 2>/dev/null
  rm -rf "$IWORK"
fi

rm -f "$TMP"
if [ "$fail" = 0 ]; then echo "ACCEPTANCE: PASS"; else echo "ACCEPTANCE: FAIL"; exit 1; fi
