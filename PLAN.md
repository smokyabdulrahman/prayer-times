# Tmux Prayer Times Plugin -- Implementation Plan

## Language Recommendation: Bash + Go

### Why not pure Bash?

A pure-Bash plugin is possible (using `curl` + `jq`), and many tmux plugins take this
approach. However, this plugin has enough complexity (HTTP requests, JSON parsing, time
comparison, caching, timezone math) that pure Bash becomes fragile and hard to maintain.
Parsing JSON with `jq` is fine, but it becomes an external dependency that not everyone
has installed.

### Why Go?

| Criteria                         | Go                              | Python                    | Bash                           |
| -------------------------------- | ------------------------------- | ------------------------- | ------------------------------ |
| **Single binary, zero deps**     | Yes                             | No (needs python3)        | Needs curl, jq, date, etc.    |
| **Cross-platform (Linux/macOS)** | Compiles to static binary       | Mostly                    | Shell differences can bite     |
| **JSON parsing**                 | Built-in `encoding/json`        | Built-in                  | Needs jq                       |
| **Time/timezone handling**       | Excellent `time` stdlib         | Excellent                 | Painful                        |
| **HTTP client**                  | Built-in `net/http`             | Built-in                  | Needs curl/wget                |
| **Startup speed**                | ~5-10ms (compiled binary)       | ~50-100ms (interpreter)   | ~50-200ms (subshells)          |
| **Distribution**                 | Pre-built binaries via Releases | pip or bundled             | Just scripts                   |

Go is the best fit because:

1. **Zero dependencies** -- compiles to a single static binary. Users don't need to
   install Go, Python, curl, jq, or anything else.
2. **Fast** -- tmux status bar scripts run every few seconds. A Go binary starts in
   milliseconds; Python/Bash are slower.
3. **Robust time handling** -- comparing "is it past Asr time?" in Bash is error-prone.
   Go's `time` package handles timezones, parsing, and comparison natively.
4. **Easy distribution** -- pre-built binaries for linux/amd64, linux/arm64,
   darwin/amd64, darwin/arm64. The Bash wrapper in the plugin auto-downloads the
   correct one.

### Architecture: Bash wrapper + Go binary

- **Bash scripts** handle tmux integration (TPM hooks, reading tmux options, calling the binary)
- **Go binary** handles all the logic (API calls, caching, time comparison, formatting output)

---

## How Tmux Plugin Integration Works

Tmux plugins work by injecting **format strings** into the status bar. Here's the pattern:

### 1. Format string interpolation

The plugin registers custom format placeholders that users add to their `status-right`
or `status-left`:

```bash
# In .tmux.conf
set -g status-right '#{prayer_times}'
```

The plugin's main `.tmux` file intercepts these placeholders and replaces them with the
output of a script.

### 2. Plugin file structure (TPM convention)

```
tmux-prayer-times/
  prayer-times.tmux          # Entry point -- TPM sources this
  scripts/
    prayer_times.sh           # Bash wrapper that tmux calls
  bin/
    tmux-prayer-times         # Go binary (downloaded or compiled)
  internal/                   # Go source code
    ...
  go.mod
  go.sum
  LICENSE
  README.md
```

### 3. User configuration via tmux options

```bash
# .tmux.conf examples:
set -g @prayer-times-city "London"
set -g @prayer-times-country "UK"
# OR
set -g @prayer-times-latitude "51.5074"
set -g @prayer-times-longitude "-0.1278"

set -g @prayer-times-method "2"              # ISNA
set -g @prayer-times-school "1"              # Hanafi
set -g @prayer-times-format "name-and-time"  # Display format
set -g @prayer-times-prayers "Fajr,Sunrise,Dhuhr,Asr,Maghrib,Isha"
```

### 4. Example outputs the user can configure

| Format option              | Example output        |
| -------------------------- | --------------------- |
| `time-remaining`           | `2h 15m`              |
| `next-prayer-time`         | `15:02`               |
| `name-and-time`            | `Asr 15:02`           |
| `name-and-remaining`       | `Asr 2h 15m`          |
| `short-name-and-time`      | `A 15:02`             |
| `short-name-and-remaining` | `A 2h 15m`            |
| `full`                     | `Asr 15:02 (2h 15m)` |

---

## Al Adhan API Reference (Relevant Endpoints)

**Base URL:** `https://api.aladhan.com/v1`

All endpoints use **HTTP GET**. No authentication required.

### Timings by Coordinates

```
GET /v1/timings/{DD-MM-YYYY}?latitude={lat}&longitude={lon}&method={id}&school={0|1}
```

### Timings by City

```
GET /v1/timingsByCity/{DD-MM-YYYY}?city={name}&country={code}&method={id}&school={0|1}
```

### Common Query Parameters

| Parameter                   | Required | Description                                                   |
| --------------------------- | -------- | ------------------------------------------------------------- |
| `latitude` / `longitude`    | Yes*     | Decimal coordinates (for `/timings`)                          |
| `city` / `country`          | Yes*     | City name and country code (for `/timingsByCity`)             |
| `method`                    | No       | Calculation method ID (0-23, 99). Default varies by location. |
| `school`                    | No       | `0` = Shafi (standard), `1` = Hanafi                          |
| `midnightMode`              | No       | `0` = Standard, `1` = Jafari                                  |
| `latitudeAdjustmentMethod`  | No       | `1` = Middle of Night, `2` = One Seventh, `3` = Angle Based   |
| `tune`                      | No       | Comma-separated minute offsets for each prayer                 |
| `iso8601`                   | No       | `true` for ISO 8601 formatted times                           |

### Response Structure

```json
{
  "code": 200,
  "status": "OK",
  "data": {
    "timings": {
      "Fajr": "05:17",
      "Sunrise": "06:48",
      "Dhuhr": "12:13",
      "Asr": "15:02",
      "Sunset": "17:39",
      "Maghrib": "17:39",
      "Isha": "19:10",
      "Imsak": "05:07",
      "Midnight": "00:14",
      "Firstthird": "22:02",
      "Lastthird": "02:25"
    },
    "date": {
      "readable": "28 Feb 2026",
      "timestamp": "1772262000",
      "hijri": { "..." },
      "gregorian": { "..." }
    },
    "meta": {
      "latitude": 51.5074,
      "longitude": -0.1278,
      "timezone": "Europe/London",
      "method": { "id": 2, "name": "..." },
      "school": "STANDARD"
    }
  }
}
```

### Calculation Methods

| ID | Name                                            |
| -- | ----------------------------------------------- |
| 0  | Shia Ithna-Ashari (Jafari)                      |
| 1  | University of Islamic Sciences, Karachi          |
| 2  | Islamic Society of North America (ISNA)          |
| 3  | Muslim World League (MWL)                        |
| 4  | Umm Al-Qura University, Makkah                  |
| 5  | Egyptian General Authority of Survey             |
| 7  | Institute of Geophysics, University of Tehran    |
| 8  | Gulf Region                                      |
| 9  | Kuwait                                           |
| 10 | Qatar                                            |
| 11 | Majlis Ugama Islam Singapura (Singapore)         |
| 12 | Union Organization Islamic de France             |
| 13 | Diyanet Isleri Baskanligi, Turkey (experimental) |
| 14 | Spiritual Administration of Muslims of Russia    |
| 15 | Moonsighting Committee Worldwide                 |
| 16 | Dubai (experimental)                             |
| 17 | JAKIM (Malaysia)                                 |
| 18 | Tunisia                                          |
| 19 | Algeria                                          |
| 20 | KEMENAG (Indonesia)                              |
| 21 | Morocco                                          |
| 22 | Comunidade Islamica de Lisboa (Portugal)         |
| 23 | Ministry of Awqaf, Jordan                        |

---

## Phased Implementation

### Phase 1: Scaffold & Core Go Binary

**Goal:** A working Go CLI that fetches prayer times and prints the next prayer.

**Tasks:**

1. Initialize Go module (`go mod init github.com/<user>/tmux-prayer-times`)
2. Implement the API client:
   - `GET /v1/timings/{date}` by coordinates
   - `GET /v1/timingsByCity/{date}` by city/country
   - Parse JSON response into Go structs
3. Implement IP-based geolocation fallback (e.g., `http://ip-api.com/json/` -- free,
   no API key required)
4. Implement core logic:
   - Parse all prayer times from the API response
   - Filter to only the user's selected prayers
   - Default tracked prayers: Fajr, Sunrise, Dhuhr, Asr, Maghrib, Isha
   - All available prayers are configurable: Fajr, Sunrise, Dhuhr, Asr, Sunset,
     Maghrib, Isha, Imsak, Midnight, Firstthird, Lastthird
   - Compare against current time to find the **next** prayer
   - Handle edge case: after Isha, next prayer is tomorrow's Fajr
5. Implement CLI flags:
   - `--latitude`, `--longitude` / `--city`, `--country`
   - `--method`, `--school`
   - `--format` (the display format)
   - `--prayers` (comma-separated list of prayers to track)
6. Output the formatted string to stdout

**Deliverable:** Running `tmux-prayer-times --city London --country UK --format name-and-time`
prints something like `Asr 15:02`.

---

### Phase 2: Caching Layer

**Goal:** Avoid hitting the API on every tmux status bar refresh (every 1-5 seconds).

**Tasks:**

1. Implement file-based cache:
   - Cache location: `~/.cache/tmux-prayer-times/`
   - Cache key: date + location hash + method + school
   - Cache format: JSON file with the day's prayer times
2. On startup:
   - If cache exists and is for today: read from cache (no network call)
   - If cache is stale or missing: fetch from API, write cache
3. Cache the auto-detected location (IP geolocation) separately with a longer TTL
   (24 hours)
4. Add `--cache-dir` flag to override cache location

**Deliverable:** Subsequent invocations within the same day are instant (no network).

---

### Phase 3: Tmux Plugin Integration

**Goal:** Full TPM-compatible plugin that users can install with one line.

**Tasks:**

1. Create `prayer-times.tmux` (the TPM entry point):
   - Read all `@prayer-times-*` tmux options
   - Register the `#{prayer_times}` format string interpolation
   - Set up the update hook
2. Create `scripts/prayer_times.sh`:
   - Read tmux options and pass them as flags to the Go binary
   - Auto-download the correct Go binary on first run (from GitHub Releases)
   - Handle binary not found / download failure gracefully
3. Create `scripts/install.sh`:
   - Detect OS and architecture
   - Download pre-built binary from GitHub Releases
   - Make it executable in `bin/`
4. Support manual installation (git clone + source in tmux.conf)

**Deliverable:** User can add `set -g @plugin 'username/tmux-prayer-times'` to their
`.tmux.conf` and see prayer times in their status bar after `prefix + I`.

---

### Phase 4: Display Format System & Configuration

**Goal:** Rich, configurable display options.

**Tasks:**

1. Implement all format modes:
   - `time-remaining` -- `2h 15m`
   - `next-prayer-time` -- `15:02`
   - `name-and-time` -- `Asr 15:02`
   - `name-and-remaining` -- `Asr 2h 15m`
   - `short-name-and-time` -- `A 15:02`
   - `short-name-and-remaining` -- `A 2h 15m`
   - `full` -- `Asr 15:02 (2h 15m)`
   - `custom` -- user-defined Go template string, e.g., `{{.Name}} in {{.Remaining}}`
2. Implement prayer name mappings:
   - Full: `Fajr`, `Dhuhr`, `Asr`, `Maghrib`, `Isha`, `Sunrise`
   - Short: `F`, `D`, `A`, `M`, `I`, `S`
3. Add `@prayer-times-icon` option (prepend a configurable icon/emoji)
4. Add `@prayer-times-time-format` option (`12h` or `24h`)
5. Add `@prayer-times-prayers` option to select which prayers to track

**Deliverable:** Users have full control over what the status bar shows.

---

### Phase 5: Build & Release Pipeline

**Goal:** Automated builds and easy distribution.

**Tasks:**

1. Set up GitHub Actions workflow:
   - Build Go binary for: `linux/amd64`, `linux/arm64`, `darwin/amd64`, `darwin/arm64`
   - Run tests on each push
   - Create GitHub Release with pre-built binaries on tag push
2. Add Makefile for local development:
   - `make build` -- build for current platform
   - `make release` -- cross-compile all platforms
   - `make test` -- run tests
   - `make install` -- install locally for testing
3. Write the install script that auto-detects platform and downloads the right binary

**Deliverable:** Tagging `v0.1.0` and pushing produces downloadable binaries automatically.

---

### Phase 6: Polish & Documentation

**Goal:** Production-ready plugin.

**Tasks:**

1. Write comprehensive README:
   - Installation (TPM and manual)
   - Configuration reference (all options with defaults)
   - Screenshots / output examples
   - Supported calculation methods table
2. Robust error handling:
   - Network failure: show cached data or a fallback message
   - Invalid config: show helpful error in status bar (not a crash)
   - API rate limiting: respect it gracefully
3. Add `--version` flag
4. Add `--list-methods` flag (prints the calculation methods table)
5. Handle the "after Isha" edge case robustly (fetch or cache tomorrow's Fajr)
6. Optional Ramadan awareness (show Imsak time during Ramadan month)

**Deliverable:** A polished v1.0 release.

---

## Final File Structure

```
tmux-prayer-times/
  prayer-times.tmux                # TPM entry point
  scripts/
    prayer_times.sh                # Bash wrapper (reads tmux opts, calls binary)
    install.sh                     # Binary downloader
  cmd/
    tmux-prayer-times/
      main.go                      # CLI entry point
  internal/
    api/
      client.go                    # Al Adhan API client
      types.go                     # Response structs
    cache/
      cache.go                     # File-based caching
    geo/
      detect.go                    # IP-based geolocation
    prayer/
      prayer.go                    # Core logic (next prayer, time remaining)
      format.go                    # Output formatting
  .github/
    workflows/
      release.yml                  # CI/CD
  go.mod
  go.sum
  Makefile
  LICENSE
  README.md
  PLAN.md
```
