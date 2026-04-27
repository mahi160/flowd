#!/usr/bin/env bash
set -euo pipefail

# ── config ────────────────────────────────────────────────────────────────────
GITHUB_USER="mahi160"
GITHUB_REPO="flowd"
BRANCH="main"
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
  Darwin|Linux) ;;
  *) die "unsupported OS: $OS" ;;
esac

# ── dependencies ──────────────────────────────────────────────────────────────
check() { command -v "$1" &>/dev/null || die "$1 not found — $2"; }
check go   "install Go 1.22+: https://go.dev/dl"
check tmux "install tmux via your package manager"

GO_VER="$(go version | awk '{print $3}' | sed 's/go//')"
GO_OK=$(awk -v v="$GO_VER" 'BEGIN{split(v,a,"."); print (a[1]>1||(a[1]==1&&a[2]>=22))?"yes":"no"}')
[ "$GO_OK" = "yes" ] || die "Go 1.22+ required (found $GO_VER)"
ok "Go $GO_VER"

# ── download zip ──────────────────────────────────────────────────────────────
ZIP_URL="https://github.com/${GITHUB_USER}/${GITHUB_REPO}/archive/refs/heads/${BRANCH}.zip"
BUILD_DIR="${TMPDIR:-/tmp}/flowd-build-$$"
ZIP_PATH="${BUILD_DIR}.zip"
trap 'rm -rf "$BUILD_DIR" "$ZIP_PATH"' EXIT

info "downloading $ZIP_URL …"
if command -v curl &>/dev/null; then
  curl -fsSL "$ZIP_URL" -o "$ZIP_PATH"
elif command -v wget &>/dev/null; then
  wget -q "$ZIP_URL" -O "$ZIP_PATH"
else
  die "curl or wget required"
fi
ok "downloaded"

# ── extract ───────────────────────────────────────────────────────────────────
info "extracting …"
mkdir -p "$BUILD_DIR"
unzip -q "$ZIP_PATH" -d "$BUILD_DIR"
SRC_DIR="${BUILD_DIR}/${GITHUB_REPO}-${BRANCH}"
ok "extracted"

# ── build ─────────────────────────────────────────────────────────────────────
info "building …"
cd "$SRC_DIR"
go build -ldflags="-s -w" -o "$BINARY" ./cmd/fw
ok "build complete"

# ── install ───────────────────────────────────────────────────────────────────
DEST="${INSTALL_DIR}/${BINARY}"
if [ -w "$INSTALL_DIR" ]; then
  mv "$BINARY" "$DEST"
else
  info "need sudo to write to $INSTALL_DIR"
  sudo mv "$BINARY" "$DEST"
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
echo -e "${BOLD}done.${RESET}"
echo ""
if [ -f "${HOME}/.config/flowd/config.yaml" ]; then
  ok "existing config found — skipping init"
  echo "  run: fw start"
else
  echo "  run: fw init"
  echo "       fw start"
fi
