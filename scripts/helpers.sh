#!/usr/bin/env bash
# helpers.sh -- shared tmux option helpers for the prayer-times plugin.

# Read a tmux global option. Returns the default if the option is unset or empty.
get_tmux_option() {
    local option="$1"
    local default_value="$2"
    local option_value
    option_value="$(tmux show-option -gqv "$option")"
    if [ -z "$option_value" ]; then
        echo "$default_value"
    else
        echo "$option_value"
    fi
}

# Write a tmux global option.
set_tmux_option() {
    local option="$1"
    local value="$2"
    tmux set-option -gq "$option" "$value"
}
