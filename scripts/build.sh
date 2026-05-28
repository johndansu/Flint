#!/usr/bin/env bash
# Build Flint binaries for all supported platforms.
# Usage: ./scripts/build.sh [version]
# Example: ./scripts/build.sh 0.1.0
set -euo pipefail

VERSION="${1:-dev}"
LDFLAGS="-s -w -X main.version=${VERSION}"
DIST="dist"

TARGETS=(
  "darwin/amd64"
  "darwin/arm64"
  "linux/amd64"
  "linux/arm64"
  "windows/amd64"
)

mkdir -p "$DIST"

for TARGET in "${TARGETS[@]}"; do
  OS="${TARGET%%/*}"
  ARCH="${TARGET##*/}"
  EXT=""
  [[ "$OS" == "windows" ]] && EXT=".exe"

  echo "  flint  ${OS}/${ARCH}"
  GOOS="$OS" GOARCH="$ARCH" go build \
    -ldflags="$LDFLAGS" \
    -trimpath \
    -o "${DIST}/flint-${OS}-${ARCH}${EXT}" \
    ./core/cmd/flint/

  echo "  flintd ${OS}/${ARCH}"
  GOOS="$OS" GOARCH="$ARCH" go build \
    -ldflags="$LDFLAGS" \
    -trimpath \
    -o "${DIST}/flintd-${OS}-${ARCH}${EXT}" \
    ./core/cmd/daemon/
done

echo ""
echo "Built $(ls "$DIST" | wc -l | tr -d ' ') artifacts → $DIST/"
ls -lh "$DIST"
