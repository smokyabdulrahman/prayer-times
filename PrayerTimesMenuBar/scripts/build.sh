#!/usr/bin/env bash
#
# Build script for PrayerTimesMenuBar.app
#
# Creates a self-contained macOS app bundle with the prayer-times CLI binary
# embedded in Resources.
#
# Usage:
#   ./scripts/build.sh                        # debug build, current arch
#   ./scripts/build.sh --release              # release build, current arch
#   ./scripts/build.sh --release --arch arm64 # release build, specific arch
#   ./scripts/build.sh --release --output dist/PrayerTimesMenuBar_arm64.app
#
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
REPO_ROOT="$(cd "$PROJECT_DIR/.." && pwd)"

APP_NAME="PrayerTimesMenuBar"
BUNDLE_NAME="${APP_NAME}.app"

# Defaults
SWIFT_CONFIG="debug"
TARGET_ARCH=""
OUTPUT_DIR=""

# Parse flags
while [[ $# -gt 0 ]]; do
    case "$1" in
        --release)
            SWIFT_CONFIG="release"
            shift
            ;;
        --arch)
            TARGET_ARCH="$2"
            shift 2
            ;;
        --output)
            OUTPUT_DIR="$2"
            shift 2
            ;;
        *)
            echo "Unknown flag: $1" >&2
            exit 1
            ;;
    esac
done

# Resolve architecture
if [[ -z "$TARGET_ARCH" ]]; then
    TARGET_ARCH="$(uname -m)"
fi

# Normalize arch names
case "$TARGET_ARCH" in
    x86_64|amd64) GO_ARCH="amd64"; SWIFT_ARCH="x86_64" ;;
    arm64|aarch64) GO_ARCH="arm64"; SWIFT_ARCH="arm64" ;;
    *)
        echo "ERROR: Unsupported architecture: ${TARGET_ARCH}" >&2
        exit 1
        ;;
esac

BUILD_DIR="${PROJECT_DIR}/.build/bundle"

echo "==> Building prayer-times CLI (Go, darwin/${GO_ARCH})..."
GOOS=darwin GOARCH="$GO_ARCH" CGO_ENABLED=0 go build \
    -ldflags "-s -w" \
    -o "${BUILD_DIR}/prayer-times" \
    "${REPO_ROOT}/cmd/prayer-times"

echo "==> Building ${APP_NAME} (Swift, ${SWIFT_CONFIG}, ${SWIFT_ARCH})..."
swift build \
    -c "$SWIFT_CONFIG" \
    --arch "$SWIFT_ARCH" \
    --package-path "$PROJECT_DIR"

# Locate the Swift binary
SWIFT_BIN="${PROJECT_DIR}/.build/${SWIFT_CONFIG}/${APP_NAME}"
if [[ ! -f "$SWIFT_BIN" ]]; then
    # When --arch is specified, SPM may place the binary in an arch-specific path
    SWIFT_BIN="${PROJECT_DIR}/.build/${SWIFT_ARCH}-apple-macosx/${SWIFT_CONFIG}/${APP_NAME}"
fi
if [[ ! -f "$SWIFT_BIN" ]]; then
    echo "ERROR: Swift binary not found" >&2
    echo "  Checked: ${PROJECT_DIR}/.build/${SWIFT_CONFIG}/${APP_NAME}" >&2
    echo "  Checked: ${PROJECT_DIR}/.build/${SWIFT_ARCH}-apple-macosx/${SWIFT_CONFIG}/${APP_NAME}" >&2
    exit 1
fi

# Determine output location
if [[ -z "$OUTPUT_DIR" ]]; then
    BUNDLE_DIR="${BUILD_DIR}/${BUNDLE_NAME}"
else
    BUNDLE_DIR="$OUTPUT_DIR"
fi

echo "==> Assembling ${BUNDLE_NAME}..."
rm -rf "$BUNDLE_DIR"
mkdir -p "${BUNDLE_DIR}/Contents/MacOS"
mkdir -p "${BUNDLE_DIR}/Contents/Resources"

# Copy Swift binary
cp "$SWIFT_BIN" "${BUNDLE_DIR}/Contents/MacOS/${APP_NAME}"

# Copy Go CLI binary into Resources
cp "${BUILD_DIR}/prayer-times" "${BUNDLE_DIR}/Contents/Resources/prayer-times"
chmod +x "${BUNDLE_DIR}/Contents/Resources/prayer-times"

# Copy Info.plist
cp "${PROJECT_DIR}/Resources/Info.plist" "${BUNDLE_DIR}/Contents/Info.plist"

echo "==> Build complete!"
echo "    ${BUNDLE_DIR}"
echo ""
echo "    Run with:  open ${BUNDLE_DIR}"
echo "    Or:        ${BUNDLE_DIR}/Contents/MacOS/${APP_NAME}"
