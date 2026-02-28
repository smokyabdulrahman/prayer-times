package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/aalrahma/tmux-prayer-times/internal/api"
	"github.com/aalrahma/tmux-prayer-times/internal/cache"
	"github.com/aalrahma/tmux-prayer-times/internal/geo"
	"github.com/aalrahma/tmux-prayer-times/internal/prayer"
)

// version is set at build time via ldflags:
//
//	go build -ldflags "-X main.version=v1.0.0"
var version = "dev"

// calculationMethods lists all supported Al Adhan API calculation methods.
var calculationMethods = []struct {
	ID   int
	Name string
}{
	{0, "Shia Ithna-Ashari (Jafari)"},
	{1, "University of Islamic Sciences, Karachi"},
	{2, "Islamic Society of North America (ISNA)"},
	{3, "Muslim World League (MWL)"},
	{4, "Umm Al-Qura University, Makkah"},
	{5, "Egyptian General Authority of Survey"},
	{7, "Institute of Geophysics, University of Tehran"},
	{8, "Gulf Region"},
	{9, "Kuwait"},
	{10, "Qatar"},
	{11, "Majlis Ugama Islam Singapura (Singapore)"},
	{12, "Union Organization Islamic de France"},
	{13, "Diyanet Isleri Baskanligi, Turkey (experimental)"},
	{14, "Spiritual Administration of Muslims of Russia"},
	{15, "Moonsighting Committee Worldwide"},
	{16, "Dubai (experimental)"},
	{17, "JAKIM (Malaysia)"},
	{18, "Tunisia"},
	{19, "Algeria"},
	{20, "KEMENAG (Indonesia)"},
	{21, "Morocco"},
	{22, "Comunidade Islamica de Lisboa (Portugal)"},
	{23, "Ministry of Awqaf, Jordan"},
}

func main() {
	// Location flags
	latitude := flag.Float64("latitude", 0, "Latitude for prayer time calculation")
	longitude := flag.Float64("longitude", 0, "Longitude for prayer time calculation")
	city := flag.String("city", "", "City name (alternative to coordinates)")
	country := flag.String("country", "", "Country code (used with --city)")

	// Calculation flags
	method := flag.Int("method", -1, "Calculation method ID (0-23). -1 for API default.")
	school := flag.Int("school", -1, "Juristic school: 0=Shafi, 1=Hanafi. -1 for API default.")

	// Display flags
	format := flag.String("format", prayer.FormatNameAndTime, "Display format: time-remaining, next-prayer-time, name-and-time, name-and-remaining, short-name-and-time, short-name-and-remaining, full, or a custom Go template (e.g. '{{.Name}} in {{.Remaining}}'). Template fields: .Name, .ShortName, .Time, .Remaining, .Hours, .Minutes")
	timeFormat := flag.String("time-format", "24h", "Time format: 12h or 24h")
	prayers := flag.String("prayers", "", "Comma-separated list of prayers to track (default: Fajr,Sunrise,Dhuhr,Asr,Maghrib,Isha)")

	// Cache flags
	cacheDir := flag.String("cache-dir", "", "Cache directory (default: ~/.cache/tmux-prayer-times/)")

	// Info flags
	showVersion := flag.Bool("version", false, "Print version and exit")
	listMethods := flag.Bool("list-methods", false, "Print supported calculation methods and exit")

	flag.Parse()

	if *showVersion {
		fmt.Printf("tmux-prayer-times %s\n", version)
		return
	}

	if *listMethods {
		printMethods()
		return
	}

	if err := run(*latitude, *longitude, *city, *country, *method, *school, *format, *timeFormat, *prayers, *cacheDir); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

// printMethods prints the table of supported calculation methods.
func printMethods() {
	fmt.Println("Supported calculation methods:")
	fmt.Println()
	fmt.Printf("  %-4s %s\n", "ID", "Name")
	fmt.Printf("  %-4s %s\n", "──", "────")
	for _, m := range calculationMethods {
		fmt.Printf("  %-4d %s\n", m.ID, m.Name)
	}
	fmt.Println()
	fmt.Println("Use --method <ID> to select a calculation method.")
	fmt.Println("If omitted, the API picks a default based on your location.")
}

// locationMode describes how the user specified their location.
type locationMode int

const (
	locationCoords locationMode = iota
	locationCity
	locationAuto
)

func run(lat, lon float64, city, country string, method, school int, format, timeFmt, prayersFlag, cacheDir string) error {
	// Determine which prayers to track.
	selectedPrayers := prayer.DefaultPrayerNames
	if prayersFlag != "" {
		selectedPrayers = strings.Split(prayersFlag, ",")
		for i := range selectedPrayers {
			selectedPrayers[i] = strings.TrimSpace(selectedPrayers[i])
		}
	}

	// Determine time format string.
	goTimeFmt := "15:04" // 24h
	if timeFmt == "12h" {
		goTimeFmt = "3:04 PM"
	}

	// Initialize cache.
	c, err := cache.New(cacheDir)
	if err != nil {
		// Cache init failure is non-fatal; we just skip caching.
		c = nil
		fmt.Fprintf(os.Stderr, "warning: cache disabled: %v\n", err)
	}

	now := time.Now()

	// Resolve location mode and coordinates.
	mode, lat, lon, city, country, tz, err := resolveLocation(lat, lon, city, country, c)
	if err != nil {
		return err
	}

	// Fetch today's timings (from cache or API).
	timings, meta, err := fetchTimings(now, mode, lat, lon, city, country, method, school, c)
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

		tTimings, _, fetchErr := fetchTimings(tomorrow, mode, lat, lon, city, country, method, school, c)
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
	output := prayer.FormatOutput(*next, now, format, goTimeFmt)
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
