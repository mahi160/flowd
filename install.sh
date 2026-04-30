#!/usr/bin/env bash
set -euo pipefail

# ── config ────────────────────────────────────────────────────────────────────
REPO="${FLOWD_REPO:-git@github.com:mahi/flowd.git}"
INSTALL_DIR="${FLOWD_INSTALL_DIR:-/usr/local/bin}"
BINARY="fw"
# ─────────────────────────────────────────────────────────────────────────────

RED='\033[0;31m'; GREEN='\033[0;32m'; CYAN='\033[0;36m'; BOLD='\033[1m'; RESET='\033[0m'

info()  { echo -e "${CYAN}→${RESET} $*"; }
ok()    { echo -e "${GREEN}✓${RESET} $*"; }
die()   { echo -e "${RED}✗ $*${RESET}" >&2; exit 1; }

echo -e "${BOLD}flowd installer${RESET}"
echo "────────────────────────────────"

# ── OS / arch ─────────────────────────────────────────────────────────────────
OS="$(uname -s)"
ARCH="$(uname -m)"
info "platform: $OS / $ARCH"

case "$OS" in
  Darwin|Linux) ;;
  *) die "unsupported OS: $OS" ;;
esac

# ── dependencies ──────────────────────────────────────────────────────────────
check() {
  command -v "$1" &>/dev/null || die "$1 not found — $2"
}

check git  "install git: https://git-scm.com"
check go   "install Go: https://go.dev/dl (1.22+)"
check tmux "install tmux via your package manager"

GO_MIN="1.22"
GO_VER="$(go version | awk '{print $3}' | sed 's/go//')"
# simple version check — compare major.minor
GO_OK=$(awk -v v="$GO_VER" -v m="$GO_MIN" 'BEGIN{
  split(v,a,"."); split(m,b,".")
  print (a[1]>b[1] || (a[1]==b[1] && a[2]>=b[2])) ? "yes" : "no"
}')
[ "$GO_OK" = "yes" ] || die "Go $GO_MIN+ required (found $GO_VER)"
ok "Go $GO_VER"

# ── clone / update ────────────────────────────────────────────────────────────
BUILD_DIR="${TMPDIR:-/tmp}/flowd-build-$$"
trap 'rm -rf "$BUILD_DIR"' EXIT

info "cloning $REPO …"
git clone --depth 1 "$REPO" "$BUILD_DIR" 2>&1 | tail -1

# ── build ─────────────────────────────────────────────────────────────────────
info "building …"
cd "$BUILD_DIR"
go build -ldflags="-s -w" -o "$BINARY" ./cmd/fw
ok "build complete"

# ── install ───────────────────────────────────────────────────────────────────
DEST="$INSTALL_DIR/$BINARY"

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
EXISTING_CFG="${HOME}/.config/flowd/config.yaml"
if [ -f "$EXISTING_CFG" ]; then
  ok "existing config found at $EXISTING_CFG — skipping init"
  echo "  run: fw start"
else
  echo "  run: fw init"
  echo "       fw start"
fi
