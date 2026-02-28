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

var flagQueryDays string

func newQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "query <prayer>",
		Short: "Query a specific prayer time",
		Long:  "Query a specific prayer time for today, or across multiple days with --days.\n\nValid prayer names: Fajr, Sunrise, Dhuhr, Asr, Sunset, Maghrib, Isha, Imsak, Midnight, Firstthird, Lastthird",
		Args:  cobra.ExactArgs(1),
		RunE:  runQuery,
	}

	cmd.Flags().StringVar(&flagQueryDays, "days", "", "Number of days to show (or 'week'/'month')")

	return cmd
}

func runQuery(cmd *cobra.Command, args []string) error {
	prayerName := args[0]

	// Validate prayer name.
	valid := false
	for _, name := range prayer.AllPrayerNames {
		if strings.EqualFold(name, prayerName) {
			prayerName = name // normalize case
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("unknown prayer %q; valid names: %s", args[0], strings.Join(prayer.AllPrayerNames, ", "))
	}

	cfg := effectiveConfig(cmd)

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

	// Determine number of days.
	days := 1
	if flagQueryDays != "" {
		switch flagQueryDays {
		case "week":
			days = 7
		case "month":
			days = 30
		default:
			n, err := fmt.Sscanf(flagQueryDays, "%d", &days)
			if err != nil || n != 1 || days < 1 {
				return fmt.Errorf("invalid --days value %q: must be a positive integer, 'week', or 'month'", flagQueryDays)
			}
		}
	}

	// Single day: use the daily endpoint.
	if days == 1 {
		return runQuerySingleDay(cmd, prayerName, now, loc, method, school, c, goTimeFmt)
	}

	// Multi-day: use the calendar endpoint.
	return runQueryMultiDay(cmd, prayerName, days, now, loc, method, school, c, goTimeFmt)
}

func runQuerySingleDay(cmd *cobra.Command, prayerName string, now time.Time, loc resolvedLocation, method, school int, c *cache.Cache, goTimeFmt string) error {
	result, err := fetchTimings(now, loc, method, school, c)
	if err != nil {
		return err
	}

	tz := loc.Timezone
	if tz == "" {
		tz = result.Meta.Timezone
	}
	tzLoc, err := time.LoadLocation(tz)
	if err != nil {
		return fmt.Errorf("invalid timezone %q: %w", tz, err)
	}
	now = now.In(tzLoc)

	parsed, err := prayer.ParseTimings(result.Timings, now, tzLoc, []string{prayerName})
	if err != nil {
		return err
	}

	if len(parsed) == 0 {
		return fmt.Errorf("no timing found for %s", prayerName)
	}

	p := parsed[0]
	timeStr := p.Time.Format(goTimeFmt)

	if FlagJSON {
		out := queryJSONSingle{
			Prayer: strings.ToLower(prayerName),
			Time:   timeStr,
			Date:   now.Format("02 Jan 2006"),
			Hijri:  result.DateInfo.Hijri.Format(),
		}
		data, err := json.MarshalIndent(out, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
		return nil
	}

	fmt.Printf("%s %s\n", prayerName, timeStr)
	return nil
}

func runQueryMultiDay(cmd *cobra.Command, prayerName string, days int, now time.Time, loc resolvedLocation, method, school int, c *cache.Cache, goTimeFmt string) error {
	daysList, err := fetchCalendarDays(now, days, loc, method, school, c)
	if err != nil {
		return err
	}

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

	locationStr := buildLocationStr(loc, &fetchResult{Meta: daysList[0].Meta})

	if FlagJSON {
		return printQueryJSON(daysList, prayerName, locationStr, tz, goTimeFmt, tzLoc)
	}

	// Rich terminal output.
	fmt.Println()
	fmt.Printf("  %s\n", display.Bold(fmt.Sprintf("%s Times \u2014 %d Days", prayerName, days)))
	fmt.Println()
	fmt.Printf("  %s\n", locationStr)
	fmt.Println()

	tbl := display.NewTable([]string{"Date", prayerName})

	for i, dd := range daysList {
		dateInTZ := dd.Date.In(tzLoc)
		dateLabel := dateInTZ.Format("Mon 02 Jan")

		parsed, err := prayer.ParseTimings(dd.Timings, dateInTZ, tzLoc, []string{prayerName})
		if err != nil {
			return err
		}

		timeStr := ""
		if len(parsed) > 0 {
			timeStr = parsed[0].Time.Format(goTimeFmt)
		}

		tbl.AddRow([]string{dateLabel, timeStr})

		if dateInTZ.Format("2006-01-02") == todayStr {
			tbl.SetHighlightRow(i)
		}
	}

	fmt.Print(tbl.Render())
	fmt.Println()
	return nil
}

type queryJSONSingle struct {
	Prayer string `json:"prayer"`
	Time   string `json:"time"`
	Date   string `json:"date"`
	Hijri  string `json:"hijri"`
}

type queryJSONMulti struct {
	Location todayJSONLocation `json:"location"`
	Prayer   string            `json:"prayer"`
	Days     []queryJSONDay    `json:"days"`
}

type queryJSONDay struct {
	Date  string `json:"date"`
	Hijri string `json:"hijri"`
	Time  string `json:"time"`
}

func printQueryJSON(daysList []dayData, prayerName, locationStr, tz, goTimeFmt string, tzLoc *time.Location) error {
	out := queryJSONMulti{
		Location: todayJSONLocation{
			Timezone:  tz,
			Latitude:  daysList[0].Meta.Latitude,
			Longitude: daysList[0].Meta.Longitude,
		},
		Prayer: strings.ToLower(prayerName),
	}

	if parts := strings.SplitN(locationStr, ", ", 2); len(parts) == 2 {
		out.Location.City = parts[0]
		out.Location.Country = parts[1]
	}

	for _, dd := range daysList {
		dateInTZ := dd.Date.In(tzLoc)
		parsed, err := prayer.ParseTimings(dd.Timings, dateInTZ, tzLoc, []string{prayerName})
		if err != nil {
			return err
		}

		timeStr := ""
		if len(parsed) > 0 {
			timeStr = parsed[0].Time.Format(goTimeFmt)
		}

		out.Days = append(out.Days, queryJSONDay{
			Date:  dateInTZ.Format("02 Jan 2006"),
			Hijri: dd.DateInfo.Hijri.Format(),
			Time:  timeStr,
		})
	}

	data, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	fmt.Println(string(data))
	return nil
}
