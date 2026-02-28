package prayer

import (
	"fmt"
	"strings"
	"time"

	"github.com/smokyabdulrahman/prayer-times/internal/api"
)

// Prayer represents a single prayer with its name and time.
type Prayer struct {
	Name string
	Time time.Time
}

// AllPrayerNames lists every prayer/event the API can return, in chronological order.
var AllPrayerNames = []string{
	"Fajr", "Sunrise", "Dhuhr", "Asr", "Sunset", "Maghrib", "Isha",
	"Imsak", "Midnight", "Firstthird", "Lastthird",
}

// DefaultPrayerNames are the prayers tracked by default.
var DefaultPrayerNames = []string{
	"Fajr", "Sunrise", "Dhuhr", "Asr", "Maghrib", "Isha",
}

// ShortNames maps full prayer names to single-character abbreviations.
var ShortNames = map[string]string{
	"Fajr":       "F",
	"Sunrise":    "S",
	"Dhuhr":      "D",
	"Asr":        "A",
	"Sunset":     "St",
	"Maghrib":    "M",
	"Isha":       "I",
	"Imsak":      "Im",
	"Midnight":   "Mi",
	"Firstthird": "F3",
	"Lastthird":  "L3",
}

// ParseTimings converts API timings into a slice of Prayer structs for the given date.
// It filters to only include the specified prayer names.
// The location is used to construct proper time.Time values in the correct timezone.
func ParseTimings(timings api.Timings, date time.Time, loc *time.Location, selected []string) ([]Prayer, error) {
	timingMap := map[string]string{
		"Fajr":       timings.Fajr,
		"Sunrise":    timings.Sunrise,
		"Dhuhr":      timings.Dhuhr,
		"Asr":        timings.Asr,
		"Sunset":     timings.Sunset,
		"Maghrib":    timings.Maghrib,
		"Isha":       timings.Isha,
		"Imsak":      timings.Imsak,
		"Midnight":   timings.Midnight,
		"Firstthird": timings.Firstthird,
		"Lastthird":  timings.Lastthird,
	}

	var prayers []Prayer
	for _, name := range selected {
		raw, ok := timingMap[name]
		if !ok {
			return nil, fmt.Errorf("unknown prayer name: %s", name)
		}

		t, err := parseTimeStr(raw, date, loc)
		if err != nil {
			return nil, fmt.Errorf("failed to parse time for %s (%q): %w", name, raw, err)
		}

		prayers = append(prayers, Prayer{Name: name, Time: t})
	}

	return prayers, nil
}

// NextPrayer finds the next upcoming prayer from the given slice, relative to now.
// If all prayers for today have passed, it returns nil (caller should fetch tomorrow's Fajr).
func NextPrayer(prayers []Prayer, now time.Time) *Prayer {
	for i := range prayers {
		if prayers[i].Time.After(now) {
			return &prayers[i]
		}
	}
	return nil
}

// TimeRemaining returns the duration until the given prayer time.
func TimeRemaining(prayer Prayer, now time.Time) time.Duration {
	return prayer.Time.Sub(now)
}

// FormatRemaining formats a duration as "Xh Ym" or "Ym" if less than an hour.
func FormatRemaining(d time.Duration) string {
	if d < 0 {
		return "0m"
	}
	h := int(d.Hours())
	m := int(d.Minutes()) % 60

	if h > 0 {
		return fmt.Sprintf("%dh %dm", h, m)
	}
	return fmt.Sprintf("%dm", m)
}

// parseTimeStr parses a time string like "15:02" or "15:02 (BST)" into a time.Time
// on the given date in the given location.
func parseTimeStr(raw string, date time.Time, loc *time.Location) (time.Time, error) {
	// Strip timezone suffix like " (BST)" that the API sometimes appends.
	s := strings.TrimSpace(raw)
	if idx := strings.Index(s, " "); idx != -1 {
		s = s[:idx]
	}

	parts := strings.Split(s, ":")
	if len(parts) != 2 {
		return time.Time{}, fmt.Errorf("invalid time format: %q", raw)
	}

	var hour, min int
	if _, err := fmt.Sscanf(parts[0], "%d", &hour); err != nil {
		return time.Time{}, fmt.Errorf("invalid hour in %q: %w", raw, err)
	}
	if _, err := fmt.Sscanf(parts[1], "%d", &min); err != nil {
		return time.Time{}, fmt.Errorf("invalid minute in %q: %w", raw, err)
	}

	return time.Date(date.Year(), date.Month(), date.Day(), hour, min, 0, 0, loc), nil
}
