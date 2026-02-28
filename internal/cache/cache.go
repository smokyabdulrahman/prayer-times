package cache

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/aalrahma/tmux-prayer-times/internal/api"
	"github.com/aalrahma/tmux-prayer-times/internal/geo"
)

const (
	prayerCacheFile = "timings_%s.json" // keyed by hash
	geoCacheFile    = "geolocation.json"
	geoTTL          = 24 * time.Hour
)

// Cache provides file-based caching for prayer times and geolocation data.
type Cache struct {
	dir string
}

// PrayerCacheEntry stores a day's prayer times along with metadata for validation.
type PrayerCacheEntry struct {
	Date    string      `json:"date"` // YYYY-MM-DD
	Method  int         `json:"method"`
	School  int         `json:"school"`
	Timings api.Timings `json:"timings"`
	Meta    api.Meta    `json:"meta"`
}

// GeoCacheEntry stores a cached geolocation result with a timestamp.
type GeoCacheEntry struct {
	Location geo.Location `json:"location"`
	CachedAt time.Time    `json:"cached_at"`
}

// New creates a Cache rooted at the given directory.
// If dir is empty, it defaults to ~/.cache/tmux-prayer-times/.
func New(dir string) (*Cache, error) {
	if dir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("cannot determine home directory: %w", err)
		}
		dir = filepath.Join(home, ".cache", "tmux-prayer-times")
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("cannot create cache directory %s: %w", dir, err)
	}

	return &Cache{dir: dir}, nil
}

// cacheKey builds a deterministic hash from the parameters that affect prayer times.
// This ensures different locations/methods/schools get separate cache files.
func cacheKey(date string, lat, lon float64, city, country string, method, school int) string {
	raw := fmt.Sprintf("%s|%.6f|%.6f|%s|%s|%d|%d", date, lat, lon, city, country, method, school)
	h := sha256.Sum256([]byte(raw))
	return fmt.Sprintf("%x", h[:8]) // 16 hex chars is plenty for uniqueness
}

// LoadTimings attempts to read cached prayer times for the given parameters.
// Returns nil if the cache is missing or stale (wrong date).
func (c *Cache) LoadTimings(date time.Time, lat, lon float64, city, country string, method, school int) *PrayerCacheEntry {
	dateStr := date.Format("2006-01-02")
	key := cacheKey(dateStr, lat, lon, city, country, method, school)
	path := filepath.Join(c.dir, fmt.Sprintf(prayerCacheFile, key))

	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	var entry PrayerCacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil
	}

	// Validate the date matches -- stale cache for a previous day is useless.
	if entry.Date != dateStr {
		return nil
	}

	return &entry
}

// SaveTimings writes prayer times to the cache.
func (c *Cache) SaveTimings(date time.Time, lat, lon float64, city, country string, method, school int, resp *api.Response) error {
	dateStr := date.Format("2006-01-02")
	key := cacheKey(dateStr, lat, lon, city, country, method, school)
	path := filepath.Join(c.dir, fmt.Sprintf(prayerCacheFile, key))

	entry := PrayerCacheEntry{
		Date:    dateStr,
		Method:  method,
		School:  school,
		Timings: resp.Data.Timings,
		Meta:    resp.Data.Meta,
	}

	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal cache entry: %w", err)
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}

	return nil
}

// LoadGeo attempts to read a cached geolocation result.
// Returns nil if the cache is missing or older than the TTL (24 hours).
func (c *Cache) LoadGeo() *geo.Location {
	path := filepath.Join(c.dir, geoCacheFile)

	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	var entry GeoCacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil
	}

	if time.Since(entry.CachedAt) > geoTTL {
		return nil
	}

	return &entry.Location
}

// SaveGeo writes a geolocation result to the cache.
func (c *Cache) SaveGeo(loc *geo.Location) error {
	path := filepath.Join(c.dir, geoCacheFile)

	entry := GeoCacheEntry{
		Location: *loc,
		CachedAt: time.Now(),
	}

	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal geo cache: %w", err)
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("failed to write geo cache: %w", err)
	}

	return nil
}
