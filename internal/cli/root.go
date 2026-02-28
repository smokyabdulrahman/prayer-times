package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Global flags shared across all subcommands.
var (
	FlagCity      string
	FlagCountry   string
	FlagLatitude  float64
	FlagLongitude float64
	FlagMethod    int
	FlagSchool    int
	FlagJSON      bool
	FlagCacheDir  string
)

// NewRootCmd creates the root command for the prayer-times CLI.
// The version parameter is set by the calling binary via ldflags.
func NewRootCmd(version string) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:     "prayer-times",
		Short:   "Islamic prayer times CLI",
		Long:    "A full-featured CLI for Islamic prayer times powered by the Al Adhan API.",
		Version: version,
		// Default action: show today's prayer times (stubbed for now, implemented in Phase 3).
		RunE: func(cmd *cobra.Command, args []string) error {
			// Phase 3 will implement the rich "today" display.
			// For now, delegate to the `next` subcommand behavior.
			return runNext(cmd, args)
		},
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
