#!/bin/bash
set -e

OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

case "$ARCH" in
    x86_64)  ARCH="amd64" ;;
    arm64|aarch64) ARCH="arm64" ;;
    *) echo "Unsupported architecture: $ARCH" >&2; exit 1 ;;
esac

BINARY="bin/hostfilemanager-${OS}-${ARCH}"

if [ ! -f "$BINARY" ]; then
    echo "Binary not found: $BINARY"
    echo "Run ./build.sh first."
    exit 1
fi

exec "$BINARY" "$@"
