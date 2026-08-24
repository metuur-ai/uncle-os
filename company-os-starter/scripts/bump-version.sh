#!/usr/bin/env bash
#
# bump-version.sh — keep the hardcoded release-tag literals in sync.
#
# The binary versions itself from the git tag at build time (-X .../model.version),
# so there is no VERSION file. But the release tag is also written by hand into a
# handful of docs (CI pinning examples) and into install.sh's usage comment. Those
# copies drift silently. This script is the one place that knows where they live.
#
# It does NOT touch per-artifact `version:` fields in requirements.yaml / SKILL.md.
# Those are independent content versions: `@spec req://...@<version>#<clause>`
# markers bind code to an exact requirement version and `supersedes:` chains
# encode their own history. Bumping those in lockstep would orphan both.
#
# Usage:
#   bump-version.sh v1.2.0        set every site to an explicit tag
#   bump-version.sh major|minor|patch
#                                 increment from the current (unanimous) value
#   bump-version.sh --check       verify every site agrees; exit 1 on drift
#   bump-version.sh --show        print the current value and each site
#
set -euo pipefail

# ── Sites ─────────────────────────────────────────────────────────────────────
# "<path relative to repo root>|<literal prefix that precedes the vX.Y.Z>"
# Add a line here when a new hardcoded copy of the release tag appears.
SITES=(
  "company-os-starter/docs/TUTORIAL.md|COMPANY_OS_VERSION: "
  "company-os-starter/docs/FEDERATION-RUNBOOK.md|COMPANY_OS_VERSION: "
  "company-os-starter/docs/user-guide/how-to/run-the-validation-gate.md|COMPANY_OS_VERSION: "
  "company-os-starter/docs/user-guide/how-to/sync-a-knowledge-catalog.md|COMPANY_OS_VERSION: "
  "TUTORIAL.md|COMPANY_OS_VERSION: "
  "company-os-starter/install.sh|VERSION="
)

SEMVER='v[0-9]+\.[0-9]+\.[0-9]+'

die() { printf 'bump-version: %s\n' "$*" >&2; exit 1; }

cd "$(git rev-parse --show-toplevel)" || die "not inside a git repository"

# ── Collect every occurrence as "<file>:<line>:<version>" ─────────────────────
collect() {
  local site file prefix
  for site in "${SITES[@]}"; do
    file="${site%%|*}"
    prefix="${site#*|}"
    [[ -f "$file" ]] || die "missing site file: $file"
    grep -nE "${prefix}${SEMVER}" "$file" 2>/dev/null | while IFS=: read -r lineno rest; do
      printf '%s:%s:%s\n' "$file" "$lineno" \
        "$(printf '%s' "$rest" | grep -oE "${prefix}${SEMVER}" | grep -oE "${SEMVER}" | head -1)"
    done || true
  done
}

# Distinct versions currently in use.
distinct() { collect | cut -d: -f3 | sort -u; }

report() {
  local line
  while IFS= read -r line; do
    printf '  %-70s %s\n' "${line%:*}" "${line##*:}"
  done < <(collect)
}

# The single current version, or empty if the sites disagree.
current() {
  local vals
  vals="$(distinct)"
  [[ "$(printf '%s\n' "$vals" | grep -c .)" -eq 1 ]] && printf '%s' "$vals" || printf ''
}

# ── Rewrite every site ────────────────────────────────────────────────────────
apply() {
  local new="$1" site file prefix tmp changed=0
  for site in "${SITES[@]}"; do
    file="${site%%|*}"
    prefix="${site#*|}"
    tmp="$(mktemp)"
    sed -E "s|(${prefix})${SEMVER}|\1${new}|g" "$file" > "$tmp"
    if cmp -s "$file" "$tmp"; then
      rm -f "$tmp"
    else
      # Write back through the existing inode rather than mv'ing the temp over it,
      # so mode and ownership survive (install.sh must stay executable) without
      # needing chmod --reference / stat -f, which differ between GNU and BSD.
      cat "$tmp" > "$file"
      rm -f "$tmp"
      printf '  updated %s\n' "$file"
      changed=1
    fi
  done
  [[ $changed -eq 1 ]] || printf '  (already at %s, nothing to do)\n' "$new"
}

# ── Modes ─────────────────────────────────────────────────────────────────────
[[ $# -eq 1 ]] || die "expected exactly one argument; try --check, --show, a tag like v1.2.0, or major|minor|patch"
arg="$1"

case "$arg" in
  --show)
    cur="$(current)"
    printf 'Release-tag sites:\n'; report
    printf '\nLatest git tag: %s\n' "$(git tag --list 'v*' --sort=-v:refname | head -1 || echo '(none)')"
    if [[ -n "$cur" ]]; then printf 'Current: %s (all sites agree)\n' "$cur"
    else printf 'Current: DRIFTED — sites disagree: %s\n' "$(distinct | tr '\n' ' ')"; fi
    ;;

  --check)
    cur="$(current)"
    if [[ -n "$cur" ]]; then
      printf 'ok: all %d release-tag sites at %s\n' "$(collect | grep -c .)" "$cur"
      exit 0
    fi
    printf 'FAIL: release-tag literals have drifted apart.\n\n' >&2
    report >&2
    printf '\nDistinct values: %s\n' "$(distinct | tr '\n' ' ')" >&2
    printf 'Reconcile with: %s <version>\n' "$0" >&2
    exit 1
    ;;

  major|minor|patch)
    cur="$(current)"
    [[ -n "$cur" ]] || die "sites disagree ($(distinct | tr '\n' ' ')); pass an explicit version to reconcile them first"
    IFS=. read -r maj min pat <<<"${cur#v}"
    case "$arg" in
      major) maj=$((maj + 1)); min=0; pat=0 ;;
      minor) min=$((min + 1)); pat=0 ;;
      patch) pat=$((pat + 1)) ;;
    esac
    new="v${maj}.${min}.${pat}"
    printf 'Bumping %s -> %s (%s)\n' "$cur" "$new" "$arg"
    apply "$new"
    ;;

  v[0-9]*)
    [[ "$arg" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]] || die "invalid version '$arg'; expected vMAJOR.MINOR.PATCH"
    cur="$(current)"
    if [[ -n "$cur" ]]; then printf 'Setting %s -> %s\n' "$cur" "$arg"
    else printf 'Reconciling drifted sites (%s) -> %s\n' "$(distinct | tr '\n' ' ')" "$arg"; fi
    apply "$arg"
    ;;

  *)
    die "unrecognized argument '$arg'; try --check, --show, a tag like v1.2.0, or major|minor|patch"
    ;;
esac
