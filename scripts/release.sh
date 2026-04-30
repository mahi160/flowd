#!/usr/bin/env bash
# Usage: ./scripts/release.sh [patch|minor|major]
set -euo pipefail

BUMP="${1:-patch}"
DASHBOARD_ARTIFACT="internal/fw/static/dashboard.html"

need() {
  command -v "$1" >/dev/null 2>&1 || {
    echo "error: missing required command: $1" >&2
    exit 1
  }
}

need git
need go
need npm

case "$BUMP" in
  major|minor|patch) ;;
  *)
    echo "usage: $0 [patch|minor|major]" >&2
    exit 1
    ;;
esac

# Get the latest semver tag (ignore non-semver like v0.8)
CURRENT=$(git tag --list 'v[0-9]*.[0-9]*.[0-9]*' | sort -V | tail -1)
if [[ -z "$CURRENT" ]]; then
  echo "error: no semver tag found (expected vX.Y.Z)" >&2
  exit 1
fi

VERSION="${CURRENT#v}"
MAJOR=$(echo "$VERSION" | cut -d. -f1)
MINOR=$(echo "$VERSION" | cut -d. -f2)
PATCH=$(echo "$VERSION" | cut -d. -f3)

case "$BUMP" in
  major) MAJOR=$((MAJOR + 1)); MINOR=0; PATCH=0 ;;
  minor) MINOR=$((MINOR + 1)); PATCH=0 ;;
  patch) PATCH=$((PATCH + 1)) ;;
esac

NEXT="v${MAJOR}.${MINOR}.${PATCH}"

if git rev-parse "$NEXT" >/dev/null 2>&1; then
  echo "error: tag $NEXT already exists" >&2
  exit 1
fi

echo "  current: $CURRENT"
echo "  next:    $NEXT"
echo

read -rp "  build, commit dashboard artifact, tag and push $NEXT? (y/n) [y]: " CONFIRM
CONFIRM="${CONFIRM:-y}"
if [[ "$CONFIRM" != "y" ]]; then
  echo "aborted."
  exit 0
fi

echo
echo "==> building dashboard"
make dashboard

# The release script may auto-commit only the generated dashboard artifact.
# Any other dirty file means the release would not match reviewed source.
DIRTY_OTHER=$(git status --porcelain -- . ":!$DASHBOARD_ARTIFACT")
if [[ -n "$DIRTY_OTHER" ]]; then
  echo >&2
  echo "error: working tree has uncommitted changes outside $DASHBOARD_ARTIFACT:" >&2
  echo "$DIRTY_OTHER" >&2
  echo >&2
  echo "commit or stash those changes first, then rerun release." >&2
  exit 1
fi

if ! git diff --quiet -- "$DASHBOARD_ARTIFACT"; then
  echo "==> committing dashboard artifact"
  git add "$DASHBOARD_ARTIFACT"
  git commit -m "build: update dashboard artifact"
fi

echo "==> testing"
go test ./...

echo "==> building release binary smoke test"
go build -ldflags "-X github.com/mahi160/flowd/internal/fw.Version=${NEXT#v}" -o fw ./cmd/fw

echo "==> tagging $NEXT"
git tag "$NEXT"

if git remote get-url origin >/dev/null 2>&1; then
  echo "==> pushing commits and tag"
  git push origin HEAD
  git push origin "$NEXT"
fi

echo
echo "released $NEXT ✓"
echo "users can now run: fw update"
