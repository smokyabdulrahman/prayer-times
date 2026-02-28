package prayer

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
	"time"
)

// Format constants for display modes.
const (
	FormatTimeRemaining      = "time-remaining"
	FormatNextPrayerTime     = "next-prayer-time"
	FormatNameAndTime        = "name-and-time"
	FormatNameAndRemaining   = "name-and-remaining"
	FormatShortNameAndTime   = "short-name-and-time"
	FormatShortNameAndRemain = "short-name-and-remaining"
	FormatFull               = "full"
)

// FormatData is the data passed to custom Go templates.
type FormatData struct {
	Name      string // Full prayer name, e.g. "Asr"
	ShortName string // Abbreviated name, e.g. "A"
	Time      string // Formatted prayer time, e.g. "15:02" or "3:02 PM"
	Remaining string // Time remaining, e.g. "2h 15m"
	Hours     int    // Whole hours remaining
	Minutes   int    // Remaining minutes after hours
}

// FormatOutput formats a prayer for display according to the chosen format mode.
// timeFormat should be "15:04" for 24h or "3:04 PM" for 12h.
//
// If mode contains "{{", it is treated as a custom Go template string.
// Available template fields: .Name, .ShortName, .Time, .Remaining, .Hours, .Minutes
//
// Example: "{{.Name}} in {{.Remaining}}" -> "Asr in 2h 15m"
func FormatOutput(p Prayer, now time.Time, mode string, timeFormat string) string {
	d := TimeRemaining(p, now)
	remaining := FormatRemaining(d)
	timeStr := p.Time.Format(timeFormat)
	short := ShortNames[p.Name]

	// Custom template mode: any format string containing "{{" is a Go template.
	if strings.Contains(mode, "{{") {
		return formatCustom(mode, FormatData{
			Name:      p.Name,
			ShortName: short,
			Time:      timeStr,
			Remaining: remaining,
			Hours:     int(d.Hours()),
			Minutes:   int(d.Minutes()) % 60,
		})
	}

	switch mode {
	case FormatTimeRemaining:
		return remaining
	case FormatNextPrayerTime:
		return timeStr
	case FormatNameAndTime:
		return fmt.Sprintf("%s %s", p.Name, timeStr)
	case FormatNameAndRemaining:
		return fmt.Sprintf("%s %s", p.Name, remaining)
	case FormatShortNameAndTime:
		return fmt.Sprintf("%s %s", short, timeStr)
	case FormatShortNameAndRemain:
		return fmt.Sprintf("%s %s", short, remaining)
	case FormatFull:
		return fmt.Sprintf("%s %s (%s)", p.Name, timeStr, remaining)
	default:
		// Default to name-and-time.
		return fmt.Sprintf("%s %s", p.Name, timeStr)
	}
}

// formatCustom executes a user-provided Go template string against the FormatData.
func formatCustom(tmpl string, data FormatData) string {
	t, err := template.New("custom").Parse(tmpl)
	if err != nil {
		return fmt.Sprintf("template-err: %v", err)
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return fmt.Sprintf("template-err: %v", err)
	}

	return buf.String()
}
