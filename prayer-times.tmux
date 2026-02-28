#!/usr/bin/env bash
# prayer-times.tmux -- TPM entry point for the tmux-prayer-times plugin.
#
# This file is sourced by TPM (Tmux Plugin Manager) when the plugin is loaded.
# It replaces #{prayer_times} placeholders in the status bar with the actual
# script invocation that tmux will execute on each status-interval refresh.

CURRENT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# shellcheck source=scripts/helpers.sh
source "$CURRENT_DIR/scripts/helpers.sh"

# ---------------------------------------------------------------------------
# Interpolation: replace #{prayer_times} -> #(path/to/prayer_times.sh)
# ---------------------------------------------------------------------------
prayer_times_placeholder="\#{prayer_times}"
prayer_times_command="#($CURRENT_DIR/scripts/prayer_times.sh)"

do_interpolation() {
    local input="$1"
    echo "${input//$prayer_times_placeholder/$prayer_times_command}"
}

update_tmux_option() {
    local option="$1"
    local option_value
    option_value="$(get_tmux_option "$option" "")"
    if [ -z "$option_value" ]; then
        return
    fi
    local new_value
    new_value="$(do_interpolation "$option_value")"
    set_tmux_option "$option" "$new_value"
}

# ---------------------------------------------------------------------------
# Main
# ---------------------------------------------------------------------------
main() {
    update_tmux_option "status-right"
    update_tmux_option "status-left"
}

main
