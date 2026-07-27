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
# Options (env):
#   INSTALL_DIR=/custom/bin        binary location            (default ~/.local/bin)
#   VERSION=v0.4.0                 release tag                (default: latest)
#   BASE_URL=https://...           override the download base
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

# resolve_binary — echo a path to the platform binary, downloading if needed.
resolve_binary() {
  local plat="$1" name="$TOOL_NAME-$1" p
  # A checkout or unpacked release: prefer what is already on disk.
  if [[ -n "$SCRIPT_DIR" ]]; then
    for p in "$SCRIPT_DIR/dist/$name" "$SCRIPT_DIR/$TOOL_NAME"; do
      if [[ -f "$p" ]]; then echo "$p"; return; fi
    done
  fi
  local tmp
  tmp="$(mktemp -d)/$TOOL_NAME"
  download "$BASE_URL/$name" "$tmp"
  echo "$tmp"
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
  info "company-os validate                # the CI gate"
  info "company-os tui                     # interactive, needs a real terminal"
  echo
  info "No workspace yet?  mkdir my-os && cd my-os && company-os init"
}

main "$@"
