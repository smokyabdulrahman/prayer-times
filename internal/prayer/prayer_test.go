package prayer

import (
	"testing"
	"time"

	"github.com/smokyabdulrahman/prayer-times/internal/api"
)

// helper to build a time.Time on a given date in UTC.
func makeTime(t *testing.T, hour, min int) time.Time {
	t.Helper()
	return time.Date(2026, 2, 28, hour, min, 0, 0, time.UTC)
}

// ---------------------------------------------------------------------------
// parseTimeStr
// ---------------------------------------------------------------------------

func TestParseTimeStr(t *testing.T) {
	date := time.Date(2026, 2, 28, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name    string
		raw     string
		wantH   int
		wantM   int
		wantErr bool
	}{
		{"simple HH:MM", "15:02", 15, 2, false},
		{"midnight", "00:00", 0, 0, false},
		{"with timezone suffix", "15:02 (BST)", 15, 2, false},
		{"with spaces and suffix", "  05:17  (EET) ", 5, 17, false},
		{"invalid format", "bad", 0, 0, true},
		{"empty string", "", 0, 0, true},
		{"missing minute", "15:", 0, 0, true},
		{"non-numeric", "ab:cd", 0, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseTimeStr(tt.raw, date, time.UTC)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("parseTimeStr(%q) expected error, got nil", tt.raw)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseTimeStr(%q) unexpected error: %v", tt.raw, err)
			}
			if got.Hour() != tt.wantH || got.Minute() != tt.wantM {
				t.Errorf("parseTimeStr(%q) = %02d:%02d, want %02d:%02d",
					tt.raw, got.Hour(), got.Minute(), tt.wantH, tt.wantM)
			}
			if got.Year() != 2026 || got.Month() != 2 || got.Day() != 28 {
				t.Errorf("parseTimeStr(%q) wrong date: got %v", tt.raw, got.Format("2006-01-02"))
			}
		})
	}
}

func TestParseTimeStr_Location(t *testing.T) {
	loc, _ := time.LoadLocation("America/New_York")
	date := time.Date(2026, 6, 15, 0, 0, 0, 0, loc)

	got, err := parseTimeStr("12:30", date, loc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Location() != loc {
		t.Errorf("expected location %v, got %v", loc, got.Location())
	}
}

// ---------------------------------------------------------------------------
// ParseTimings
// ---------------------------------------------------------------------------

func sampleTimings() api.Timings {
	return api.Timings{
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
}

func TestParseTimings_DefaultPrayers(t *testing.T) {
	date := time.Date(2026, 2, 28, 0, 0, 0, 0, time.UTC)
	prayers, err := ParseTimings(sampleTimings(), date, time.UTC, DefaultPrayerNames)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(prayers) != len(DefaultPrayerNames) {
		t.Fatalf("expected %d prayers, got %d", len(DefaultPrayerNames), len(prayers))
	}
	for i, name := range DefaultPrayerNames {
		if prayers[i].Name != name {
			t.Errorf("prayer[%d].Name = %q, want %q", i, prayers[i].Name, name)
		}
	}
}

func TestParseTimings_SelectedSubset(t *testing.T) {
	date := time.Date(2026, 2, 28, 0, 0, 0, 0, time.UTC)
	selected := []string{"Fajr", "Maghrib", "Isha"}
	prayers, err := ParseTimings(sampleTimings(), date, time.UTC, selected)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(prayers) != 3 {
		t.Fatalf("expected 3 prayers, got %d", len(prayers))
	}
	if prayers[0].Name != "Fajr" || prayers[1].Name != "Maghrib" || prayers[2].Name != "Isha" {
		t.Errorf("unexpected prayer names: %v", prayers)
	}
}

func TestParseTimings_UnknownPrayer(t *testing.T) {
	date := time.Date(2026, 2, 28, 0, 0, 0, 0, time.UTC)
	_, err := ParseTimings(sampleTimings(), date, time.UTC, []string{"Tahajjud"})
	if err == nil {
		t.Fatal("expected error for unknown prayer, got nil")
	}
}

func TestParseTimings_TimezoneSuffix(t *testing.T) {
	timings := sampleTimings()
	timings.Fajr = "05:17 (BST)"
	timings.Isha = "19:10 (GMT)"

	date := time.Date(2026, 2, 28, 0, 0, 0, 0, time.UTC)
	prayers, err := ParseTimings(timings, date, time.UTC, []string{"Fajr", "Isha"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if prayers[0].Time.Hour() != 5 || prayers[0].Time.Minute() != 17 {
		t.Errorf("Fajr time = %v, want 05:17", prayers[0].Time.Format("15:04"))
	}
	if prayers[1].Time.Hour() != 19 || prayers[1].Time.Minute() != 10 {
		t.Errorf("Isha time = %v, want 19:10", prayers[1].Time.Format("15:04"))
	}
}

// ---------------------------------------------------------------------------
// NextPrayer
// ---------------------------------------------------------------------------

func TestNextPrayer_MiddleOfDay(t *testing.T) {
	date := time.Date(2026, 2, 28, 0, 0, 0, 0, time.UTC)
	prayers, _ := ParseTimings(sampleTimings(), date, time.UTC, DefaultPrayerNames)

	// At 13:00 — Dhuhr (12:13) has passed, next should be Asr (15:02)
	now := time.Date(2026, 2, 28, 13, 0, 0, 0, time.UTC)
	next := NextPrayer(prayers, now)
	if next == nil {
		t.Fatal("expected a next prayer, got nil")
	}
	if next.Name != "Asr" {
		t.Errorf("expected Asr, got %s", next.Name)
	}
}

func TestNextPrayer_BeforeFirstPrayer(t *testing.T) {
	date := time.Date(2026, 2, 28, 0, 0, 0, 0, time.UTC)
	prayers, _ := ParseTimings(sampleTimings(), date, time.UTC, DefaultPrayerNames)

	// At 03:00 — before Fajr (05:17)
	now := time.Date(2026, 2, 28, 3, 0, 0, 0, time.UTC)
	next := NextPrayer(prayers, now)
	if next == nil {
		t.Fatal("expected a next prayer, got nil")
	}
	if next.Name != "Fajr" {
		t.Errorf("expected Fajr, got %s", next.Name)
	}
}

func TestNextPrayer_AfterAllPrayers(t *testing.T) {
	date := time.Date(2026, 2, 28, 0, 0, 0, 0, time.UTC)
	prayers, _ := ParseTimings(sampleTimings(), date, time.UTC, DefaultPrayerNames)

	// At 22:00 — after Isha (19:10)
	now := time.Date(2026, 2, 28, 22, 0, 0, 0, time.UTC)
	next := NextPrayer(prayers, now)
	if next != nil {
		t.Errorf("expected nil after all prayers, got %s", next.Name)
	}
}

func TestNextPrayer_ExactTime(t *testing.T) {
	date := time.Date(2026, 2, 28, 0, 0, 0, 0, time.UTC)
	prayers, _ := ParseTimings(sampleTimings(), date, time.UTC, DefaultPrayerNames)

	// Exactly at Dhuhr time (12:13) — should move to Asr since Dhuhr is not After now
	now := time.Date(2026, 2, 28, 12, 13, 0, 0, time.UTC)
	next := NextPrayer(prayers, now)
	if next == nil {
		t.Fatal("expected a next prayer, got nil")
	}
	if next.Name != "Asr" {
		t.Errorf("expected Asr, got %s", next.Name)
	}
}

func TestNextPrayer_EmptyList(t *testing.T) {
	now := time.Date(2026, 2, 28, 12, 0, 0, 0, time.UTC)
	next := NextPrayer([]Prayer{}, now)
	if next != nil {
		t.Errorf("expected nil for empty prayer list, got %v", next)
	}
}

// ---------------------------------------------------------------------------
// TimeRemaining
// ---------------------------------------------------------------------------

func TestTimeRemaining(t *testing.T) {
	p := Prayer{Name: "Asr", Time: makeTime(t, 15, 2)}
	now := makeTime(t, 13, 0)

	d := TimeRemaining(p, now)
	if d.Hours() < 2.0 || d.Hours() > 2.1 {
		t.Errorf("expected ~2h, got %v", d)
	}
}

func TestTimeRemaining_Negative(t *testing.T) {
	p := Prayer{Name: "Fajr", Time: makeTime(t, 5, 0)}
	now := makeTime(t, 10, 0)

	d := TimeRemaining(p, now)
	if d >= 0 {
		t.Errorf("expected negative duration, got %v", d)
	}
}

// ---------------------------------------------------------------------------
// FormatRemaining
// ---------------------------------------------------------------------------

func TestFormatRemaining(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		want     string
	}{
		{"hours and minutes", 2*time.Hour + 15*time.Minute, "2h 15m"},
		{"only minutes", 45 * time.Minute, "45m"},
		{"exactly one hour", 1 * time.Hour, "1h 0m"},
		{"zero", 0, "0m"},
		{"negative", -30 * time.Minute, "0m"},
		{"large", 10*time.Hour + 59*time.Minute, "10h 59m"},
		{"just over an hour", 1*time.Hour + 1*time.Minute, "1h 1m"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatRemaining(tt.duration)
			if got != tt.want {
				t.Errorf("FormatRemaining(%v) = %q, want %q", tt.duration, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// ShortNames
// ---------------------------------------------------------------------------

func TestShortNames_AllDefaults(t *testing.T) {
	for _, name := range DefaultPrayerNames {
		if _, ok := ShortNames[name]; !ok {
			t.Errorf("ShortNames missing entry for default prayer %q", name)
		}
	}
}

func TestShortNames_AllPrayers(t *testing.T) {
	for _, name := range AllPrayerNames {
		if _, ok := ShortNames[name]; !ok {
			t.Errorf("ShortNames missing entry for prayer %q", name)
		}
	}
}
