#!/usr/bin/env bash
# prayer_times.sh -- called by tmux via #(...) command substitution.
# Reads tmux options, resolves the Go binary, and prints the formatted prayer time.

set -euo pipefail

CURRENT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PLUGIN_DIR="$(dirname "$CURRENT_DIR")"

# shellcheck source=helpers.sh
source "$CURRENT_DIR/helpers.sh"

# ---------------------------------------------------------------------------
# Resolve the Go binary
# ---------------------------------------------------------------------------
# Prefer the binary in the plugin's bin/ directory, then fall back to PATH.
BINARY=""

if [ -x "$PLUGIN_DIR/bin/prayer-times" ]; then
    BINARY="$PLUGIN_DIR/bin/prayer-times"
elif command -v prayer-times >/dev/null 2>&1; then
    BINARY="prayer-times"
fi

if [ -z "$BINARY" ]; then
    # Attempt auto-install on first run.
    if [ -x "$CURRENT_DIR/install.sh" ]; then
        "$CURRENT_DIR/install.sh" >/dev/null 2>&1
    fi
    if [ -x "$PLUGIN_DIR/bin/prayer-times" ]; then
        BINARY="$PLUGIN_DIR/bin/prayer-times"
    fi
fi

if [ -z "$BINARY" ]; then
    echo "pray-err:no-binary"
    exit 0  # exit 0 so tmux doesn't show an error
fi

# ---------------------------------------------------------------------------
# Read tmux options and build CLI flags
# ---------------------------------------------------------------------------
# Global flags (apply to all commands, placed before the subcommand).
build_global_flags() {
    local flags=()

    local city
    city="$(get_tmux_option "@prayer-times-city" "")"
    if [ -n "$city" ]; then
        flags+=("--city" "$city")
    fi

    local country
    country="$(get_tmux_option "@prayer-times-country" "")"
    if [ -n "$country" ]; then
        flags+=("--country" "$country")
    fi

    local latitude
    latitude="$(get_tmux_option "@prayer-times-latitude" "")"
    if [ -n "$latitude" ]; then
        flags+=("--latitude" "$latitude")
    fi

    local longitude
    longitude="$(get_tmux_option "@prayer-times-longitude" "")"
    if [ -n "$longitude" ]; then
        flags+=("--longitude" "$longitude")
    fi

    local method
    method="$(get_tmux_option "@prayer-times-method" "")"
    if [ -n "$method" ]; then
        flags+=("--method" "$method")
    fi

    local school
    school="$(get_tmux_option "@prayer-times-school" "")"
    if [ -n "$school" ]; then
        flags+=("--school" "$school")
    fi

    local time_format
    time_format="$(get_tmux_option "@prayer-times-time-format" "")"
    if [ -n "$time_format" ]; then
        flags+=("--time-format" "$time_format")
    fi

    local cache_dir
    cache_dir="$(get_tmux_option "@prayer-times-cache-dir" "")"
    if [ -n "$cache_dir" ]; then
        flags+=("--cache-dir" "$cache_dir")
    fi

    echo "${flags[@]}"
}

# Subcommand-specific flags for `next`.
build_next_flags() {
    local flags=()

    local format
    format="$(get_tmux_option "@prayer-times-format" "")"
    if [ -n "$format" ]; then
        flags+=("--format" "$format")
    fi

    local prayers
    prayers="$(get_tmux_option "@prayer-times-prayers" "")"
    if [ -n "$prayers" ]; then
        flags+=("--prayers" "$prayers")
    fi

    echo "${flags[@]}"
}

# ---------------------------------------------------------------------------
# Run the binary: prayer-times [global-flags] next [next-flags]
# ---------------------------------------------------------------------------
# shellcheck disable=SC2046
output=$("$BINARY" $(build_global_flags) next $(build_next_flags) 2>/dev/null) || output=""

# Prepend icon if configured.
icon="$(get_tmux_option "@prayer-times-icon" "")"
if [ -n "$icon" ]; then
    output="${icon} ${output}"
fi

echo -n "$output"
