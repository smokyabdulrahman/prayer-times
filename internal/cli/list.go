package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/smokyabdulrahman/prayer-times/internal/api"
	"github.com/smokyabdulrahman/prayer-times/internal/cache"
	"github.com/smokyabdulrahman/prayer-times/internal/display"
	"github.com/smokyabdulrahman/prayer-times/internal/prayer"
	"github.com/spf13/cobra"
)

// dayData holds a single day's parsed data for list/query output.
type dayData struct {
	Date     time.Time
	Timings  api.Timings
	DateInfo api.DateInfo
	Meta     api.Meta
}

// runList is the handler for the list subcommand.
func runList(cmd *cobra.Command, args []string, defaultDays int) error {
	days := defaultDays
	if len(args) > 0 {
		n, err := strconv.Atoi(args[0])
		if err != nil || n < 1 {
			return fmt.Errorf("invalid number of days: %q (must be a positive integer)", args[0])
		}
		days = n
	}

	cfg := effectiveConfig(cmd)

	selectedPrayers := prayer.DefaultPrayerNames
	if cfg.Prayers != "" {
		selectedPrayers = strings.Split(cfg.Prayers, ",")
		for i := range selectedPrayers {
			selectedPrayers[i] = strings.TrimSpace(selectedPrayers[i])
		}
	}

	goTimeFmt := "15:04"
	if cfg.TimeFormat == "12h" {
		goTimeFmt = "3:04 PM"
	}

	c, err := cache.New(cfg.CacheDir)
	if err != nil {
		c = nil
		fmt.Fprintf(os.Stderr, "warning: cache disabled: %v\n", err)
	}

	now := time.Now()

	loc, err := resolveLocation(cfg.Latitude, cfg.Longitude, cfg.City, cfg.Country, c)
	if err != nil {
		return err
	}

	method := cfg.MethodOrDefault(-1)
	school := cfg.SchoolOrDefault(-1)

	// Fetch calendar data for the needed days.
	daysList, err := fetchCalendarDays(now, days, loc, method, school, c)
	if err != nil {
		return err
	}

	// Determine timezone from the first day's meta.
	tz := loc.Timezone
	if tz == "" && len(daysList) > 0 {
		tz = daysList[0].Meta.Timezone
	}
	tzLoc, err := time.LoadLocation(tz)
	if err != nil {
		return fmt.Errorf("invalid timezone %q: %w", tz, err)
	}

	now = now.In(tzLoc)
	todayStr := now.Format("2006-01-02")

	// Build location string.
	locationStr := buildLocationStr(loc, &fetchResult{Meta: daysList[0].Meta})

	if FlagJSON {
		return printListJSON(daysList, selectedPrayers, locationStr, tz, goTimeFmt, tzLoc)
	}

	// Rich terminal output.
	fmt.Println()
	fmt.Printf("  %s\n", display.Bold(fmt.Sprintf("Prayer Times \u2014 %d Days", days)))
	fmt.Println()
	fmt.Printf("  %s\n", locationStr)
	fmt.Println()

	// Build table.
	headers := []string{"Date"}
	headers = append(headers, selectedPrayers...)
	tbl := display.NewTable(headers)

	for i, dd := range daysList {
		dateInTZ := dd.Date.In(tzLoc)
		dateLabel := dateInTZ.Format("Mon 02 Jan")

		parsed, err := prayer.ParseTimings(dd.Timings, dateInTZ, tzLoc, selectedPrayers)
		if err != nil {
			return err
		}

		row := []string{dateLabel}
		for _, p := range parsed {
			row = append(row, p.Time.Format(goTimeFmt))
		}
		tbl.AddRow(row)

		// Highlight today's row.
		if dateInTZ.Format("2006-01-02") == todayStr {
			tbl.SetHighlightRow(i)
		}
	}

	fmt.Print(tbl.Render())
	fmt.Println()
	return nil
}

// fetchCalendarDays fetches prayer data for `days` consecutive days starting from `start`.
// It uses the calendar endpoint for efficiency (fetches whole months) with caching.
func fetchCalendarDays(start time.Time, days int, loc resolvedLocation, method, school int, c *cache.Cache) ([]dayData, error) {
	client := api.NewClient()

	// Determine which year/month combos we need.
	type yearMonth struct {
		year, month int
	}
	needed := make(map[yearMonth]bool)
	for i := 0; i < days; i++ {
		d := start.AddDate(0, 0, i)
		needed[yearMonth{d.Year(), int(d.Month())}] = true
	}

	// Fetch each needed month (from cache or API).
	// monthData maps year/month -> slice of api.Data (one per day).
	monthData := make(map[yearMonth][]api.Data)

	for ym := range needed {
		// Try cache first.
		if c != nil {
			if entry := c.LoadCalendar(ym.year, ym.month, loc.Lat, loc.Lon, loc.City, loc.Country, method, school); entry != nil {
				monthData[ym] = entry.Days
				continue
			}
		}

		// Fetch from API.
		var resp *api.CalendarResponse
		var err error

		switch loc.Mode {
		case locationCity:
			resp, err = client.FetchCalendarByCity(ym.year, ym.month, loc.City, loc.Country, method, school)
		default:
			resp, err = client.FetchCalendarByCoordinates(ym.year, ym.month, loc.Lat, loc.Lon, method, school)
		}
		if err != nil {
			return nil, fmt.Errorf("failed to fetch calendar for %d-%02d: %w", ym.year, ym.month, err)
		}

		monthData[ym] = resp.Data

		// Cache (best-effort).
		if c != nil {
			_ = c.SaveCalendar(ym.year, ym.month, loc.Lat, loc.Lon, loc.City, loc.Country, method, school, resp)
		}
	}

	// Assemble the days in order.
	var result []dayData
	for i := 0; i < days; i++ {
		d := start.AddDate(0, 0, i)
		ym := yearMonth{d.Year(), int(d.Month())}
		daysInMonth := monthData[ym]

		dayIdx := d.Day() - 1 // 0-indexed
		if dayIdx < 0 || dayIdx >= len(daysInMonth) {
			return nil, fmt.Errorf("day %d out of range for %d-%02d (got %d days)", d.Day(), ym.year, ym.month, len(daysInMonth))
		}

		apiData := daysInMonth[dayIdx]
		result = append(result, dayData{
			Date:     d,
			Timings:  apiData.Timings,
			DateInfo: apiData.Date,
			Meta:     apiData.Meta,
		})
	}

	return result, nil
}

// listJSONOutput is the JSON structure for the list command.
type listJSONOutput struct {
	Location todayJSONLocation `json:"location"`
	Days     []listJSONDay     `json:"days"`
}

type listJSONDay struct {
	Date    string            `json:"date"`
	Hijri   string            `json:"hijri"`
	Timings map[string]string `json:"timings"`
}

func printListJSON(daysList []dayData, selectedPrayers []string, locationStr, tz, goTimeFmt string, tzLoc *time.Location) error {
	out := listJSONOutput{
		Location: todayJSONLocation{
			Timezone:  tz,
			Latitude:  daysList[0].Meta.Latitude,
			Longitude: daysList[0].Meta.Longitude,
		},
	}

	if parts := strings.SplitN(locationStr, ", ", 2); len(parts) == 2 {
		out.Location.City = parts[0]
		out.Location.Country = parts[1]
	}

	for _, dd := range daysList {
		dateInTZ := dd.Date.In(tzLoc)
		parsed, err := prayer.ParseTimings(dd.Timings, dateInTZ, tzLoc, selectedPrayers)
		if err != nil {
			return err
		}

		timings := make(map[string]string)
		for _, p := range parsed {
			timings[strings.ToLower(p.Name)] = p.Time.Format(goTimeFmt)
		}

		out.Days = append(out.Days, listJSONDay{
			Date:    dateInTZ.Format("02 Jan 2006"),
			Hijri:   dd.DateInfo.Hijri.Format(),
			Timings: timings,
		})
	}

	data, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	fmt.Println(string(data))
	return nil
}
