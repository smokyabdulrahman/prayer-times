package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/smokyabdulrahman/prayer-times/internal/api"
	"github.com/smokyabdulrahman/prayer-times/internal/geo"
)

func sampleAPIResponse() *api.Response {
	return &api.Response{
		Code:   200,
		Status: "OK",
		Data: api.Data{
			Timings: api.Timings{
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
			},
			Meta: api.Meta{
				Latitude:  51.5074,
				Longitude: -0.1278,
				Timezone:  "Europe/London",
				Method:    api.MethodInfo{ID: 2, Name: "ISNA"},
				School:    "STANDARD",
			},
		},
	}
}

// ---------------------------------------------------------------------------
// New
// ---------------------------------------------------------------------------

func TestNew_DefaultDir(t *testing.T) {
	// We can't easily test the default without mocking UserHomeDir,
	// so just test with an explicit dir.
	dir := t.TempDir()
	c, err := New(dir)
	if err != nil {
		t.Fatalf("New(%q) error: %v", dir, err)
	}
	if c == nil {
		t.Fatal("New returned nil")
	}
}

func TestNew_CreatesDirectory(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "subdir", "cache")
	_, err := New(dir)
	if err != nil {
		t.Fatalf("New(%q) error: %v", dir, err)
	}
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Errorf("directory %q was not created", dir)
	}
}

// ---------------------------------------------------------------------------
// SaveTimings / LoadTimings round-trip
// ---------------------------------------------------------------------------

func TestTimings_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	c, _ := New(dir)

	date := time.Date(2026, 2, 28, 0, 0, 0, 0, time.UTC)
	resp := sampleAPIResponse()

	err := c.SaveTimings(date, 51.5074, -0.1278, "", "", 2, 0, resp)
	if err != nil {
		t.Fatalf("SaveTimings error: %v", err)
	}

	entry := c.LoadTimings(date, 51.5074, -0.1278, "", "", 2, 0)
	if entry == nil {
		t.Fatal("LoadTimings returned nil after save")
	}

	if entry.Timings.Fajr != "05:17" {
		t.Errorf("Fajr = %q, want %q", entry.Timings.Fajr, "05:17")
	}
	if entry.Timings.Isha != "19:10" {
		t.Errorf("Isha = %q, want %q", entry.Timings.Isha, "19:10")
	}
	if entry.Meta.Timezone != "Europe/London" {
		t.Errorf("Timezone = %q, want %q", entry.Meta.Timezone, "Europe/London")
	}
}

func TestTimings_CacheMiss(t *testing.T) {
	dir := t.TempDir()
	c, _ := New(dir)

	date := time.Date(2026, 2, 28, 0, 0, 0, 0, time.UTC)
	entry := c.LoadTimings(date, 51.5, -0.1, "", "", 2, 0)
	if entry != nil {
		t.Error("expected nil for cache miss, got entry")
	}
}

func TestTimings_StaleDate(t *testing.T) {
	dir := t.TempDir()
	c, _ := New(dir)

	today := time.Date(2026, 2, 28, 0, 0, 0, 0, time.UTC)
	resp := sampleAPIResponse()

	// Save for today.
	_ = c.SaveTimings(today, 51.5, -0.1, "", "", 2, 0, resp)

	// Load for tomorrow -- should miss because date doesn't match.
	tomorrow := today.AddDate(0, 0, 1)
	entry := c.LoadTimings(tomorrow, 51.5, -0.1, "", "", 2, 0)
	if entry != nil {
		t.Error("expected nil for stale date, got entry")
	}
}

func TestTimings_DifferentParams(t *testing.T) {
	dir := t.TempDir()
	c, _ := New(dir)

	date := time.Date(2026, 2, 28, 0, 0, 0, 0, time.UTC)
	resp := sampleAPIResponse()

	// Save with method=2.
	_ = c.SaveTimings(date, 51.5, -0.1, "", "", 2, 0, resp)

	// Load with method=3 -- different key, should miss.
	entry := c.LoadTimings(date, 51.5, -0.1, "", "", 3, 0)
	if entry != nil {
		t.Error("expected nil for different method, got entry")
	}
}

func TestTimings_CityKey(t *testing.T) {
	dir := t.TempDir()
	c, _ := New(dir)

	date := time.Date(2026, 2, 28, 0, 0, 0, 0, time.UTC)
	resp := sampleAPIResponse()

	// Save with city/country.
	_ = c.SaveTimings(date, 0, 0, "London", "UK", -1, -1, resp)

	entry := c.LoadTimings(date, 0, 0, "London", "UK", -1, -1)
	if entry == nil {
		t.Fatal("expected entry for city-keyed cache, got nil")
	}

	// Different city should miss.
	entry = c.LoadTimings(date, 0, 0, "Paris", "FR", -1, -1)
	if entry != nil {
		t.Error("expected nil for different city, got entry")
	}
}

func TestTimings_CorruptedFile(t *testing.T) {
	dir := t.TempDir()
	c, _ := New(dir)

	date := time.Date(2026, 2, 28, 0, 0, 0, 0, time.UTC)
	resp := sampleAPIResponse()

	// Save a valid entry, then corrupt the file.
	_ = c.SaveTimings(date, 51.5, -0.1, "", "", 2, 0, resp)

	// Find and corrupt the file.
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if filepath.Ext(e.Name()) == ".json" && e.Name() != "geolocation.json" {
			path := filepath.Join(dir, e.Name())
			os.WriteFile(path, []byte("not-json"), 0o644)
		}
	}

	entry := c.LoadTimings(date, 51.5, -0.1, "", "", 2, 0)
	if entry != nil {
		t.Error("expected nil for corrupted cache file, got entry")
	}
}

// ---------------------------------------------------------------------------
// SaveGeo / LoadGeo round-trip
// ---------------------------------------------------------------------------

func TestGeo_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	c, _ := New(dir)

	loc := &geo.Location{
		Latitude:  51.5074,
		Longitude: -0.1278,
		City:      "London",
		Country:   "United Kingdom",
		Timezone:  "Europe/London",
	}

	err := c.SaveGeo(loc)
	if err != nil {
		t.Fatalf("SaveGeo error: %v", err)
	}

	got := c.LoadGeo()
	if got == nil {
		t.Fatal("LoadGeo returned nil after save")
	}
	if got.Latitude != 51.5074 {
		t.Errorf("Latitude = %v, want %v", got.Latitude, 51.5074)
	}
	if got.City != "London" {
		t.Errorf("City = %q, want %q", got.City, "London")
	}
	if got.Timezone != "Europe/London" {
		t.Errorf("Timezone = %q, want %q", got.Timezone, "Europe/London")
	}
}

func TestGeo_CacheMiss(t *testing.T) {
	dir := t.TempDir()
	c, _ := New(dir)

	got := c.LoadGeo()
	if got != nil {
		t.Error("expected nil for geo cache miss, got entry")
	}
}

func TestGeo_ExpiredTTL(t *testing.T) {
	dir := t.TempDir()
	c, _ := New(dir)

	// Write a geo cache entry with a timestamp 25 hours ago (past 24h TTL).
	entry := GeoCacheEntry{
		Location: geo.Location{
			Latitude:  51.5074,
			Longitude: -0.1278,
			City:      "London",
			Country:   "United Kingdom",
			Timezone:  "Europe/London",
		},
		CachedAt: time.Now().Add(-25 * time.Hour),
	}

	data, _ := json.Marshal(entry)
	path := filepath.Join(dir, "geolocation.json")
	os.WriteFile(path, data, 0o644)

	got := c.LoadGeo()
	if got != nil {
		t.Error("expected nil for expired geo cache, got entry")
	}
}

func TestGeo_CorruptedFile(t *testing.T) {
	dir := t.TempDir()
	c, _ := New(dir)

	path := filepath.Join(dir, "geolocation.json")
	os.WriteFile(path, []byte("{bad json"), 0o644)

	got := c.LoadGeo()
	if got != nil {
		t.Error("expected nil for corrupted geo cache, got entry")
	}
}

// ---------------------------------------------------------------------------
// SaveCalendar / LoadCalendar round-trip
// ---------------------------------------------------------------------------

func sampleCalendarResponse(days int) *api.CalendarResponse {
	data := make([]api.Data, days)
	for i := 0; i < days; i++ {
		data[i] = api.Data{
			Timings: api.Timings{
				Fajr:    "05:17",
				Sunrise: "06:48",
				Dhuhr:   "12:13",
				Asr:     "15:02",
				Maghrib: "17:39",
				Isha:    "19:10",
			},
			Date: api.DateInfo{
				Readable: fmt.Sprintf("%d Feb 2026", i+1),
			},
			Meta: api.Meta{
				Latitude:  51.5074,
				Longitude: -0.1278,
				Timezone:  "Europe/London",
				Method:    api.MethodInfo{ID: 2, Name: "ISNA"},
				School:    "STANDARD",
			},
		}
	}
	return &api.CalendarResponse{
		Code:   200,
		Status: "OK",
		Data:   data,
	}
}

func TestCalendar_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	c, _ := New(dir)

	resp := sampleCalendarResponse(28)

	err := c.SaveCalendar(2026, 2, 51.5074, -0.1278, "", "", 2, 0, resp)
	if err != nil {
		t.Fatalf("SaveCalendar error: %v", err)
	}

	entry := c.LoadCalendar(2026, 2, 51.5074, -0.1278, "", "", 2, 0)
	if entry == nil {
		t.Fatal("LoadCalendar returned nil after save")
	}

	if entry.Year != 2026 {
		t.Errorf("Year = %d, want %d", entry.Year, 2026)
	}
	if entry.Month != 2 {
		t.Errorf("Month = %d, want %d", entry.Month, 2)
	}
	if len(entry.Days) != 28 {
		t.Errorf("Days count = %d, want 28", len(entry.Days))
	}
	if entry.Days[0].Timings.Fajr != "05:17" {
		t.Errorf("Day[0].Fajr = %q, want %q", entry.Days[0].Timings.Fajr, "05:17")
	}
	if entry.Days[0].Meta.Timezone != "Europe/London" {
		t.Errorf("Timezone = %q, want %q", entry.Days[0].Meta.Timezone, "Europe/London")
	}
}

func TestCalendar_CacheMiss(t *testing.T) {
	dir := t.TempDir()
	c, _ := New(dir)

	entry := c.LoadCalendar(2026, 2, 51.5, -0.1, "", "", 2, 0)
	if entry != nil {
		t.Error("expected nil for calendar cache miss, got entry")
	}
}

func TestCalendar_DifferentMonth(t *testing.T) {
	dir := t.TempDir()
	c, _ := New(dir)

	resp := sampleCalendarResponse(28)
	_ = c.SaveCalendar(2026, 2, 51.5, -0.1, "", "", 2, 0, resp)

	// Load for a different month -- should miss.
	entry := c.LoadCalendar(2026, 3, 51.5, -0.1, "", "", 2, 0)
	if entry != nil {
		t.Error("expected nil for different month, got entry")
	}
}

func TestCalendar_DifferentYear(t *testing.T) {
	dir := t.TempDir()
	c, _ := New(dir)

	resp := sampleCalendarResponse(28)
	_ = c.SaveCalendar(2026, 2, 51.5, -0.1, "", "", 2, 0, resp)

	// Load for a different year -- should miss.
	entry := c.LoadCalendar(2027, 2, 51.5, -0.1, "", "", 2, 0)
	if entry != nil {
		t.Error("expected nil for different year, got entry")
	}
}

func TestCalendar_DifferentParams(t *testing.T) {
	dir := t.TempDir()
	c, _ := New(dir)

	resp := sampleCalendarResponse(28)
	_ = c.SaveCalendar(2026, 2, 51.5, -0.1, "", "", 2, 0, resp)

	// Load with different method -- should miss.
	entry := c.LoadCalendar(2026, 2, 51.5, -0.1, "", "", 3, 0)
	if entry != nil {
		t.Error("expected nil for different method, got entry")
	}
}

func TestCalendar_CityKey(t *testing.T) {
	dir := t.TempDir()
	c, _ := New(dir)

	resp := sampleCalendarResponse(28)
	_ = c.SaveCalendar(2026, 2, 0, 0, "London", "UK", -1, -1, resp)

	entry := c.LoadCalendar(2026, 2, 0, 0, "London", "UK", -1, -1)
	if entry == nil {
		t.Fatal("expected entry for city-keyed calendar cache, got nil")
	}

	// Different city should miss.
	entry = c.LoadCalendar(2026, 2, 0, 0, "Paris", "FR", -1, -1)
	if entry != nil {
		t.Error("expected nil for different city, got entry")
	}
}

func TestCalendar_CorruptedFile(t *testing.T) {
	dir := t.TempDir()
	c, _ := New(dir)

	resp := sampleCalendarResponse(28)
	_ = c.SaveCalendar(2026, 2, 51.5, -0.1, "", "", 2, 0, resp)

	// Find and corrupt the calendar cache file.
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if filepath.Ext(e.Name()) == ".json" && e.Name() != "geolocation.json" &&
			len(e.Name()) > 9 && e.Name()[:9] == "calendar_" {
			path := filepath.Join(dir, e.Name())
			os.WriteFile(path, []byte("not-json"), 0o644)
		}
	}

	entry := c.LoadCalendar(2026, 2, 51.5, -0.1, "", "", 2, 0)
	if entry != nil {
		t.Error("expected nil for corrupted calendar cache file, got entry")
	}
}

// ---------------------------------------------------------------------------
// calendarKey
// ---------------------------------------------------------------------------

func TestCalendarKey_Deterministic(t *testing.T) {
	k1 := calendarKey(2026, 2, 51.5, -0.1, "", "", 2, 0)
	k2 := calendarKey(2026, 2, 51.5, -0.1, "", "", 2, 0)
	if k1 != k2 {
		t.Errorf("calendarKey not deterministic: %q != %q", k1, k2)
	}
}

func TestCalendarKey_DifferentInputs(t *testing.T) {
	k1 := calendarKey(2026, 2, 51.5, -0.1, "", "", 2, 0)
	k2 := calendarKey(2026, 2, 51.5, -0.1, "", "", 3, 0)  // different method
	k3 := calendarKey(2026, 3, 51.5, -0.1, "", "", 2, 0)  // different month
	k4 := calendarKey(2027, 2, 51.5, -0.1, "", "", 2, 0)  // different year
	k5 := calendarKey(2026, 2, 40.7, -74.0, "", "", 2, 0) // different coords

	keys := []string{k1, k2, k3, k4, k5}
	seen := make(map[string]bool)
	for _, k := range keys {
		if seen[k] {
			t.Errorf("duplicate calendar key: %q", k)
		}
		seen[k] = true
	}
}

// ---------------------------------------------------------------------------
// cacheKey
// ---------------------------------------------------------------------------

func TestCacheKey_Deterministic(t *testing.T) {
	k1 := cacheKey("2026-02-28", 51.5, -0.1, "", "", 2, 0)
	k2 := cacheKey("2026-02-28", 51.5, -0.1, "", "", 2, 0)
	if k1 != k2 {
		t.Errorf("cacheKey not deterministic: %q != %q", k1, k2)
	}
}

func TestCacheKey_DifferentInputs(t *testing.T) {
	k1 := cacheKey("2026-02-28", 51.5, -0.1, "", "", 2, 0)
	k2 := cacheKey("2026-02-28", 51.5, -0.1, "", "", 3, 0)  // different method
	k3 := cacheKey("2026-03-01", 51.5, -0.1, "", "", 2, 0)  // different date
	k4 := cacheKey("2026-02-28", 40.7, -74.0, "", "", 2, 0) // different coords

	keys := []string{k1, k2, k3, k4}
	seen := make(map[string]bool)
	for _, k := range keys {
		if seen[k] {
			t.Errorf("duplicate cache key: %q", k)
		}
		seen[k] = true
	}
}

func TestCacheKey_Length(t *testing.T) {
	k := cacheKey("2026-02-28", 51.5, -0.1, "", "", 2, 0)
	// 8 bytes -> 16 hex chars
	if len(k) != 16 {
		t.Errorf("cacheKey length = %d, want 16", len(k))
	}
}
