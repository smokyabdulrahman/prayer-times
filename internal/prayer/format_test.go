package prayer

import (
	"strings"
	"testing"
	"time"
)

// helper: a fixed prayer and "now" time for format tests.
func formatTestPrayer() (Prayer, time.Time) {
	pTime := time.Date(2026, 2, 28, 15, 2, 0, 0, time.UTC)
	now := time.Date(2026, 2, 28, 12, 47, 0, 0, time.UTC)
	return Prayer{Name: "Asr", Time: pTime}, now
}

func TestFormatOutput_AllBuiltinModes(t *testing.T) {
	p, now := formatTestPrayer()

	tests := []struct {
		mode string
		want string
	}{
		{FormatTimeRemaining, "2h 15m"},
		{FormatNextPrayerTime, "15:02"},
		{FormatNameAndTime, "Asr 15:02"},
		{FormatNameAndRemaining, "Asr 2h 15m"},
		{FormatShortNameAndTime, "A 15:02"},
		{FormatShortNameAndRemain, "A 2h 15m"},
		{FormatFull, "Asr 15:02 (2h 15m)"},
	}

	for _, tt := range tests {
		t.Run(tt.mode, func(t *testing.T) {
			got := FormatOutput(p, now, tt.mode, "15:04")
			if got != tt.want {
				t.Errorf("FormatOutput(%q) = %q, want %q", tt.mode, got, tt.want)
			}
		})
	}
}

func TestFormatOutput_12HourFormat(t *testing.T) {
	p, now := formatTestPrayer()

	got := FormatOutput(p, now, FormatNameAndTime, "3:04 PM")
	if got != "Asr 3:02 PM" {
		t.Errorf("12h format = %q, want %q", got, "Asr 3:02 PM")
	}
}

func TestFormatOutput_UnknownModeDefaultsToNameAndTime(t *testing.T) {
	p, now := formatTestPrayer()

	got := FormatOutput(p, now, "nonexistent-format", "15:04")
	if got != "Asr 15:02" {
		t.Errorf("unknown mode = %q, want %q", got, "Asr 15:02")
	}
}

func TestFormatOutput_CustomTemplate(t *testing.T) {
	p, now := formatTestPrayer()

	tests := []struct {
		name string
		tmpl string
		want string
	}{
		{
			"name and remaining",
			"{{.Name}} in {{.Remaining}}",
			"Asr in 2h 15m",
		},
		{
			"short name and time",
			"{{.ShortName}} @ {{.Time}}",
			"A @ 15:02",
		},
		{
			"hours and minutes fields",
			"{{.Hours}}h {{.Minutes}}m until {{.Name}}",
			"2h 15m until Asr",
		},
		{
			"all fields",
			"{{.Name}}|{{.ShortName}}|{{.Time}}|{{.Remaining}}|{{.Hours}}|{{.Minutes}}",
			"Asr|A|15:02|2h 15m|2|15",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatOutput(p, now, tt.tmpl, "15:04")
			if got != tt.want {
				t.Errorf("custom template %q = %q, want %q", tt.tmpl, got, tt.want)
			}
		})
	}
}

func TestFormatOutput_InvalidTemplate(t *testing.T) {
	p, now := formatTestPrayer()

	got := FormatOutput(p, now, "{{.Invalid", "15:04")
	if !strings.HasPrefix(got, "template-err:") {
		t.Errorf("invalid template should return 'template-err:...', got %q", got)
	}
}

func TestFormatOutput_TemplateBadField(t *testing.T) {
	p, now := formatTestPrayer()

	// Accessing a non-existent field should produce a template execution error.
	got := FormatOutput(p, now, "{{.NonExistent}}", "15:04")
	if !strings.HasPrefix(got, "template-err:") {
		t.Errorf("bad field template should return 'template-err:...', got %q", got)
	}
}

func TestFormatOutput_LessThanOneHour(t *testing.T) {
	pTime := time.Date(2026, 2, 28, 13, 30, 0, 0, time.UTC)
	now := time.Date(2026, 2, 28, 13, 5, 0, 0, time.UTC)
	p := Prayer{Name: "Dhuhr", Time: pTime}

	got := FormatOutput(p, now, FormatTimeRemaining, "15:04")
	if got != "25m" {
		t.Errorf("time-remaining < 1h = %q, want %q", got, "25m")
	}
}

func TestFormatOutput_ZeroRemaining(t *testing.T) {
	now := time.Date(2026, 2, 28, 15, 2, 0, 0, time.UTC)
	p := Prayer{Name: "Asr", Time: now}

	got := FormatOutput(p, now, FormatTimeRemaining, "15:04")
	if got != "0m" {
		t.Errorf("zero remaining = %q, want %q", got, "0m")
	}
}
