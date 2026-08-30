#!/usr/bin/env bash
# install.sh — install the company-os CLI
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/metuur-ai/uncle-os/main/company-os-starter/install.sh | bash
#   or from a checkout / unpacked release:  bash install.sh
#
# What it installs:
#   company-os -> $INSTALL_DIR (default ~/.local/bin)
#
# This script installs the CLI and nothing else. It writes no file into any
# workspace: the canonical agent skills ship inside the binary, and
# `company-os skills install`, run from a workspace root, is what puts them
# on disk.
#
# It always installs the LATEST published release, even when run from a
# checkout — a checkout's dist/ is whatever its owner last cross-compiled and is
# routinely older than what is published. Set LOCAL_BUILD=1 to install your own
# build instead. An unpacked release bundle (install.sh beside the artifacts and
# SHA256SUMS) uses those artifacts, verified against the bundled checksums.
#
# Options (env):
#   INSTALL_DIR=/custom/bin        binary location            (default ~/.local/bin)
#   VERSION=v1.1.2                 release tag                (default: latest)
#   BASE_URL=https://...           override the download base
#   LOCAL_BUILD=1                  install ./dist or ./company-os from a checkout
#
# WHY THIS EXISTS AND NOT JUST A BROWSER DOWNLOAD (R-6.3):
# The binaries are NOT signed and NOT notarized. On macOS that matters only for
# the path this script avoids. `com.apple.quarantine` is set by the *downloading
# application* — Safari, Chrome, Mail — not by curl, wget or tar. A browser
# download of an unsigned binary does not fail loudly on first exec; it HANGS
# with no output while Gatekeeper waits on a verdict that never comes. Fetched
# with curl, the same bytes carry no quarantine attribute and run immediately.
#
# `spctl -a` still reports `rejected` afterwards, and that is expected: spctl
# asks "would Gatekeeper admit this?", which is a different question from
# "will this execute?". Gatekeeper only adjudicates quarantined files.
#
# This is the same approach the local-search CLI ships with, for the same reason.
set -euo pipefail

# ── Config ────────────────────────────────────────────────────────────────────

TOOL_NAME="company-os"
INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"
REPO="${REPO:-metuur-ai/uncle-os}"
VERSION="${VERSION:-latest}"

if [[ "$VERSION" == "latest" ]]; then
  BASE_URL="${BASE_URL:-https://github.com/$REPO/releases/latest/download}"
else
  BASE_URL="${BASE_URL:-https://github.com/$REPO/releases/download/$VERSION}"
fi

# ── Helpers ───────────────────────────────────────────────────────────────────

red()   { printf '\033[31m%s\033[0m\n' "$*"; }
green() { printf '\033[32m%s\033[0m\n' "$*"; }
bold()  { printf '\033[1m%s\033[0m\n'  "$*"; }
info()  { printf '  %s\n' "$*"; }
warn()  { printf '\033[33m  %s\033[0m\n' "$*"; }

die() { red "Error: $*" >&2; exit 1; }

# ensure_on_path <dir> — warn + print how to add <dir> to PATH if it is missing.
ensure_on_path() {
  local dir="$1"
  case ":${PATH}:" in
    *":${dir}:"*) return 0 ;;
  esac
  warn "$dir is not on your PATH — company-os won't be found until you add it."
  info "zsh:  echo 'export PATH=\"$dir:\$PATH\"' >> ~/.zshrc  && source ~/.zshrc"
  info "bash: echo 'export PATH=\"$dir:\$PATH\"' >> ~/.bashrc && source ~/.bashrc"
}

# install_file <src> <dest> — copy with +x, elevating to sudo if the dir is
# unwritable. The copy lands on a sibling path and is mv'd over the target:
# rename(2) is atomic, so there is no window where a half-written binary sits on
# PATH, and it unlinks rather than truncates the old inode — which is what an
# in-place cp cannot do on Linux while the old binary is running (ETXTBSY).
install_file() {
  local src="$1" dest="$2" dir tmp
  dir="$(dirname "$dest")"
  tmp="$dir/.$(basename "$dest").new.$$"
  if { [[ -d "$dir" ]] || mkdir -p "$dir" 2>/dev/null; } && [[ -w "$dir" ]]; then
    cp "$src" "$tmp" && chmod 0755 "$tmp" && mv -f "$tmp" "$dest"
  else
    info "Elevated permissions required for $dir"
    sudo mkdir -p "$dir"
    sudo cp "$src" "$tmp" && sudo chmod 0755 "$tmp" && sudo mv -f "$tmp" "$dest"
  fi
}

# ── Detect platform ───────────────────────────────────────────────────────────

detect_platform() {
  local os arch
  case "$(uname -s)" in
    Darwin) os="darwin" ;;
    Linux)  os="linux"  ;;
    *) die "Unsupported OS: $(uname -s). Build from source: cd company-os-starter && make install" ;;
  esac
  case "$(uname -m)" in
    x86_64|amd64)  arch="amd64" ;;
    arm64|aarch64) arch="arm64" ;;
    *) die "Unsupported architecture: $(uname -m)" ;;
  esac
  echo "${os}-${arch}"
}

# ── Resolve source (local dist/ vs. remote download) ───────────────────────────

# When piped (`curl … | bash`) there is no script file on disk: BASH_SOURCE[0]
# is empty and $0 is "bash". SCRIPT_DIR must stay empty then, so we never
# mistake the current working directory for a checkout — we download instead.
if [[ -n "${BASH_SOURCE[0]:-}" && -f "${BASH_SOURCE[0]}" ]]; then
  SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
else
  SCRIPT_DIR=""
fi

download() {
  local url="$1" dest="$2"
  info "Downloading $url" >&2
  if command -v curl &>/dev/null; then
    curl -fsSL "$url" -o "$dest" || download_failed "$url"
  elif command -v wget &>/dev/null; then
    wget -q "$url" -O "$dest" || download_failed "$url"
  else
    die "Neither curl nor wget found. Install one and retry."
  fi
}

# A 404 here is far more likely to mean "no release has been published yet" than
# "the network is down", and the two need different responses. Say so rather
# than leaving the user to guess from `Download failed`.
download_failed() {
  red "Error: could not download $1" >&2
  echo >&2
  echo "  Most likely no release has been published for this platform yet." >&2
  echo "  Check: https://github.com/$REPO/releases" >&2
  echo >&2
  echo "  To build from source instead (needs the Go toolchain, build-time only):" >&2
  echo "      git clone https://github.com/$REPO" >&2
  echo "      cd uncle-os/company-os-starter && make install" >&2
  exit 1
}

# ── Integrity ─────────────────────────────────────────────────────────────────
# TLS proves who served the bytes, not that they are the bytes `make release`
# produced. SHA256SUMS ships as a release asset next to the binaries; check
# against it before putting anything on PATH.
#
# Scope, stated plainly: this detects a corrupted or swapped asset. It is NOT a
# signature — anyone who can replace the binary in a release can replace
# SHA256SUMS in the same release. Only signing would close that, and this
# project deliberately does not sign (R-6.3).

# sha256_of <file> — echo the hex digest, or return 1 if no hasher exists.
sha256_of() {
  if   command -v sha256sum &>/dev/null; then sha256sum   "$1" | awk '{print $1}'
  elif command -v shasum    &>/dev/null; then shasum -a 256 "$1" | awk '{print $1}'
  else return 1
  fi
}

# fetch_quiet <url> <dest> — like download(), but a failure is the caller's to
# interpret rather than fatal.
fetch_quiet() {
  if   command -v curl &>/dev/null; then curl -fsSL "$1" -o "$2" 2>/dev/null
  elif command -v wget &>/dev/null; then wget -q    "$1" -O "$2" 2>/dev/null
  else return 1
  fi
}

# verify_checksum <file> <asset-name>
#
# A MISMATCH is always fatal. A missing SHA256SUMS (a release predating it, or a
# custom BASE_URL) and a missing hasher are both reported as skipped rather than
# fatal: sha256sum/shasum are effectively universal, and refusing to install on
# the rare host with neither buys no security — a missing coreutils is not an
# attacker. What is never acceptable is claiming a check that did not run.
verify_checksum() {
  local file="$1" name="$2" sums expected actual
  sums="$(mktemp)"

  if ! fetch_quiet "$BASE_URL/SHA256SUMS" "$sums"; then
    rm -f "$sums"
    warn "no SHA256SUMS published at this release — integrity NOT verified"
    return 0
  fi

  # `shasum -a 256 *` writes "<hash>  <name>"; a leading '*' appears in the
  # binary-mode format some tools emit.
  expected="$(awk -v n="$name" '$2 == n || $2 == "*"n {print $1; exit}' "$sums")"
  rm -f "$sums"

  if [[ -z "$expected" ]]; then
    warn "$name absent from SHA256SUMS — integrity NOT verified"
    return 0
  fi

  if ! actual="$(sha256_of "$file")"; then
    warn "no sha256sum/shasum on this system — integrity NOT verified"
    return 0
  fi

  if [[ "$actual" != "$expected" ]]; then
    red "Error: checksum mismatch for $name" >&2
    echo >&2
    echo "  expected  $expected" >&2
    echo "  actual    $actual" >&2
    echo >&2
    echo "  The download does not match the published SHA256SUMS. Not installing." >&2
    echo "  Retry; if it persists, report it at https://github.com/$REPO/issues" >&2
    exit 1
  fi
  info "Verified sha256 ${actual:0:16}…"
}

# resolve_binary — echo a path to the platform binary, downloading if needed.
#
# DOWNLOADING IS THE DEFAULT, including from a checkout. This script's job is to
# install the published release, and a checkout's `dist/` holds whatever its
# owner last cross-compiled — which is routinely older than `latest` and carries
# no relationship to any tag. Preferring it silently installed a months-old
# binary while reporting success, which is the failure this ordering prevents.
#
# The one case where an adjacent binary IS authoritative: an UNPACKED RELEASE.
# `make release` copies install.sh into dist/ beside the artifacts and
# SHA256SUMS, so a sibling artifact there is the release this script shipped
# with, and it can be checksum-verified locally. That is detected by the
# SHA256SUMS file sitting next to the script, never by the presence of a binary.
#
# LOCAL_BUILD=1 opts a developer back into "use my checkout's build" explicitly.
resolve_binary() {
  local plat="$1" name="$TOOL_NAME-$1"

  # 1. Unpacked release bundle: script, artifact and SHA256SUMS side by side.
  if [[ -n "$SCRIPT_DIR" && -f "$SCRIPT_DIR/SHA256SUMS" && -f "$SCRIPT_DIR/$name" ]]; then
    info "Using the artifact bundled beside this script" >&2
    verify_local_checksum "$SCRIPT_DIR/$name" "$name" "$SCRIPT_DIR/SHA256SUMS" >&2
    echo "$SCRIPT_DIR/$name"
    return
  fi

  # 2. Explicit developer opt-in to a local build.
  if [[ -n "${LOCAL_BUILD:-}" && -n "$SCRIPT_DIR" ]]; then
    local p
    for p in "$SCRIPT_DIR/dist/$name" "$SCRIPT_DIR/$TOOL_NAME"; do
      if [[ -f "$p" ]]; then
        warn "LOCAL_BUILD set — installing $p, NOT the published release" >&2
        echo "$p"
        return
      fi
    done
    die "LOCAL_BUILD set but no binary found; run: make build"
  fi

  # 3. The default: fetch the published release.
  local tmp
  tmp="$(mktemp -d)/$TOOL_NAME"
  download "$BASE_URL/$name" "$tmp"
  verify_checksum "$tmp" "$name" >&2
  echo "$tmp"
}

# verify_local_checksum <file> <name> <sums> — the bundled-artifact case.
#
# Split from verify_checksum, which fetches SHA256SUMS over the network: here
# the file is already on disk and downloading a second copy would check the
# bundle against something other than itself.
verify_local_checksum() {
  local file="$1" name="$2" sums="$3" expected actual
  expected="$(awk -v n="$name" '$2 == n || $2 == "*"n {print $1; exit}' "$sums")"
  if [[ -z "$expected" ]]; then
    warn "$name absent from the bundled SHA256SUMS — integrity NOT verified"
    return 0
  fi
  if ! actual="$(sha256_of "$file")"; then
    warn "no sha256sum/shasum on this system — integrity NOT verified"
    return 0
  fi
  if [[ "$actual" != "$expected" ]]; then
    red "Error: checksum mismatch for $name" >&2
    echo >&2
    echo "  expected  $expected" >&2
    echo "  actual    $actual" >&2
    echo >&2
    echo "  The bundled artifact does not match its own SHA256SUMS. Not installing." >&2
    exit 1
  fi
  info "Verified sha256 ${actual:0:16}…"
}

# ── Install ───────────────────────────────────────────────────────────────────

main() {
  bold "Installing $TOOL_NAME"
  local plat path dest="$INSTALL_DIR/$TOOL_NAME"
  plat="$(detect_platform)"
  path="$(resolve_binary "$plat")"

  info "CLI:    $dest"
  install_file "$path" "$dest"

  # Verify by running it. A binary that installed but cannot execute is the
  # failure mode this whole script exists to avoid — do not report success
  # without evidence.
  "$dest" --version &>/dev/null \
    || die "installed but failed to run: $dest"
  green "  installed $("$dest" --version)"

  ensure_on_path "$INSTALL_DIR"

  echo
  bold "Next"
  info "company-os --help                  # the whole surface"
  info "cd <a workspace root>              # or pass --root everywhere"
  info "company-os skills install          # put the canonical agent skills in the workspace"
  info "company-os validate                # the CI gate"
  info "company-os tui                     # interactive, needs a real terminal"
  echo
  info "No workspace yet?  mkdir my-os && cd my-os && company-os init"
}

main "$@"
