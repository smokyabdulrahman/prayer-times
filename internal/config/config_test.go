package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// tempConfigPath returns a path to a config file inside a temp directory.
func tempConfigPath(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "config.json")
}

// --- Defaults ---

func TestDefaults(t *testing.T) {
	d := Defaults()

	if d.Method == nil {
		t.Fatal("Defaults().Method should not be nil")
	}
	if *d.Method != -1 {
		t.Errorf("Defaults().Method = %d, want -1", *d.Method)
	}

	if d.School == nil {
		t.Fatal("Defaults().School should not be nil")
	}
	if *d.School != -1 {
		t.Errorf("Defaults().School = %d, want -1", *d.School)
	}

	if d.TimeFormat != "24h" {
		t.Errorf("Defaults().TimeFormat = %q, want %q", d.TimeFormat, "24h")
	}

	// Everything else should be zero.
	if d.City != "" {
		t.Errorf("Defaults().City = %q, want empty", d.City)
	}
	if d.Country != "" {
		t.Errorf("Defaults().Country = %q, want empty", d.Country)
	}
	if d.Latitude != 0 {
		t.Errorf("Defaults().Latitude = %f, want 0", d.Latitude)
	}
	if d.Longitude != 0 {
		t.Errorf("Defaults().Longitude = %f, want 0", d.Longitude)
	}
	if d.Prayers != "" {
		t.Errorf("Defaults().Prayers = %q, want empty", d.Prayers)
	}
	if d.CacheDir != "" {
		t.Errorf("Defaults().CacheDir = %q, want empty", d.CacheDir)
	}
}

// --- Dir and Path with XDG ---

func TestDir_XDGConfigHome(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "/tmp/xdg-test")

	dir, err := Dir()
	if err != nil {
		t.Fatalf("Dir() error: %v", err)
	}

	want := filepath.Join("/tmp/xdg-test", "prayer-times")
	if dir != want {
		t.Errorf("Dir() = %q, want %q", dir, want)
	}
}

func TestDir_FallbackToHome(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "")

	dir, err := Dir()
	if err != nil {
		t.Fatalf("Dir() error: %v", err)
	}

	home, _ := os.UserHomeDir()
	want := filepath.Join(home, ".config", "prayer-times")
	if dir != want {
		t.Errorf("Dir() = %q, want %q", dir, want)
	}
}

func TestPath_XDGConfigHome(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "/tmp/xdg-test")

	p, err := Path()
	if err != nil {
		t.Fatalf("Path() error: %v", err)
	}

	want := filepath.Join("/tmp/xdg-test", "prayer-times", "config.json")
	if p != want {
		t.Errorf("Path() = %q, want %q", p, want)
	}
}

// --- LoadFrom ---

func TestLoadFrom_NonExistentFile(t *testing.T) {
	cfg, err := LoadFrom("/no/such/file.json")
	if err != nil {
		t.Fatalf("LoadFrom non-existent should not error, got: %v", err)
	}
	// Should return an empty Config.
	if cfg.City != "" || cfg.Country != "" {
		t.Error("LoadFrom non-existent should return empty config")
	}
	if cfg.Method != nil {
		t.Error("LoadFrom non-existent should have nil Method")
	}
}

func TestLoadFrom_ValidJSON(t *testing.T) {
	path := tempConfigPath(t)

	method := 4
	data := Config{
		City:       "Riyadh",
		Country:    "Saudi Arabia",
		Method:     &method,
		TimeFormat: "12h",
	}
	raw, _ := json.MarshalIndent(data, "", "  ")
	if err := os.WriteFile(path, raw, 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadFrom(path)
	if err != nil {
		t.Fatalf("LoadFrom error: %v", err)
	}

	if cfg.City != "Riyadh" {
		t.Errorf("City = %q, want %q", cfg.City, "Riyadh")
	}
	if cfg.Country != "Saudi Arabia" {
		t.Errorf("Country = %q, want %q", cfg.Country, "Saudi Arabia")
	}
	if cfg.Method == nil || *cfg.Method != 4 {
		t.Errorf("Method = %v, want 4", cfg.Method)
	}
	if cfg.TimeFormat != "12h" {
		t.Errorf("TimeFormat = %q, want %q", cfg.TimeFormat, "12h")
	}
}

func TestLoadFrom_InvalidJSON(t *testing.T) {
	path := tempConfigPath(t)
	if err := os.WriteFile(path, []byte("{bad json"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := LoadFrom(path)
	if err == nil {
		t.Fatal("LoadFrom with invalid JSON should error")
	}
}

func TestLoadFrom_EmptyJSON(t *testing.T) {
	path := tempConfigPath(t)
	if err := os.WriteFile(path, []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadFrom(path)
	if err != nil {
		t.Fatalf("LoadFrom error: %v", err)
	}
	if cfg.City != "" || cfg.Country != "" {
		t.Error("LoadFrom empty JSON should return empty config")
	}
	if cfg.Method != nil {
		t.Error("LoadFrom empty JSON should have nil Method")
	}
}

func TestLoadFrom_MethodZero(t *testing.T) {
	// Method 0 (Jafari) is valid. Ensure it round-trips correctly and
	// is distinguishable from "not set" (nil).
	path := tempConfigPath(t)
	if err := os.WriteFile(path, []byte(`{"method": 0}`), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadFrom(path)
	if err != nil {
		t.Fatalf("LoadFrom error: %v", err)
	}
	if cfg.Method == nil {
		t.Fatal("Method should not be nil for method=0")
	}
	if *cfg.Method != 0 {
		t.Errorf("Method = %d, want 0", *cfg.Method)
	}
}

// --- SaveTo ---

func TestSaveTo_CreatesDirectoryAndFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sub", "dir", "config.json")

	method := 2
	cfg := &Config{
		City:   "London",
		Method: &method,
	}

	if err := cfg.SaveTo(path); err != nil {
		t.Fatalf("SaveTo error: %v", err)
	}

	// Verify file exists and is valid JSON.
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("file not created: %v", err)
	}

	var loaded Config
	if err := json.Unmarshal(data, &loaded); err != nil {
		t.Fatalf("saved file has invalid JSON: %v", err)
	}
	if loaded.City != "London" {
		t.Errorf("loaded City = %q, want %q", loaded.City, "London")
	}
	if loaded.Method == nil || *loaded.Method != 2 {
		t.Errorf("loaded Method = %v, want 2", loaded.Method)
	}
}

func TestSaveTo_TrailingNewline(t *testing.T) {
	path := tempConfigPath(t)
	cfg := &Config{City: "Test"}

	if err := cfg.SaveTo(path); err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(path)
	if len(data) == 0 || data[len(data)-1] != '\n' {
		t.Error("saved file should end with a newline")
	}
}

func TestSaveAndLoad_RoundTrip(t *testing.T) {
	path := tempConfigPath(t)

	method := 0 // Jafari -- tests zero value round-trip.
	school := 1
	original := &Config{
		City:       "Riyadh",
		Country:    "Saudi Arabia",
		Latitude:   24.7136,
		Longitude:  46.6753,
		Method:     &method,
		School:     &school,
		TimeFormat: "12h",
		Prayers:    "Fajr,Dhuhr,Asr,Maghrib,Isha",
		CacheDir:   "/tmp/cache",
	}

	if err := original.SaveTo(path); err != nil {
		t.Fatalf("SaveTo error: %v", err)
	}

	loaded, err := LoadFrom(path)
	if err != nil {
		t.Fatalf("LoadFrom error: %v", err)
	}

	// Compare all fields.
	if loaded.City != original.City {
		t.Errorf("City = %q, want %q", loaded.City, original.City)
	}
	if loaded.Country != original.Country {
		t.Errorf("Country = %q, want %q", loaded.Country, original.Country)
	}
	if loaded.Latitude != original.Latitude {
		t.Errorf("Latitude = %f, want %f", loaded.Latitude, original.Latitude)
	}
	if loaded.Longitude != original.Longitude {
		t.Errorf("Longitude = %f, want %f", loaded.Longitude, original.Longitude)
	}
	if loaded.Method == nil || *loaded.Method != *original.Method {
		t.Errorf("Method = %v, want %d", loaded.Method, *original.Method)
	}
	if loaded.School == nil || *loaded.School != *original.School {
		t.Errorf("School = %v, want %d", loaded.School, *original.School)
	}
	if loaded.TimeFormat != original.TimeFormat {
		t.Errorf("TimeFormat = %q, want %q", loaded.TimeFormat, original.TimeFormat)
	}
	if loaded.Prayers != original.Prayers {
		t.Errorf("Prayers = %q, want %q", loaded.Prayers, original.Prayers)
	}
	if loaded.CacheDir != original.CacheDir {
		t.Errorf("CacheDir = %q, want %q", loaded.CacheDir, original.CacheDir)
	}
}

// --- ResetAt ---

func TestResetAt_DeletesFile(t *testing.T) {
	path := tempConfigPath(t)

	// Create a config file first.
	cfg := &Config{City: "London"}
	if err := cfg.SaveTo(path); err != nil {
		t.Fatal(err)
	}

	if err := ResetAt(path); err != nil {
		t.Fatalf("ResetAt error: %v", err)
	}

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("ResetAt should have deleted the file")
	}
}

func TestResetAt_NonExistentFile(t *testing.T) {
	// Resetting a non-existent file should not error.
	err := ResetAt("/no/such/file.json")
	if err != nil {
		t.Errorf("ResetAt on non-existent file should not error, got: %v", err)
	}
}

// --- Set ---

func TestSet_City(t *testing.T) {
	cfg := &Config{}
	if err := cfg.Set("city", "London"); err != nil {
		t.Fatal(err)
	}
	if cfg.City != "London" {
		t.Errorf("City = %q, want %q", cfg.City, "London")
	}
}

func TestSet_Country(t *testing.T) {
	cfg := &Config{}
	if err := cfg.Set("country", "UK"); err != nil {
		t.Fatal(err)
	}
	if cfg.Country != "UK" {
		t.Errorf("Country = %q, want %q", cfg.Country, "UK")
	}
}

func TestSet_Latitude(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		want    float64
		wantErr bool
	}{
		{"valid positive", "51.5074", 51.5074, false},
		{"valid negative", "-33.8688", -33.8688, false},
		{"zero", "0", 0, false},
		{"boundary 90", "90", 90, false},
		{"boundary -90", "-90", -90, false},
		{"too high", "91", 0, true},
		{"too low", "-91", 0, true},
		{"not a number", "abc", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{}
			err := cfg.Set("latitude", tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("Set(latitude, %q) error = %v, wantErr = %v", tt.value, err, tt.wantErr)
			}
			if !tt.wantErr && cfg.Latitude != tt.want {
				t.Errorf("Latitude = %f, want %f", cfg.Latitude, tt.want)
			}
		})
	}
}

func TestSet_Longitude(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		want    float64
		wantErr bool
	}{
		{"valid positive", "46.6753", 46.6753, false},
		{"valid negative", "-73.5674", -73.5674, false},
		{"zero", "0", 0, false},
		{"boundary 180", "180", 180, false},
		{"boundary -180", "-180", -180, false},
		{"too high", "181", 0, true},
		{"too low", "-181", 0, true},
		{"not a number", "xyz", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{}
			err := cfg.Set("longitude", tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("Set(longitude, %q) error = %v, wantErr = %v", tt.value, err, tt.wantErr)
			}
			if !tt.wantErr && cfg.Longitude != tt.want {
				t.Errorf("Longitude = %f, want %f", cfg.Longitude, tt.want)
			}
		})
	}
}

func TestSet_Method(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		want    int
		wantErr bool
	}{
		{"valid zero (Jafari)", "0", 0, false},
		{"valid 4", "4", 4, false},
		{"valid 23", "23", 23, false},
		{"too high", "24", 0, true},
		{"negative", "-1", 0, true},
		{"not a number", "abc", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{}
			err := cfg.Set("method", tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("Set(method, %q) error = %v, wantErr = %v", tt.value, err, tt.wantErr)
			}
			if !tt.wantErr {
				if cfg.Method == nil {
					t.Fatal("Method should not be nil")
				}
				if *cfg.Method != tt.want {
					t.Errorf("Method = %d, want %d", *cfg.Method, tt.want)
				}
			}
		})
	}
}

func TestSet_School(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		want    int
		wantErr bool
	}{
		{"Shafi", "0", 0, false},
		{"Hanafi", "1", 1, false},
		{"invalid 2", "2", 0, true},
		{"negative", "-1", 0, true},
		{"not a number", "abc", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{}
			err := cfg.Set("school", tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("Set(school, %q) error = %v, wantErr = %v", tt.value, err, tt.wantErr)
			}
			if !tt.wantErr {
				if cfg.School == nil {
					t.Fatal("School should not be nil")
				}
				if *cfg.School != tt.want {
					t.Errorf("School = %d, want %d", *cfg.School, tt.want)
				}
			}
		})
	}
}

func TestSet_TimeFormat(t *testing.T) {
	tests := []struct {
		value   string
		wantErr bool
	}{
		{"12h", false},
		{"24h", false},
		{"invalid", true},
		{"", true},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			cfg := &Config{}
			err := cfg.Set("time_format", tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("Set(time_format, %q) error = %v, wantErr = %v", tt.value, err, tt.wantErr)
			}
			if !tt.wantErr && cfg.TimeFormat != tt.value {
				t.Errorf("TimeFormat = %q, want %q", cfg.TimeFormat, tt.value)
			}
		})
	}
}

func TestSet_Prayers(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"single valid", "Fajr", false},
		{"multiple valid", "Fajr,Dhuhr,Asr,Maghrib,Isha", false},
		{"all prayers", "Fajr,Sunrise,Dhuhr,Asr,Sunset,Maghrib,Isha,Imsak,Midnight,Firstthird,Lastthird", false},
		{"invalid name", "InvalidPrayer", true},
		{"mixed valid/invalid", "Fajr,InvalidPrayer", true},
		{"empty name in list", "Fajr,,Dhuhr", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{}
			err := cfg.Set("prayers", tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("Set(prayers, %q) error = %v, wantErr = %v", tt.value, err, tt.wantErr)
			}
			if !tt.wantErr && cfg.Prayers != tt.value {
				t.Errorf("Prayers = %q, want %q", cfg.Prayers, tt.value)
			}
		})
	}
}

func TestSet_CacheDir(t *testing.T) {
	cfg := &Config{}
	if err := cfg.Set("cache_dir", "/tmp/my-cache"); err != nil {
		t.Fatal(err)
	}
	if cfg.CacheDir != "/tmp/my-cache" {
		t.Errorf("CacheDir = %q, want %q", cfg.CacheDir, "/tmp/my-cache")
	}
}

func TestSet_UnknownKey(t *testing.T) {
	cfg := &Config{}
	err := cfg.Set("unknown_key", "value")
	if err == nil {
		t.Fatal("Set with unknown key should error")
	}
}

// --- Get ---

func TestGet_AllKeys(t *testing.T) {
	method := 4
	school := 1
	cfg := &Config{
		City:       "Riyadh",
		Country:    "Saudi Arabia",
		Latitude:   24.7136,
		Longitude:  46.6753,
		Method:     &method,
		School:     &school,
		TimeFormat: "12h",
		Prayers:    "Fajr,Dhuhr,Asr,Maghrib,Isha",
		CacheDir:   "/tmp/cache",
	}

	tests := []struct {
		key  string
		want string
	}{
		{"city", "Riyadh"},
		{"country", "Saudi Arabia"},
		{"latitude", "24.7136"},
		{"longitude", "46.6753"},
		{"method", "4"},
		{"school", "1"},
		{"time_format", "12h"},
		{"prayers", "Fajr,Dhuhr,Asr,Maghrib,Isha"},
		{"cache_dir", "/tmp/cache"},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			got, err := cfg.Get(tt.key)
			if err != nil {
				t.Fatalf("Get(%q) error: %v", tt.key, err)
			}
			if got != tt.want {
				t.Errorf("Get(%q) = %q, want %q", tt.key, got, tt.want)
			}
		})
	}
}

func TestGet_EmptyConfig(t *testing.T) {
	cfg := &Config{}

	// All values should be empty strings for an empty config.
	for _, key := range ValidKeys {
		got, err := cfg.Get(key)
		if err != nil {
			t.Errorf("Get(%q) error: %v", key, err)
		}
		if got != "" {
			t.Errorf("Get(%q) = %q, want empty for empty config", key, got)
		}
	}
}

func TestGet_UnknownKey(t *testing.T) {
	cfg := &Config{}
	_, err := cfg.Get("unknown_key")
	if err == nil {
		t.Fatal("Get with unknown key should error")
	}
}

func TestGet_MethodZero(t *testing.T) {
	method := 0
	cfg := &Config{Method: &method}

	got, err := cfg.Get("method")
	if err != nil {
		t.Fatal(err)
	}
	if got != "0" {
		t.Errorf("Get(method) = %q, want %q", got, "0")
	}
}

// --- MethodOrDefault / SchoolOrDefault ---

func TestMethodOrDefault_Set(t *testing.T) {
	method := 4
	cfg := &Config{Method: &method}
	if got := cfg.MethodOrDefault(2); got != 4 {
		t.Errorf("MethodOrDefault = %d, want 4", got)
	}
}

func TestMethodOrDefault_Nil(t *testing.T) {
	cfg := &Config{}
	if got := cfg.MethodOrDefault(2); got != 2 {
		t.Errorf("MethodOrDefault = %d, want 2 (default)", got)
	}
}

func TestMethodOrDefault_Zero(t *testing.T) {
	method := 0
	cfg := &Config{Method: &method}
	if got := cfg.MethodOrDefault(2); got != 0 {
		t.Errorf("MethodOrDefault = %d, want 0 (Jafari)", got)
	}
}

func TestSchoolOrDefault_Set(t *testing.T) {
	school := 1
	cfg := &Config{School: &school}
	if got := cfg.SchoolOrDefault(0); got != 1 {
		t.Errorf("SchoolOrDefault = %d, want 1", got)
	}
}

func TestSchoolOrDefault_Nil(t *testing.T) {
	cfg := &Config{}
	if got := cfg.SchoolOrDefault(0); got != 0 {
		t.Errorf("SchoolOrDefault = %d, want 0 (default)", got)
	}
}

func TestSchoolOrDefault_Zero(t *testing.T) {
	school := 0
	cfg := &Config{School: &school}
	if got := cfg.SchoolOrDefault(1); got != 0 {
		t.Errorf("SchoolOrDefault = %d, want 0", got)
	}
}

// --- ValidKeys ---

func TestValidKeys_ContainsExpected(t *testing.T) {
	expected := []string{
		"city", "country", "latitude", "longitude",
		"method", "school", "time_format", "prayers", "cache_dir",
	}

	if len(ValidKeys) != len(expected) {
		t.Errorf("ValidKeys has %d entries, want %d", len(ValidKeys), len(expected))
	}

	keySet := make(map[string]bool)
	for _, k := range ValidKeys {
		keySet[k] = true
	}
	for _, k := range expected {
		if !keySet[k] {
			t.Errorf("ValidKeys missing %q", k)
		}
	}
}

// --- isValidPrayerName ---

func TestIsValidPrayerName(t *testing.T) {
	valid := []string{"Fajr", "Sunrise", "Dhuhr", "Asr", "Sunset", "Maghrib", "Isha", "Imsak", "Midnight", "Firstthird", "Lastthird"}
	for _, name := range valid {
		if !isValidPrayerName(name) {
			t.Errorf("isValidPrayerName(%q) = false, want true", name)
		}
	}

	invalid := []string{"fajr", "FAJR", "Prayer", "", "Invalid"}
	for _, name := range invalid {
		if isValidPrayerName(name) {
			t.Errorf("isValidPrayerName(%q) = true, want false", name)
		}
	}
}

// --- OmitEmpty JSON behavior ---

func TestConfig_OmitEmpty_JSON(t *testing.T) {
	// An empty config should produce minimal JSON.
	cfg := &Config{}
	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}

	got := string(data)
	if got != "{}" {
		t.Errorf("empty config JSON = %s, want {}", got)
	}
}

func TestConfig_OmitEmpty_MethodZero(t *testing.T) {
	// Method 0 should be included in JSON (not omitted).
	method := 0
	cfg := &Config{Method: &method}
	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}

	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatal(err)
	}

	if _, ok := m["method"]; !ok {
		t.Error("method=0 should be present in JSON, but was omitted")
	}
}

// --- Set then Get round-trip ---

func TestSetThenGet_RoundTrip(t *testing.T) {
	tests := []struct {
		key, value string
	}{
		{"city", "Riyadh"},
		{"country", "Saudi Arabia"},
		{"latitude", "24.7136"},
		{"longitude", "46.6753"},
		{"method", "4"},
		{"school", "1"},
		{"time_format", "12h"},
		{"prayers", "Fajr,Dhuhr,Asr,Maghrib,Isha"},
		{"cache_dir", "/tmp/cache"},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			cfg := &Config{}
			if err := cfg.Set(tt.key, tt.value); err != nil {
				t.Fatalf("Set(%q, %q) error: %v", tt.key, tt.value, err)
			}
			got, err := cfg.Get(tt.key)
			if err != nil {
				t.Fatalf("Get(%q) error: %v", tt.key, err)
			}
			if got != tt.value {
				t.Errorf("Set/Get round-trip: got %q, want %q", got, tt.value)
			}
		})
	}
}

// --- Full integration: Set -> SaveTo -> LoadFrom -> Get ---

func TestSetSaveLoadGet_Integration(t *testing.T) {
	path := tempConfigPath(t)

	cfg := &Config{}
	cfg.Set("city", "London")
	cfg.Set("country", "UK")
	cfg.Set("method", "3")
	cfg.Set("time_format", "12h")

	if err := cfg.SaveTo(path); err != nil {
		t.Fatal(err)
	}

	loaded, err := LoadFrom(path)
	if err != nil {
		t.Fatal(err)
	}

	checks := []struct {
		key, want string
	}{
		{"city", "London"},
		{"country", "UK"},
		{"method", "3"},
		{"time_format", "12h"},
	}

	for _, c := range checks {
		got, _ := loaded.Get(c.key)
		if got != c.want {
			t.Errorf("After save/load: Get(%q) = %q, want %q", c.key, got, c.want)
		}
	}
}
