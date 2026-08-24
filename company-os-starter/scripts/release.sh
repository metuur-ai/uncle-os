#!/usr/bin/env bash
# release.sh — cut a release: bump the literals, gate, commit, tag, push.
#
#   ./scripts/release.sh v1.0.0     # explicit version
#   ./scripts/release.sh minor      # or major|minor|patch
#   ./scripts/release.sh v1.0.0 --dry-run
#
# What this does NOT do: build or upload anything. Pushing a `v*` tag triggers
# .github/workflows/release.yml, which runs the gate again on a clean checkout,
# builds the cross-compiled artifacts, and publishes them with SHA256SUMS. CI is
# the only publisher — a human running `make release && gh release create` by
# hand is how a release ends up not matching its own tag.
#
# The gate runs locally BEFORE the tag is pushed. CI running it afterwards is
# too late: a tag is permanent, and a tag whose workflow failed is an install
# line that 404s while looking live.
set -euo pipefail

ROOT="$(git rev-parse --show-toplevel)"
cd "$ROOT/company-os-starter"

BRANCH=main
REMOTE=origin
DRY_RUN=0
ASSUME_YES=0
SKIP_GATE=0
ARG=""

bold() { printf '\033[1m%s\033[0m\n' "$*"; }
info() { printf '  %s\n' "$*"; }
ok()   { printf '\033[32m  %s\033[0m\n' "$*"; }
warn() { printf '\033[33m  %s\033[0m\n' "$*"; }
die()  { printf '\033[31mrelease: %s\033[0m\n' "$*" >&2; exit 1; }

for a in "$@"; do
  case "$a" in
    --dry-run)   DRY_RUN=1 ;;
    -y|--yes)    ASSUME_YES=1 ;;
    --skip-gate) SKIP_GATE=1 ;;
    # Print the header block: everything after the shebang up to the first
    # non-comment line. awk, not `sed -n '2,Np'`, so edits to the header cannot
    # silently start leaking code into --help.
    -h|--help)   awk 'NR>1 && !/^#/{exit} NR>1{sub(/^# ?/,""); print}' "$0"; exit 0 ;;
    -*)          die "unrecognized flag '$a'" ;;
    *)           [[ -z "$ARG" ]] || die "expected one version argument, got '$ARG' and '$a'"; ARG="$a" ;;
  esac
done
[[ -n "$ARG" ]] || die "expected a version: v1.2.0, or major|minor|patch"

# The bump touches tracked files; a dirty tree makes "what shipped" unanswerable
# and makes the revert-on-failure below unsafe.
# Safe because preflight proved the tree was clean: the only tracked changes
# that can exist at this point are the bump's own.
restore() { git -C "$ROOT" checkout --quiet -- . 2>/dev/null || true; }

bold "Preflight"
git diff --quiet && git diff --cached --quiet \
  || die "working tree is dirty; commit or stash first"

cur_branch="$(git rev-parse --abbrev-ref HEAD)"
[[ "$cur_branch" == "$BRANCH" ]] \
  || die "on '$cur_branch'; releases are cut from '$BRANCH'"

git fetch --quiet "$REMOTE" "$BRANCH" --tags || die "cannot reach $REMOTE"
local_head="$(git rev-parse HEAD)"
remote_head="$(git rev-parse "$REMOTE/$BRANCH")"
[[ "$local_head" == "$remote_head" ]] \
  || die "$BRANCH is out of sync with $REMOTE/$BRANCH; pull/push first"
ok "clean tree, on $BRANCH, in sync with $REMOTE"

# ── Bump ──────────────────────────────────────────────────────────────────────
# Done before the tag-collision check because major|minor|patch is only resolved
# to a concrete version by the bump itself. Anything that fails from here on
# restores the working tree.
bold "Bumping release-tag literals"
./scripts/bump-version.sh "$ARG" | sed 's/^/  /'

VERSION="$(./scripts/bump-version.sh --check | awk '{print $NF}')" || {
  restore; die "sites still disagree after the bump"
}
[[ "$VERSION" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]] || { restore; die "could not resolve the new version"; }

if git rev-parse -q --verify "refs/tags/$VERSION" >/dev/null; then
  restore; die "tag $VERSION already exists locally"
fi
if [[ -n "$(git ls-remote --tags "$REMOTE" "refs/tags/$VERSION")" ]]; then
  restore; die "tag $VERSION already published on $REMOTE — releases are immutable; bump again"
fi
ok "$VERSION is unused on $REMOTE"

# ── Gate ──────────────────────────────────────────────────────────────────────
if [[ $SKIP_GATE -eq 1 ]]; then
  warn "gate skipped (--skip-gate) — CI will still run it, after the tag is permanent"
else
  bold "Gate (make check)"
  if ! make --no-print-directory check >/tmp/release-check.log 2>&1; then
    tail -30 /tmp/release-check.log >&2
    restore
    die "gate failed; nothing tagged or pushed (full log: /tmp/release-check.log)"
  fi
  ok "make check passed"
fi

# ── Confirm ───────────────────────────────────────────────────────────────────
bold "About to publish $VERSION"
info "commit  release: $VERSION  (on $BRANCH)"
info "tag     $VERSION (annotated)"
info "push    $REMOTE $BRANCH + $VERSION  -> triggers the release workflow"
git -C "$ROOT" diff --stat | sed 's/^/  /'

if [[ $DRY_RUN -eq 1 ]]; then
  restore
  warn "dry run — working tree restored, nothing committed or pushed"
  exit 0
fi

if [[ $ASSUME_YES -ne 1 ]]; then
  read -r -p "  Push $VERSION to $REMOTE? [y/N] " reply
  case "$reply" in
    y|Y|yes) ;;
    *) restore; die "aborted; working tree restored" ;;
  esac
fi

# ── Publish ───────────────────────────────────────────────────────────────────
# Push the branch first. A tag whose commit is not on the remote branch is a
# release nobody can check out.
bold "Publishing"
git -C "$ROOT" add -A
git -C "$ROOT" commit -q -m "release: $VERSION"
git -C "$ROOT" tag -a "$VERSION" -m "release: $VERSION"
git -C "$ROOT" push --quiet "$REMOTE" "$BRANCH"
git -C "$ROOT" push --quiet "$REMOTE" "$VERSION"
ok "pushed $VERSION"

slug="$(git remote get-url "$REMOTE" | sed -E 's#(git@github.com:|https://github.com/)##; s#\.git$##')"
printf '\n'
info "CI is building the artifacts now:"
info "  https://github.com/$slug/actions/workflows/release.yml"
printf '\n'
info "Next: watch the run, then verify the published release installs:"
info "  gh run watch --repo $slug"
info "  curl -fsSL https://raw.githubusercontent.com/$slug/$VERSION/company-os-starter/install.sh | bash"
