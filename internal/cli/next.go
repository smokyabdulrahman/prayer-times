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
	flagFormat  string
	flagPrayers string
)

func newNextCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "next",
		Short: "Show the next prayer with countdown",
		Long:  "Display the next upcoming prayer time with a countdown.\nThis is equivalent to the old tmux-prayer-times default behavior.",
		RunE:  runNext,
	}

	cmd.Flags().StringVar(&flagFormat, "format", prayer.FormatFull, "Display format: time-remaining, next-prayer-time, name-and-time, name-and-remaining, short-name-and-time, short-name-and-remaining, full, or a custom Go template")
	cmd.Flags().StringVar(&flagPrayers, "prayers", "", "Comma-separated list of prayers to track (overrides config)")

	return cmd
}

// locationMode describes how the user specified their location.
type locationMode int

const (
	locationCoords locationMode = iota
	locationCity
	locationAuto
)

// resolvedLocation holds the result of location resolution.
type resolvedLocation struct {
	Mode     locationMode
	Lat, Lon float64
	City     string
	Country  string
	Timezone string // optional hint from geo-detection
}

// fetchResult holds the data returned from a prayer times fetch.
type fetchResult struct {
	Timings  api.Timings
	Meta     api.Meta
	DateInfo api.DateInfo
}

func runNext(cmd *cobra.Command, args []string) error {
	// Get merged config (CLI flags > config file > defaults).
	cfg := effectiveConfig(cmd)

	// Determine which prayers to track.
	// Priority: --prayers flag > config > defaults.
	selectedPrayers := prayer.DefaultPrayerNames
	if cmd.Flags().Changed("prayers") && flagPrayers != "" {
		selectedPrayers = strings.Split(flagPrayers, ",")
		for i := range selectedPrayers {
			selectedPrayers[i] = strings.TrimSpace(selectedPrayers[i])
		}
	} else if cfg.Prayers != "" {
		selectedPrayers = strings.Split(cfg.Prayers, ",")
		for i := range selectedPrayers {
			selectedPrayers[i] = strings.TrimSpace(selectedPrayers[i])
		}
	}

	// Determine time format from merged config (already merged via effectiveConfig).
	timeFmt := cfg.TimeFormat
	goTimeFmt := "15:04" // 24h
	if timeFmt == "12h" {
		goTimeFmt = "3:04 PM"
	}

	// Initialize cache.
	c, err := cache.New(cfg.CacheDir)
	if err != nil {
		// Cache init failure is non-fatal; we just skip caching.
		c = nil
		fmt.Fprintf(os.Stderr, "warning: cache disabled: %v\n", err)
	}

	now := time.Now()

	// Resolve location mode and coordinates.
	// Priority: CLI flags > config > cached geo > IP auto-detect.
	loc, err := resolveLocation(cfg.Latitude, cfg.Longitude, cfg.City, cfg.Country, c)
	if err != nil {
		return err
	}

	// Get method/school from merged config.
	method := cfg.MethodOrDefault(-1)
	school := cfg.SchoolOrDefault(-1)

	// Fetch today's timings (from cache or API).
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

	// Re-anchor "now" to the API's timezone so comparisons work correctly
	// when the user is querying a different timezone than their local one.
	now = now.In(tzLoc)
	today := now

	// Parse today's prayer times.
	prayers, err := prayer.ParseTimings(result.Timings, today, tzLoc, selectedPrayers)
	if err != nil {
		return err
	}

	// Find the next prayer.
	next := prayer.NextPrayer(prayers, now)

	// If all today's prayers have passed, fetch tomorrow's first prayer.
	if next == nil {
		tomorrow := today.AddDate(0, 0, 1)

		tResult, fetchErr := fetchTimings(tomorrow, loc, method, school, c)
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

		tomorrowPrayers, err := prayer.ParseTimings(tResult.Timings, tomorrow, tzLoc, selectedPrayers)
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

// resolveLocation determines the effective location based on user flags, config, or auto-detection.
// Priority: CLI flags > config > cached geolocation > IP auto-detect.
func resolveLocation(lat, lon float64, city, country string, c *cache.Cache) (resolvedLocation, error) {
	switch {
	case lat != 0 || lon != 0:
		return resolvedLocation{Mode: locationCoords, Lat: lat, Lon: lon}, nil
	case city != "":
		if country == "" {
			return resolvedLocation{}, fmt.Errorf("--country is required when using --city")
		}
		return resolvedLocation{Mode: locationCity, City: city, Country: country}, nil
	default:
		// Try cached geolocation first.
		if c != nil {
			if cached := c.LoadGeo(); cached != nil {
				return resolvedLocation{
					Mode:     locationCoords,
					Lat:      cached.Latitude,
					Lon:      cached.Longitude,
					Timezone: cached.Timezone,
				}, nil
			}
		}

		// Fall back to IP-based geolocation.
		detected, err := geo.DetectLocation()
		if err != nil {
			return resolvedLocation{}, fmt.Errorf("no location specified and auto-detection failed: %w", err)
		}

		// Cache the detected location.
		if c != nil {
			_ = c.SaveGeo(detected) // best-effort
		}

		return resolvedLocation{
			Mode:     locationCoords,
			Lat:      detected.Latitude,
			Lon:      detected.Longitude,
			Timezone: detected.Timezone,
		}, nil
	}
}

// fetchTimings returns prayer timings for the given date, using the cache when available.
func fetchTimings(date time.Time, loc resolvedLocation, method, school int, c *cache.Cache) (*fetchResult, error) {
	// Try cache first.
	if c != nil {
		if entry := c.LoadTimings(date, loc.Lat, loc.Lon, loc.City, loc.Country, method, school); entry != nil {
			return &fetchResult{
				Timings:  entry.Timings,
				Meta:     entry.Meta,
				DateInfo: entry.DateInfo,
			}, nil
		}
	}

	// Cache miss -- fetch from API.
	client := api.NewClient()
	var (
		resp *api.Response
		err  error
	)

	switch loc.Mode {
	case locationCity:
		resp, err = client.FetchByCity(date, loc.City, loc.Country, method, school)
	default:
		resp, err = client.FetchByCoordinates(date, loc.Lat, loc.Lon, method, school)
	}

	if err != nil {
		return nil, err
	}

	// Write to cache (best-effort).
	if c != nil {
		_ = c.SaveTimings(date, loc.Lat, loc.Lon, loc.City, loc.Country, method, school, resp)
	}

	return &fetchResult{
		Timings:  resp.Data.Timings,
		Meta:     resp.Data.Meta,
		DateInfo: resp.Data.Date,
	}, nil
}
