package cli

import (
	"fmt"
	"os"

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
	FlagPrayers    string
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
	pf.StringVar(&FlagPrayers, "prayers", "", "Comma-separated list of prayers to track (overrides config)")

	// Register subcommands.
	rootCmd.AddCommand(newNextCmd())
	rootCmd.AddCommand(newListCmd())
	rootCmd.AddCommand(newWeekCmd())
	rootCmd.AddCommand(newMonthCmd())
	rootCmd.AddCommand(newQueryCmd())
	rootCmd.AddCommand(newConfigCmd())
	rootCmd.AddCommand(newMethodsCmd())
	rootCmd.AddCommand(newCompletionCmd())

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

	// Prayers: CLI flag > config > leave empty (commands default to DefaultPrayerNames).
	if flagWasSet(flags, root, "prayers") {
		cfg.Prayers = FlagPrayers
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

// newCompletionCmd creates the completion subcommand for shell auto-completion.
func newCompletionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate shell completion script",
		Long: `Generate a shell completion script for the specified shell.

To load completions:

Bash:
  $ source <(prayer-times completion bash)
  # To load completions for each session, execute once:
  # Linux:
  $ prayer-times completion bash > /etc/bash_completion.d/prayer-times
  # macOS:
  $ prayer-times completion bash > $(brew --prefix)/etc/bash_completion.d/prayer-times

Zsh:
  $ source <(prayer-times completion zsh)
  # To load completions for each session, execute once:
  $ prayer-times completion zsh > "${fpath[1]}/_prayer-times"

Fish:
  $ prayer-times completion fish | source
  # To load completions for each session, execute once:
  $ prayer-times completion fish > ~/.config/fish/completions/prayer-times.fish

PowerShell:
  PS> prayer-times completion powershell | Out-String | Invoke-Expression
  # To load completions for each session, add the output to your profile.
`,
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		RunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "bash":
				return cmd.Root().GenBashCompletion(os.Stdout)
			case "zsh":
				return cmd.Root().GenZshCompletion(os.Stdout)
			case "fish":
				return cmd.Root().GenFishCompletion(os.Stdout, true)
			case "powershell":
				return cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
			default:
				return fmt.Errorf("unsupported shell: %s", args[0])
			}
		},
	}
	return cmd
}
