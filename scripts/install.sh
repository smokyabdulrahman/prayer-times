#!/usr/bin/env bash
# install.sh -- Download the correct pre-built prayer-times binary.
#
# Usage:
#   ./scripts/install.sh              # auto-detect OS/arch
#   GITHUB_REPO=user/repo ./scripts/install.sh  # override repo

set -euo pipefail

CURRENT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PLUGIN_DIR="$(dirname "$CURRENT_DIR")"
BIN_DIR="$PLUGIN_DIR/bin"

GITHUB_REPO="${GITHUB_REPO:-smokyabdulrahman/prayer-times}"
BINARY_NAME="prayer-times"
ALIAS_NAME="pt"

# ---------------------------------------------------------------------------
# Detect platform
# ---------------------------------------------------------------------------
detect_os() {
    local os
    os="$(uname -s | tr '[:upper:]' '[:lower:]')"
    case "$os" in
        linux*)  echo "linux" ;;
        darwin*) echo "darwin" ;;
        *)
            echo "Unsupported OS: $os" >&2
            exit 1
            ;;
    esac
}

detect_arch() {
    local arch
    arch="$(uname -m)"
    case "$arch" in
        x86_64|amd64)  echo "amd64" ;;
        aarch64|arm64) echo "arm64" ;;
        *)
            echo "Unsupported architecture: $arch" >&2
            exit 1
            ;;
    esac
}

# ---------------------------------------------------------------------------
# Download helpers
# ---------------------------------------------------------------------------
download() {
    local url="$1"
    local dest="$2"

    if command -v curl >/dev/null 2>&1; then
        curl -fsSL -o "$dest" "$url"
    elif command -v wget >/dev/null 2>&1; then
        wget -q -O "$dest" "$url"
    else
        echo "Error: neither curl nor wget found. Cannot download binary." >&2
        exit 1
    fi
}

# ---------------------------------------------------------------------------
# Main
# ---------------------------------------------------------------------------
main() {
    local os arch asset_name download_url version

    os="$(detect_os)"
    arch="$(detect_arch)"
    asset_name="${BINARY_NAME}_${os}_${arch}.tar.gz"

    echo "Detected platform: ${os}/${arch}"

    # Get the latest release tag from GitHub API.
    local api_url="https://api.github.com/repos/${GITHUB_REPO}/releases/latest"
    local release_json
    if command -v curl >/dev/null 2>&1; then
        release_json="$(curl -fsSL "$api_url" 2>/dev/null)" || true
    elif command -v wget >/dev/null 2>&1; then
        release_json="$(wget -q -O - "$api_url" 2>/dev/null)" || true
    fi

    if [ -z "${release_json:-}" ]; then
        echo "Warning: could not fetch latest release info from GitHub." >&2
        echo "Attempting to build from source instead..." >&2
        build_from_source
        return
    fi

    # Extract the download URL for our asset.
    # Use grep + sed to avoid jq dependency.
    download_url="$(echo "$release_json" \
        | grep -o "\"browser_download_url\": *\"[^\"]*${asset_name}\"" \
        | sed 's/.*"\(https[^"]*\)".*/\1/' \
        | head -1)"

    if [ -z "$download_url" ]; then
        echo "Warning: no pre-built binary found for ${os}/${arch}." >&2
        echo "Attempting to build from source instead..." >&2
        build_from_source
        return
    fi

    version="$(echo "$release_json" \
        | grep -o '"tag_name": *"[^"]*"' \
        | sed 's/.*"\(v[^"]*\)".*/\1/' \
        | head -1)"

    echo "Downloading ${BINARY_NAME} ${version} for ${os}/${arch}..."
    mkdir -p "$BIN_DIR"

    local tmpfile
    tmpfile="$(mktemp)"
    download "$download_url" "$tmpfile"

    # Extract both prayer-times and pt binaries from the archive.
    tar -xzf "$tmpfile" -C "$BIN_DIR" "$BINARY_NAME" "$ALIAS_NAME" 2>/dev/null \
        || tar -xzf "$tmpfile" -C "$BIN_DIR" 2>/dev/null

    rm -f "$tmpfile"

    chmod +x "$BIN_DIR/$BINARY_NAME"
    [ -f "$BIN_DIR/$ALIAS_NAME" ] && chmod +x "$BIN_DIR/$ALIAS_NAME"
    echo "Installed ${BINARY_NAME} ${version} to ${BIN_DIR}/"
}

# ---------------------------------------------------------------------------
# Fallback: build from source if Go is available
# ---------------------------------------------------------------------------
build_from_source() {
    if ! command -v go >/dev/null 2>&1; then
        echo "Error: Go is not installed and no pre-built binary is available." >&2
        echo "Install Go (https://go.dev/dl/) or download a release from:" >&2
        echo "  https://github.com/${GITHUB_REPO}/releases" >&2
        exit 1
    fi

    echo "Building from source..."
    mkdir -p "$BIN_DIR"
    (cd "$PLUGIN_DIR" && go build -o "$BIN_DIR/$BINARY_NAME" ./cmd/prayer-times/)
    (cd "$PLUGIN_DIR" && go build -o "$BIN_DIR/$ALIAS_NAME" ./cmd/pt/)
    chmod +x "$BIN_DIR/$BINARY_NAME" "$BIN_DIR/$ALIAS_NAME"
    echo "Built ${BINARY_NAME} and ${ALIAS_NAME} to ${BIN_DIR}/"
}

main "$@"
