package cli

import (
	"fmt"

	"github.com/smokyabdulrahman/prayer-times/internal/config"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// Global flags shared across all subcommands.
var (
	FlagCity       string
	FlagCountry    string
	FlagLatitude   float64
	FlagLongitude  float64
	FlagMethod     int
	FlagSchool     int
	FlagJSON       bool
	FlagCacheDir   string
	FlagTimeFormat string
)

// loadedConfig holds the config loaded during PersistentPreRunE.
// Available to all subcommand handlers.
var loadedConfig *config.Config

// NewRootCmd creates the root command for the prayer-times CLI.
// The version parameter is set by the calling binary via ldflags.
func NewRootCmd(version string) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:     "prayer-times",
		Short:   "Islamic prayer times CLI",
		Long:    "A full-featured CLI for Islamic prayer times powered by the Al Adhan API.",
		Version: version,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}
			loadedConfig = cfg
			return nil
		},
		// Default action: show today's prayer schedule.
		RunE:          runToday,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	// Register global persistent flags.
	pf := rootCmd.PersistentFlags()
	pf.StringVar(&FlagCity, "city", "", "Override city (takes precedence over config)")
	pf.StringVar(&FlagCountry, "country", "", "Override country")
	pf.Float64Var(&FlagLatitude, "latitude", 0, "Override latitude")
	pf.Float64Var(&FlagLongitude, "longitude", 0, "Override longitude")
	pf.IntVar(&FlagMethod, "method", -1, "Override calculation method (0-23)")
	pf.IntVar(&FlagSchool, "school", -1, "Override school (0=Shafi, 1=Hanafi)")
	pf.BoolVar(&FlagJSON, "json", false, "Output as JSON (where supported)")
	pf.StringVar(&FlagCacheDir, "cache-dir", "", "Cache directory (default: ~/.cache/prayer-times/)")
	pf.StringVar(&FlagTimeFormat, "time-format", "", "Time format: 12h or 24h (overrides config)")

	// Register subcommands.
	rootCmd.AddCommand(newNextCmd())
	rootCmd.AddCommand(newListCmd())
	rootCmd.AddCommand(newWeekCmd())
	rootCmd.AddCommand(newMonthCmd())
	rootCmd.AddCommand(newQueryCmd())
	rootCmd.AddCommand(newConfigCmd())
	rootCmd.AddCommand(newMethodsCmd())

	return rootCmd
}

// PrintVersion prints the version string in the expected format.
func PrintVersion(version string) string {
	return fmt.Sprintf("prayer-times %s\n", version)
}

// effectiveConfig returns the merged configuration values,
// applying the priority: CLI flags > config file > defaults.
// It uses cobra's Changed() to detect whether a flag was explicitly set.
func effectiveConfig(cmd *cobra.Command) *config.Config {
	cfg := loadedConfig
	if cfg == nil {
		empty := config.Config{}
		cfg = &empty
	}

	// Apply defaults for unset config values.
	defaults := config.Defaults()

	// Merge: CLI flags override config, config overrides defaults.
	// For each field, if the CLI flag was explicitly set, use it.
	// Otherwise use the config value. If config is also unset, use default.

	flags := cmd.Flags()
	root := cmd.Root().PersistentFlags()

	if flagWasSet(flags, root, "city") {
		cfg.City = FlagCity
	}
	if flagWasSet(flags, root, "country") {
		cfg.Country = FlagCountry
	}
	if flagWasSet(flags, root, "latitude") {
		cfg.Latitude = FlagLatitude
	}
	if flagWasSet(flags, root, "longitude") {
		cfg.Longitude = FlagLongitude
	}
	if flagWasSet(flags, root, "method") {
		cfg.Method = &FlagMethod
	} else if cfg.Method == nil {
		cfg.Method = defaults.Method
	}
	if flagWasSet(flags, root, "school") {
		cfg.School = &FlagSchool
	} else if cfg.School == nil {
		cfg.School = defaults.School
	}
	if flagWasSet(flags, root, "cache-dir") {
		cfg.CacheDir = FlagCacheDir
	}

	// Time format: CLI flag > config > default ("24h").
	if flagWasSet(flags, root, "time-format") {
		cfg.TimeFormat = FlagTimeFormat
	}
	if cfg.TimeFormat == "" {
		cfg.TimeFormat = defaults.TimeFormat
	}

	return cfg
}

// flagWasSet checks if a flag was explicitly set on either the local or persistent flag set.
func flagWasSet(local, persistent *pflag.FlagSet, name string) bool {
	if f := local.Lookup(name); f != nil && f.Changed {
		return true
	}
	if f := persistent.Lookup(name); f != nil && f.Changed {
		return true
	}
	return false
}
