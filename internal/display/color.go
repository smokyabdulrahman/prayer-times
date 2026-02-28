// Package display provides terminal color utilities using raw ANSI escape codes.
//
// It respects the NO_COLOR environment variable (https://no-color.org/) and
// detects whether stdout is a terminal. Colors are automatically disabled when
// output is piped or redirected, or when NO_COLOR is set.
package display

import (
	"fmt"
	"os"
)

// ANSI escape codes for styling.
const (
	reset  = "\033[0m"
	bold   = "\033[1m"
	dim    = "\033[2m"
	green  = "\033[32m"
	yellow = "\033[33m"
	cyan   = "\033[36m"
	fgGray = "\033[90m" // bright black = gray
)

// enabled reports whether color output is active.
// It is set once at init time.
var enabled bool

func init() {
	enabled = shouldEnable()
}

// shouldEnable determines whether to use color output.
func shouldEnable() bool {
	// Respect NO_COLOR (https://no-color.org/).
	if _, ok := os.LookupEnv("NO_COLOR"); ok {
		return false
	}
	// Respect FORCE_COLOR for testing.
	if _, ok := os.LookupEnv("FORCE_COLOR"); ok {
		return true
	}
	// Disable color when stdout is not a terminal (piped/redirected).
	return isTerminal(os.Stdout)
}

// isTerminal reports whether f is connected to a terminal.
// Uses Stat().Mode() to check for a character device — no cgo or external deps.
func isTerminal(f *os.File) bool {
	fi, err := f.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}

// SetEnabled overrides the auto-detected color state.
// Useful for testing or when --json forces plain output.
func SetEnabled(b bool) {
	enabled = b
}

// Enabled reports whether color output is currently active.
func Enabled() bool {
	return enabled
}

// wrap applies an ANSI code around text, only when colors are enabled.
func wrap(code, text string) string {
	if !enabled {
		return text
	}
	return code + text + reset
}

// Bold returns text rendered in bold.
func Bold(text string) string {
	return wrap(bold, text)
}

// Dim returns text rendered in dim/faint.
func Dim(text string) string {
	return wrap(dim, text)
}

// Green returns text rendered in green.
func Green(text string) string {
	return wrap(green, text)
}

// Yellow returns text rendered in yellow.
func Yellow(text string) string {
	return wrap(yellow, text)
}

// Cyan returns text rendered in cyan.
func Cyan(text string) string {
	return wrap(cyan, text)
}

// Gray returns text rendered in gray (bright black).
func Gray(text string) string {
	return wrap(fgGray, text)
}

// Accent returns text rendered in the accent color (cyan + bold).
// Used for the "next prayer" highlight.
func Accent(text string) string {
	if !enabled {
		return text
	}
	return bold + cyan + text + reset
}

// Boldf formats and bolds a string.
func Boldf(format string, a ...interface{}) string {
	return Bold(fmt.Sprintf(format, a...))
}
