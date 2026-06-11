#!/usr/bin/env bash
set -euo pipefail

# ── config ────────────────────────────────────────────────────────────────────
GITHUB_USER="mahi160"
GITHUB_REPO="flowd"
INSTALL_DIR="${FLOWD_INSTALL_DIR:-/usr/local/bin}"
BINARY="fw"
# ─────────────────────────────────────────────────────────────────────────────

RED='\033[0;31m'; GREEN='\033[0;32m'; CYAN='\033[0;36m'; BOLD='\033[1m'; RESET='\033[0m'

info() { echo -e "${CYAN}→${RESET} $*"; }
ok()   { echo -e "${GREEN}✓${RESET} $*"; }
die()  { echo -e "${RED}✗ $*${RESET}" >&2; exit 1; }

echo -e "${BOLD}flowd installer${RESET}"
echo "────────────────────────────────"

# ── platform ──────────────────────────────────────────────────────────────────
OS="$(uname -s)"
ARCH="$(uname -m)"
info "platform: $OS / $ARCH"

case "$OS" in
  Linux)  GOOS="linux"  ;;
  Darwin) GOOS="darwin" ;;
  *) die "unsupported OS: $OS" ;;
esac

case "$ARCH" in
  x86_64)          GOARCH="amd64" ;;
  aarch64 | arm64) GOARCH="arm64" ;;
  *) die "unsupported architecture: $ARCH" ;;
esac

# ── dependency check ──────────────────────────────────────────────────────────
check() { command -v "$1" &>/dev/null || die "$1 not found — $2"; }
check tmux "install tmux via your package manager"
check git  "install git via your package manager"

# ── resolve latest release tag ────────────────────────────────────────────────
info "resolving latest release..."
if command -v curl &>/dev/null; then
  FETCH() { curl -fsSL "$1"; }
elif command -v wget &>/dev/null; then
  FETCH() { wget -qO- "$1"; }
else
  die "curl or wget required"
fi

API_URL="https://api.github.com/repos/${GITHUB_USER}/${GITHUB_REPO}/releases/latest"
TAG=$(FETCH "$API_URL" | grep '"tag_name"' | sed -E 's/.*"tag_name": *"([^"]+)".*/\1/')
[[ -n "$TAG" ]] || die "could not determine latest release tag"
ok "latest release: $TAG"

# ── download binary ───────────────────────────────────────────────────────────
ASSET_NAME="${BINARY}-${GOOS}-${GOARCH}"
DOWNLOAD_URL="https://github.com/${GITHUB_USER}/${GITHUB_REPO}/releases/download/${TAG}/${ASSET_NAME}"
TMP_BIN="$(mktemp)"
trap 'rm -f "$TMP_BIN"' EXIT

info "downloading $ASSET_NAME..."
FETCH "$DOWNLOAD_URL" > "$TMP_BIN"
chmod +x "$TMP_BIN"
ok "downloaded"

# ── install ───────────────────────────────────────────────────────────────────
DEST="${INSTALL_DIR}/${BINARY}"
if [ -w "$INSTALL_DIR" ]; then
  mv "$TMP_BIN" "$DEST"
else
  info "need sudo to write to $INSTALL_DIR"
  sudo mv "$TMP_BIN" "$DEST"
fi
ok "installed → $DEST"

# ── PATH check ────────────────────────────────────────────────────────────────
if ! command -v "$BINARY" &>/dev/null; then
  echo ""
  echo -e "${RED}warning:${RESET} $INSTALL_DIR is not in your PATH."
  echo "  add to your shell profile:"
  echo "    export PATH=\"$INSTALL_DIR:\$PATH\""
fi

# ── done ──────────────────────────────────────────────────────────────────────
echo ""
echo -e "${BOLD}done.${RESET} installed ${TAG}"
echo ""
if [ -f "${HOME}/.config/flowd/config.yaml" ]; then
  ok "existing config found — skipping init"
  echo "  run: fw start"
else
  echo "  run: fw init"
  echo "       fw start"
fi
