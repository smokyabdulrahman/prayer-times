package cli

import (
	"testing"
	"time"

	"github.com/smokyabdulrahman/prayer-times/internal/api"
	"github.com/smokyabdulrahman/prayer-times/internal/prayer"
)

func TestBuildLocationStr_CityCountry(t *testing.T) {
	loc := resolvedLocation{City: "Riyadh", Country: "Saudi Arabia"}
	result := &fetchResult{Meta: api.Meta{Latitude: 24.7136, Longitude: 46.6753}}

	got := buildLocationStr(loc, result)
	want := "Riyadh, Saudi Arabia"
	if got != want {
		t.Errorf("buildLocationStr() = %q, want %q", got, want)
	}
}

func TestBuildLocationStr_CoordsOnly(t *testing.T) {
	loc := resolvedLocation{Lat: 24.7136, Lon: 46.6753}
	result := &fetchResult{Meta: api.Meta{Latitude: 24.7136, Longitude: 46.6753}}

	got := buildLocationStr(loc, result)
	want := "24.7136, 46.6753"
	if got != want {
		t.Errorf("buildLocationStr() = %q, want %q", got, want)
	}
}

func TestFormatGregorianDate_FromAPI(t *testing.T) {
	now := time.Date(2026, 2, 28, 12, 0, 0, 0, time.UTC)
	result := &fetchResult{
		DateInfo: api.DateInfo{
			Gregorian: api.GregorianDate{
				Day:   "28",
				Month: api.GregorianMonth{Number: 2, En: "February"},
				Year:  "2026",
			},
		},
	}

	got := formatGregorianDate(now, result)
	want := "28 February 2026"
	if got != want {
		t.Errorf("formatGregorianDate() = %q, want %q", got, want)
	}
}

func TestFormatGregorianDate_Fallback(t *testing.T) {
	now := time.Date(2026, 2, 28, 12, 0, 0, 0, time.UTC)
	result := &fetchResult{} // empty DateInfo

	got := formatGregorianDate(now, result)
	want := "28 Feb 2026"
	if got != want {
		t.Errorf("formatGregorianDate() fallback = %q, want %q", got, want)
	}
}

func TestPadRight(t *testing.T) {
	tests := []struct {
		s     string
		width int
		want  string
	}{
		{"Fajr", 7, "Fajr   "},
		{"Maghrib", 7, "Maghrib"},
		{"Isha", 4, "Isha"},
		{"A", 10, "A         "},
	}

	for _, tt := range tests {
		got := padRight(tt.s, tt.width)
		if got != tt.want {
			t.Errorf("padRight(%q, %d) = %q, want %q", tt.s, tt.width, got, tt.want)
		}
	}
}

// TestCurrentAndNext_Consistency verifies that CurrentPrayer and NextPrayer
// are consistent: at any point in time, current should be the prayer before next.
func TestCurrentAndNext_Consistency(t *testing.T) {
	date := time.Date(2026, 2, 28, 0, 0, 0, 0, time.UTC)
	timings := api.Timings{
		Fajr:       "05:17",
		Sunrise:    "06:48",
		Dhuhr:      "12:13",
		Asr:        "15:02",
		Sunset:     "17:39",
		Maghrib:    "17:39",
		Isha:       "19:10",
		Imsak:      "05:07",
		Midnight:   "00:14",
		Firstthird: "22:02",
		Lastthird:  "02:25",
	}
	prayers, err := prayer.ParseTimings(timings, date, time.UTC, prayer.DefaultPrayerNames)
	if err != nil {
		t.Fatal(err)
	}

	// At 13:00 — current=Dhuhr, next=Asr
	now := time.Date(2026, 2, 28, 13, 0, 0, 0, time.UTC)
	current := prayer.CurrentPrayer(prayers, now)
	next := prayer.NextPrayer(prayers, now)

	if current == nil || next == nil {
		t.Fatal("expected both current and next")
	}
	if current.Name != "Dhuhr" {
		t.Errorf("current = %s, want Dhuhr", current.Name)
	}
	if next.Name != "Asr" {
		t.Errorf("next = %s, want Asr", next.Name)
	}
}
