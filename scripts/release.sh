#!/usr/bin/env bash
# Usage: ./scripts/release.sh [patch|minor|major]
set -euo pipefail

BUMP="${1:-patch}"

# Get the latest semver tag (ignore non-semver like v0.8)
CURRENT=$(git tag --list 'v[0-9]*.[0-9]*.[0-9]*' | sort -V | tail -1)
if [[ -z "$CURRENT" ]]; then
  echo "error: no semver tag found (expected vX.Y.Z)" >&2
  exit 1
fi

# Strip leading 'v'
VERSION="${CURRENT#v}"
MAJOR=$(echo "$VERSION" | cut -d. -f1)
MINOR=$(echo "$VERSION" | cut -d. -f2)
PATCH=$(echo "$VERSION" | cut -d. -f3)

case "$BUMP" in
  major) MAJOR=$((MAJOR + 1)); MINOR=0; PATCH=0 ;;
  minor) MINOR=$((MINOR + 1)); PATCH=0 ;;
  patch) PATCH=$((PATCH + 1)) ;;
  *)
    echo "usage: $0 [patch|minor|major]" >&2
    exit 1
    ;;
esac

NEXT="v${MAJOR}.${MINOR}.${PATCH}"

echo "  current: $CURRENT"
echo "  next:    $NEXT"
echo

read -rp "  tag and push $NEXT? (y/n) [y]: " CONFIRM
CONFIRM="${CONFIRM:-y}"
if [[ "$CONFIRM" != "y" ]]; then
  echo "aborted."
  exit 0
fi

git tag "$NEXT"
echo "  tagged $NEXT"

# Push tag if a remote exists
if git remote get-url origin &>/dev/null; then
  git push origin "$NEXT"
  echo "  pushed $NEXT to origin"
fi

echo
echo "  build with:"
echo "    go build -ldflags \"-X github.com/mahi160/flowd/internal/fw.Version=${NEXT#v}\" -o fw ./cmd/fw"
