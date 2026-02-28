package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/smokyabdulrahman/prayer-times/internal/cache"
	"github.com/smokyabdulrahman/prayer-times/internal/display"
	"github.com/smokyabdulrahman/prayer-times/internal/prayer"
	"github.com/spf13/cobra"
)

func runToday(cmd *cobra.Command, args []string) error {
	// Get merged config (CLI flags > config file > defaults).
	cfg := effectiveConfig(cmd)

	// Determine which prayers to track.
	selectedPrayers := prayer.DefaultPrayerNames
	if cfg.Prayers != "" {
		selectedPrayers = strings.Split(cfg.Prayers, ",")
		for i := range selectedPrayers {
			selectedPrayers[i] = strings.TrimSpace(selectedPrayers[i])
		}
	}

	// Determine Go time format from config.
	goTimeFmt := "15:04"
	if cfg.TimeFormat == "12h" {
		goTimeFmt = "3:04 PM"
	}

	// Initialize cache.
	c, err := cache.New(cfg.CacheDir)
	if err != nil {
		c = nil
		fmt.Fprintf(os.Stderr, "warning: cache disabled: %v\n", err)
	}

	now := time.Now()

	// Resolve location.
	loc, err := resolveLocation(cfg.Latitude, cfg.Longitude, cfg.City, cfg.Country, c)
	if err != nil {
		return err
	}

	// Get method/school from merged config.
	method := cfg.MethodOrDefault(-1)
	school := cfg.SchoolOrDefault(-1)

	// Fetch today's timings.
	result, err := fetchTimings(now, loc, method, school, c)
	if err != nil {
		return err
	}

	// Determine timezone.
	tz := loc.Timezone
	if tz == "" {
		tz = result.Meta.Timezone
	}
	tzLoc, err := time.LoadLocation(tz)
	if err != nil {
		return fmt.Errorf("invalid timezone %q: %w", tz, err)
	}

	// Re-anchor "now" to the API's timezone.
	now = now.In(tzLoc)

	// Parse today's prayer times.
	prayers, err := prayer.ParseTimings(result.Timings, now, tzLoc, selectedPrayers)
	if err != nil {
		return err
	}

	// Find current and next prayers.
	current := prayer.CurrentPrayer(prayers, now)
	next := prayer.NextPrayer(prayers, now)

	// Build location display string.
	locationStr := buildLocationStr(loc, result)

	// JSON output.
	if FlagJSON {
		return printTodayJSON(prayers, current, next, now, result, locationStr, tz, goTimeFmt)
	}

	// Rich terminal output.
	printTodayRich(prayers, current, next, now, result, locationStr, tz, goTimeFmt)
	return nil
}

// buildLocationStr builds a "City, Country" string from available data.
func buildLocationStr(loc resolvedLocation, result *fetchResult) string {
	if loc.City != "" && loc.Country != "" {
		return loc.City + ", " + loc.Country
	}
	// Fall back to coordinates.
	return fmt.Sprintf("%.4f, %.4f", result.Meta.Latitude, result.Meta.Longitude)
}

// printTodayRich renders the colored terminal output for today's prayer schedule.
func printTodayRich(prayers []prayer.Prayer, current, next *prayer.Prayer, now time.Time, result *fetchResult, locationStr, tz, goTimeFmt string) {
	fmt.Println()
	fmt.Printf("  %s\n", display.Bold("Prayer Times"))
	fmt.Println()

	// Location and date info.
	fmt.Printf("  %s\n", locationStr)
	fmt.Printf("  %s\n", tz)

	// Gregorian date.
	gregStr := formatGregorianDate(now, result)
	fmt.Printf("  %s\n", gregStr)

	// Hijri date.
	hijriStr := result.DateInfo.Hijri.Format()
	if hijriStr != "" {
		fmt.Printf("  %s\n", hijriStr)
	}

	fmt.Println()

	// Find the max prayer name length for alignment.
	maxNameLen := 0
	for _, p := range prayers {
		if len(p.Name) > maxNameLen {
			maxNameLen = len(p.Name)
		}
	}

	// Print each prayer.
	for _, p := range prayers {
		timeStr := p.Time.Format(goTimeFmt)
		nameStr := padRight(p.Name, maxNameLen)
		line := fmt.Sprintf("  %-*s  %s", maxNameLen, nameStr, timeStr)

		switch {
		case current != nil && p.Name == current.Name:
			// Current prayer: dimmed.
			fmt.Println(display.Dim(line))
		case next != nil && p.Name == next.Name:
			// Next prayer: accent color + countdown.
			remaining := prayer.FormatRemaining(prayer.TimeRemaining(p, now))
			suffix := fmt.Sprintf("  <- next in %s", remaining)
			fmt.Println(display.Accent(line) + display.Accent(suffix))
		default:
			fmt.Println(line)
		}
	}

	fmt.Println()
}

// formatGregorianDate returns a formatted Gregorian date string.
// Prefers API data; falls back to formatting `now`.
func formatGregorianDate(now time.Time, result *fetchResult) string {
	g := result.DateInfo.Gregorian
	if g.Day != "" && g.Month.En != "" && g.Year != "" {
		return g.Day + " " + g.Month.En + " " + g.Year
	}
	return now.Format("02 Jan 2006")
}

// padRight pads a string to the given width with spaces.
func padRight(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(s))
}

// todayJSON is the JSON output structure for the root command.
type todayJSON struct {
	Location todayJSONLocation `json:"location"`
	Date     todayJSONDate     `json:"date"`
	Timings  map[string]string `json:"timings"`
	Current  string            `json:"current"`
	Next     *todayJSONNext    `json:"next"`
}

type todayJSONLocation struct {
	City      string  `json:"city,omitempty"`
	Country   string  `json:"country,omitempty"`
	Timezone  string  `json:"timezone"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type todayJSONDate struct {
	Gregorian string `json:"gregorian"`
	Hijri     string `json:"hijri"`
}

type todayJSONNext struct {
	Prayer    string `json:"prayer"`
	Time      string `json:"time"`
	Remaining string `json:"remaining"`
}

// printTodayJSON renders structured JSON output.
func printTodayJSON(prayers []prayer.Prayer, current, next *prayer.Prayer, now time.Time, result *fetchResult, locationStr, tz, goTimeFmt string) error {
	timings := make(map[string]string)
	for _, p := range prayers {
		timings[strings.ToLower(p.Name)] = p.Time.Format(goTimeFmt)
	}

	out := todayJSON{
		Location: todayJSONLocation{
			Timezone:  tz,
			Latitude:  result.Meta.Latitude,
			Longitude: result.Meta.Longitude,
		},
		Date: todayJSONDate{
			Gregorian: formatGregorianDate(now, result),
			Hijri:     result.DateInfo.Hijri.Format(),
		},
		Timings: timings,
	}

	// Set city/country if available from meta or location string.
	if parts := strings.SplitN(locationStr, ", ", 2); len(parts) == 2 {
		out.Location.City = parts[0]
		out.Location.Country = parts[1]
	}

	if current != nil {
		out.Current = strings.ToLower(current.Name)
	}

	if next != nil {
		remaining := prayer.FormatRemaining(prayer.TimeRemaining(*next, now))
		out.Next = &todayJSONNext{
			Prayer:    strings.ToLower(next.Name),
			Time:      next.Time.Format(goTimeFmt),
			Remaining: remaining,
		}
	}

	data, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	fmt.Println(string(data))
	return nil
}
