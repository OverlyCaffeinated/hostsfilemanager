#!/bin/bash
set -e

BIN_DIR="bin"
mkdir -p "$BIN_DIR"

echo "Building for Linux amd64..."
GOOS=linux GOARCH=amd64 go build -o "$BIN_DIR/hostfilemanager-linux-amd64" .

echo "Building for Linux arm64..."
GOOS=linux GOARCH=arm64 go build -o "$BIN_DIR/hostfilemanager-linux-arm64" .

echo "Building for Apple Silicon (darwin/arm64)..."
GOOS=darwin GOARCH=arm64 go build -o "$BIN_DIR/hostfilemanager-darwin-arm64" .

echo "Done. Binaries are in $BIN_DIR/"
