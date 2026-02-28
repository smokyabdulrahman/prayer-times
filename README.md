# prayer-times

A full-featured CLI for Islamic prayer times, powered by the [Al Adhan API](https://aladhan.com/prayer-times-api).

Also works as a [tmux](https://github.com/tmux/tmux) status bar plugin via [TPM](https://github.com/tmux-plugins/tpm).

## Features

- Automatic location detection (IP-based geolocation) or manual coordinates/city
- 24 calculation methods (ISNA, MWL, Umm Al-Qura, and more)
- Persistent configuration at `~/.config/prayer-times/config.json`
- Subcommands: today's schedule, next prayer countdown, multi-day list, single-prayer query
- JSON output on every command (`--json`)
- 7 built-in display formats + custom Go templates
- File-based caching -- no network calls on repeated refreshes
- After-Isha handling -- automatically shows tomorrow's first prayer
- Shell completion for bash, zsh, fish, and PowerShell
- Zero runtime dependencies -- single static Go binary
- Pre-built binaries for Linux and macOS (amd64 + arm64)
- Short alias: `pt` (same binary, shorter to type)

## Installation

### Pre-built binaries

Download from the [releases page](https://github.com/smokyabdulrahman/prayer-times/releases).

Each archive contains two binaries: `prayer-times` (full name) and `pt` (short alias).

### go install

Requires Go 1.23+:

```bash
go install github.com/smokyabdulrahman/prayer-times/cmd/prayer-times@latest
go install github.com/smokyabdulrahman/prayer-times/cmd/pt@latest
```

### Build from source

```bash
git clone https://github.com/smokyabdulrahman/prayer-times.git
cd prayer-times
make build     # builds to bin/prayer-times and bin/pt
make install   # installs to $GOPATH/bin/
```

### Tmux plugin (TPM)

If you only want the tmux integration, add to `~/.tmux.conf`:

```bash
set -g @plugin 'smokyabdulrahman/prayer-times'
```

Press `prefix + I` to install. The plugin auto-downloads the correct binary.

## Quick Start

```bash
# Auto-detect location, show today's prayer schedule
prayer-times

# Specify a city
prayer-times --city London --country UK

# Show the next prayer with countdown
prayer-times next

# Short alias works the same way
pt next --city Mecca --country SA
```

## Commands

### `prayer-times` (default: today)

Show today's full prayer schedule with current/next prayer highlighted.

```bash
prayer-times
prayer-times --city Riyadh --country SA
prayer-times --json
```

### `prayer-times next`

Show the next upcoming prayer with a countdown timer. This is the command used by the tmux integration.

```bash
prayer-times next
prayer-times next --format name-and-time
prayer-times next --format "{{.ShortName}} {{.Time}} ({{.Remaining}})"
prayer-times next --json
```

**Display formats:**

| Format                     | Example output        |
| -------------------------- | --------------------- |
| `time-remaining`           | `2h 15m`              |
| `next-prayer-time`         | `15:02`               |
| `name-and-time`            | `Asr 15:02`           |
| `name-and-remaining`       | `Asr 2h 15m`          |
| `short-name-and-time`      | `A 15:02`             |
| `short-name-and-remaining` | `A 2h 15m`            |
| `full` (default)           | `Asr 15:02 (2h 15m)`  |

Custom Go templates are supported -- any format string containing `{{` is treated as a template:

| Field        | Description                         | Example  |
| ------------ | ----------------------------------- | -------- |
| `.Name`      | Full prayer name                    | `Asr`    |
| `.ShortName` | Abbreviated name                    | `A`      |
| `.Time`      | Formatted prayer time               | `15:02`  |
| `.Remaining` | Human-readable time remaining       | `2h 15m` |
| `.Hours`     | Whole hours remaining (int)         | `2`      |
| `.Minutes`   | Remaining minutes after hours (int) | `15`     |

### `prayer-times list [days]`

Show a table of prayer times for multiple days.

```bash
prayer-times list        # 7 days (default)
prayer-times list 14     # 14 days
prayer-times week        # alias for list 7
prayer-times month       # alias for list 30
prayer-times list --json
```

### `prayer-times query <prayer>`

Query a specific prayer's time for today or across multiple days.

```bash
prayer-times query Fajr
prayer-times query Maghrib --days 7
prayer-times query Isha --days month
prayer-times query Fajr --json
```

Valid prayer names: `Fajr`, `Sunrise`, `Dhuhr`, `Asr`, `Sunset`, `Maghrib`, `Isha`, `Imsak`, `Midnight`, `Firstthird`, `Lastthird`

### `prayer-times config`

View and modify persistent configuration.

```bash
prayer-times config                        # show current config
prayer-times config --json                 # show as JSON
prayer-times config set city Riyadh        # set a value
prayer-times config set country "Saudi Arabia"
prayer-times config set method 4
prayer-times config set prayers "Fajr,Dhuhr,Asr,Maghrib,Isha"
prayer-times config set time_format 12h
prayer-times config reset                  # reset to defaults
prayer-times config path                   # print config file path
```

**Valid config keys:**

| Key           | Description                                  | Example                         |
| ------------- | -------------------------------------------- | ------------------------------- |
| `city`        | City name                                    | `London`                        |
| `country`     | Country name or code                         | `UK`                            |
| `latitude`    | Latitude (-90 to 90)                         | `51.5074`                       |
| `longitude`   | Longitude (-180 to 180)                      | `-0.1278`                       |
| `method`      | Calculation method ID (0-23)                 | `2`                             |
| `school`      | Juristic school (0=Shafi, 1=Hanafi)          | `0`                             |
| `time_format` | Time display format                          | `12h` or `24h`                  |
| `prayers`     | Comma-separated list of prayers to track     | `Fajr,Dhuhr,Asr,Maghrib,Isha`  |
| `cache_dir`   | Cache directory path                         | `/tmp/prayer-cache`             |

Config is stored at `~/.config/prayer-times/config.json` (respects `$XDG_CONFIG_HOME`).

### `prayer-times methods`

List all supported calculation methods.

```bash
prayer-times methods
prayer-times methods --json
```

### `prayer-times completion`

Generate shell completion scripts.

```bash
# Bash
source <(prayer-times completion bash)

# Zsh
source <(prayer-times completion zsh)

# Fish
prayer-times completion fish | source

# PowerShell
prayer-times completion powershell | Out-String | Invoke-Expression
```

## Global Flags

These flags work with any subcommand and override config file values:

| Flag             | Description                              |
| ---------------- | ---------------------------------------- |
| `--city`         | Override city                            |
| `--country`      | Override country                         |
| `--latitude`     | Override latitude                        |
| `--longitude`    | Override longitude                       |
| `--method`       | Override calculation method (0-23)       |
| `--school`       | Override school (0=Shafi, 1=Hanafi)      |
| `--prayers`      | Override tracked prayers (comma-separated) |
| `--time-format`  | Override time format (`12h` or `24h`)    |
| `--cache-dir`    | Override cache directory                 |
| `--json`         | Output as JSON                           |

**Priority order:** CLI flags > config file > defaults

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

If omitted, the API picks a default based on your location.

## Tmux Integration

The tmux plugin displays the next prayer time in your status bar using the `#{prayer_times}` interpolation.

### Setup

Add `#{prayer_times}` to your status bar after installing the plugin:

```bash
set -g status-right '#{prayer_times}'
```

### Tmux options

All options are set as tmux global options in `~/.tmux.conf`:

```bash
# Location (pick one approach)
set -g @prayer-times-city "London"
set -g @prayer-times-country "UK"
# -- or --
set -g @prayer-times-latitude "51.5074"
set -g @prayer-times-longitude "-0.1278"
# -- or omit all for automatic IP-based detection --

# Calculation method (ID from table above)
set -g @prayer-times-method "2"

# Juristic school: 0 = Shafi, 1 = Hanafi
set -g @prayer-times-school "0"

# Display format (see next command formats above)
set -g @prayer-times-format "name-and-time"

# Time format: "24h" (default) or "12h"
set -g @prayer-times-time-format "24h"

# Tracked prayers (comma-separated)
set -g @prayer-times-prayers "Fajr,Dhuhr,Asr,Maghrib,Isha"

# Icon prefix
set -g @prayer-times-icon "🕌"

# Cache directory
set -g @prayer-times-cache-dir "/tmp/prayer-cache"
```

### How it works

1. TPM loads `prayer-times.tmux`, which replaces `#{prayer_times}` with `#(scripts/prayer_times.sh)`
2. Tmux executes the script on each status-interval tick
3. The script reads tmux options, builds CLI flags, and calls `prayer-times next`
4. The binary checks the local cache first (~7-10ms). On cache miss, it calls the API (~150-1200ms), caches the response, and prints the next prayer
5. After Isha, it automatically fetches tomorrow's times to show the next Fajr

## Contributing

```bash
git clone https://github.com/smokyabdulrahman/prayer-times.git
cd prayer-times

make build          # build both binaries
make test           # run tests with race detector
make vet            # run go vet
make release        # cross-compile for all platforms
make clean          # remove build artifacts
```

## License

[MIT](LICENSE)
