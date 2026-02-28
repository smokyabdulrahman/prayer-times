package cli

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/smokyabdulrahman/prayer-times/internal/api"
	"github.com/smokyabdulrahman/prayer-times/internal/cache"
	"github.com/smokyabdulrahman/prayer-times/internal/geo"
	"github.com/smokyabdulrahman/prayer-times/internal/prayer"
	"github.com/spf13/cobra"
)

var (
	flagFormat     string
	flagTimeFormat string
	flagPrayers    string
)

func newNextCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "next",
		Short: "Show the next prayer with countdown",
		Long:  "Display the next upcoming prayer time with a countdown.\nThis is equivalent to the old tmux-prayer-times default behavior.",
		RunE:  runNext,
	}

	cmd.Flags().StringVar(&flagFormat, "format", prayer.FormatFull, "Display format: time-remaining, next-prayer-time, name-and-time, name-and-remaining, short-name-and-time, short-name-and-remaining, full, or a custom Go template")
	cmd.Flags().StringVar(&flagTimeFormat, "time-format", "24h", "Time format: 12h or 24h")
	cmd.Flags().StringVar(&flagPrayers, "prayers", "", "Comma-separated list of prayers to track (default: Fajr,Sunrise,Dhuhr,Asr,Maghrib,Isha)")

	return cmd
}

// locationMode describes how the user specified their location.
type locationMode int

const (
	locationCoords locationMode = iota
	locationCity
	locationAuto
)

func runNext(cmd *cobra.Command, args []string) error {
	// Determine which prayers to track.
	selectedPrayers := prayer.DefaultPrayerNames
	if flagPrayers != "" {
		selectedPrayers = strings.Split(flagPrayers, ",")
		for i := range selectedPrayers {
			selectedPrayers[i] = strings.TrimSpace(selectedPrayers[i])
		}
	}

	// Determine time format string.
	goTimeFmt := "15:04" // 24h
	if flagTimeFormat == "12h" {
		goTimeFmt = "3:04 PM"
	}

	// Initialize cache.
	c, err := cache.New(FlagCacheDir)
	if err != nil {
		// Cache init failure is non-fatal; we just skip caching.
		c = nil
		fmt.Fprintf(os.Stderr, "warning: cache disabled: %v\n", err)
	}

	now := time.Now()

	// Resolve location mode and coordinates.
	mode, lat, lon, city, country, tz, err := resolveLocation(FlagLatitude, FlagLongitude, FlagCity, FlagCountry, c)
	if err != nil {
		return err
	}

	// Fetch today's timings (from cache or API).
	timings, meta, err := fetchTimings(now, mode, lat, lon, city, country, FlagMethod, FlagSchool, c)
	if err != nil {
		return err
	}

	// Determine timezone.
	if tz == "" {
		tz = meta.Timezone
	}
	loc, err := time.LoadLocation(tz)
	if err != nil {
		return fmt.Errorf("invalid timezone %q: %w", tz, err)
	}

	// Re-anchor "now" to the API's timezone so comparisons work correctly
	// when the user is querying a different timezone than their local one.
	now = now.In(loc)
	today := now

	// Parse today's prayer times.
	prayers, err := prayer.ParseTimings(*timings, today, loc, selectedPrayers)
	if err != nil {
		return err
	}

	// Find the next prayer.
	next := prayer.NextPrayer(prayers, now)

	// If all today's prayers have passed, fetch tomorrow's first prayer.
	if next == nil {
		tomorrow := today.AddDate(0, 0, 1)

		tTimings, _, fetchErr := fetchTimings(tomorrow, mode, lat, lon, city, country, FlagMethod, FlagSchool, c)
		if fetchErr != nil {
			// Network failure for tomorrow's data: show last prayer with
			// a "done" indicator rather than crashing the status bar.
			if len(prayers) > 0 {
				last := prayers[len(prayers)-1]
				fmt.Printf("%s --:--", last.Name)
				return nil
			}
			return fmt.Errorf("failed to fetch tomorrow's times: %w", fetchErr)
		}

		tomorrowPrayers, err := prayer.ParseTimings(*tTimings, tomorrow, loc, selectedPrayers)
		if err != nil {
			return err
		}

		if len(tomorrowPrayers) > 0 {
			next = &tomorrowPrayers[0]
		}
	}

	if next == nil {
		return fmt.Errorf("could not determine next prayer")
	}

	// Format and print.
	output := prayer.FormatOutput(*next, now, flagFormat, goTimeFmt)
	fmt.Print(output)

	return nil
}

// resolveLocation determines the effective location based on user flags or auto-detection.
// It returns the mode used, the resolved lat/lon/city/country, and an optional timezone hint.
func resolveLocation(lat, lon float64, city, country string, c *cache.Cache) (locationMode, float64, float64, string, string, string, error) {
	switch {
	case lat != 0 || lon != 0:
		return locationCoords, lat, lon, "", "", "", nil
	case city != "":
		if country == "" {
			return 0, 0, 0, "", "", "", fmt.Errorf("--country is required when using --city")
		}
		return locationCity, 0, 0, city, country, "", nil
	default:
		// Try cached geolocation first.
		if c != nil {
			if cached := c.LoadGeo(); cached != nil {
				return locationCoords, cached.Latitude, cached.Longitude, "", "", cached.Timezone, nil
			}
		}

		// Fall back to IP-based geolocation.
		detected, err := geo.DetectLocation()
		if err != nil {
			return 0, 0, 0, "", "", "", fmt.Errorf("no location specified and auto-detection failed: %w", err)
		}

		// Cache the detected location.
		if c != nil {
			_ = c.SaveGeo(detected) // best-effort
		}

		return locationCoords, detected.Latitude, detected.Longitude, "", "", detected.Timezone, nil
	}
}

// fetchTimings returns prayer timings for the given date, using the cache when available.
func fetchTimings(date time.Time, mode locationMode, lat, lon float64, city, country string, method, school int, c *cache.Cache) (*api.Timings, *api.Meta, error) {
	// Try cache first.
	if c != nil {
		if entry := c.LoadTimings(date, lat, lon, city, country, method, school); entry != nil {
			return &entry.Timings, &entry.Meta, nil
		}
	}

	// Cache miss -- fetch from API.
	client := api.NewClient()
	var (
		resp *api.Response
		err  error
	)

	switch mode {
	case locationCity:
		resp, err = client.FetchByCity(date, city, country, method, school)
	default:
		resp, err = client.FetchByCoordinates(date, lat, lon, method, school)
	}

	if err != nil {
		return nil, nil, err
	}

	// Write to cache (best-effort).
	if c != nil {
		_ = c.SaveTimings(date, lat, lon, city, country, method, school, resp)
	}

	return &resp.Data.Timings, &resp.Data.Meta, nil
}
