# tmux-prayer-times

Display the next Islamic prayer time in your tmux status bar.

Uses the [Al Adhan Prayer Times API](https://aladhan.com/prayer-times-api) to fetch accurate prayer times based on your location and preferred calculation method. Results are cached locally so the status bar stays fast.

## Features

- Automatic location detection (IP-based geolocation) or manual coordinates/city
- 24 calculation methods (ISNA, MWL, Umm Al-Qura, and more)
- 7 built-in display formats + custom Go templates
- File-based caching -- no network calls on repeated refreshes
- After-Isha handling -- automatically shows tomorrow's first prayer
- Zero runtime dependencies -- single static Go binary
- Pre-built binaries for Linux and macOS (amd64 + arm64)

## Installation

### With [TPM](https://github.com/tmux-plugins/tpm) (recommended)

Add to your `~/.tmux.conf`:

```bash
set -g @plugin 'smokyabdulrahman/tmux-prayer-times'
```

Press `prefix + I` to install. The plugin will auto-download the correct binary for your platform.

### Manual

```bash
git clone https://github.com/smokyabdulrahman/tmux-prayer-times.git ~/.tmux/plugins/tmux-prayer-times
~/.tmux/plugins/tmux-prayer-times/scripts/install.sh
```

Then add to your `~/.tmux.conf`:

```bash
run-shell ~/.tmux/plugins/tmux-prayer-times/prayer-times.tmux
```

Reload tmux: `tmux source-file ~/.tmux.conf`

### Building from source

Requires Go 1.23+:

```bash
git clone https://github.com/smokyabdulrahman/tmux-prayer-times.git
cd tmux-prayer-times
make build    # builds to bin/tmux-prayer-times
make install  # same as build, ready for local testing
```

## Usage

Add `#{prayer_times}` to your status bar:

```bash
set -g status-right '#{prayer_times}'
```

The plugin replaces `#{prayer_times}` with the next prayer time, refreshed on every tmux status-interval tick.

## Configuration

All options are set as tmux global options in `~/.tmux.conf`.

### Location

```bash
# Option A: City and country
set -g @prayer-times-city "London"
set -g @prayer-times-country "UK"

# Option B: Coordinates
set -g @prayer-times-latitude "51.5074"
set -g @prayer-times-longitude "-0.1278"

# Option C: Omit all location options for automatic IP-based detection
```

### Calculation method

```bash
set -g @prayer-times-method "2"   # ISNA (see table below)
```

If omitted, the API picks a default based on your location.

### Juristic school

```bash
set -g @prayer-times-school "0"   # 0 = Shafi (standard), 1 = Hanafi
```

### Display format

```bash
set -g @prayer-times-format "name-and-time"
```

| Format                     | Example output        |
| -------------------------- | --------------------- |
| `time-remaining`           | `2h 15m`              |
| `next-prayer-time`         | `15:02`               |
| `name-and-time` (default)  | `Asr 15:02`           |
| `name-and-remaining`       | `Asr 2h 15m`          |
| `short-name-and-time`      | `A 15:02`             |
| `short-name-and-remaining` | `A 2h 15m`            |
| `full`                     | `Asr 15:02 (2h 15m)` |

#### Custom templates

Any format string containing `{{` is treated as a Go template:

```bash
set -g @prayer-times-format "{{.Name}} in {{.Hours}}h {{.Minutes}}m"
# Output: "Asr in 2h 15m"
```

Available template fields:

| Field        | Description                        | Example    |
| ------------ | ---------------------------------- | ---------- |
| `.Name`      | Full prayer name                   | `Asr`      |
| `.ShortName` | Abbreviated name                   | `A`        |
| `.Time`      | Formatted prayer time              | `15:02`    |
| `.Remaining` | Human-readable time remaining      | `2h 15m`   |
| `.Hours`     | Whole hours remaining (int)        | `2`        |
| `.Minutes`   | Remaining minutes after hours (int)| `15`       |

### Time format

```bash
set -g @prayer-times-time-format "24h"   # "24h" (default) or "12h"
```

### Icon

```bash
set -g @prayer-times-icon "🕌"
# Output: "🕌 Asr 15:02"
```

### Tracked prayers

```bash
set -g @prayer-times-prayers "Fajr,Dhuhr,Asr,Maghrib,Isha"
```

Default: `Fajr,Sunrise,Dhuhr,Asr,Maghrib,Isha`

All available prayers: `Fajr`, `Sunrise`, `Dhuhr`, `Asr`, `Sunset`, `Maghrib`, `Isha`, `Imsak`, `Midnight`, `Firstthird`, `Lastthird`

### Cache directory

```bash
set -g @prayer-times-cache-dir "/tmp/prayer-cache"
# Default: ~/.cache/tmux-prayer-times/
```

## Calculation Methods

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

You can also print this table from the CLI:

```bash
tmux-prayer-times --list-methods
```

## CLI Reference

The Go binary can be used standalone outside tmux:

```bash
# Basic usage
tmux-prayer-times --city London --country UK

# With coordinates and format
tmux-prayer-times --latitude 51.5074 --longitude -0.1278 --format full

# 12-hour time, Hanafi school, ISNA method
tmux-prayer-times --city Toronto --country CA --method 2 --school 1 --time-format 12h

# Custom template
tmux-prayer-times --city Makkah --country SA --format '{{.ShortName}} {{.Time}} ({{.Remaining}})'

# Info commands
tmux-prayer-times --version
tmux-prayer-times --list-methods
```

## How It Works

1. **TPM loads `prayer-times.tmux`** which replaces `#{prayer_times}` in your status bar with `#(scripts/prayer_times.sh)`
2. **Tmux executes the script** on each status-interval refresh (typically every 2-5 seconds)
3. **The script reads tmux options**, builds CLI flags, and calls the Go binary
4. **The Go binary** checks the local cache first (~7-10ms). On cache miss, it calls the Al Adhan API (~150-1200ms), caches the response, and prints the formatted next prayer time
5. **After Isha**, the binary automatically fetches tomorrow's times to show the next Fajr

## License

[MIT](LICENSE)
