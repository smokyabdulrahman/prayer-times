package cli

import (
	"fmt"
	"strings"

	"github.com/smokyabdulrahman/prayer-times/internal/config"
	"github.com/spf13/cobra"
)

func newListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list [days]",
		Short: "Show prayer times for multiple days",
		Long:  "Display a grid of prayer times for N days (default: 7).",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement in Phase 4.
			fmt.Println("list command is not yet implemented")
			return nil
		},
	}
}

func newWeekCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "week",
		Short: "Show prayer times for the next 7 days",
		Long:  "Alias for 'list 7'. Display a grid of prayer times for 7 days.",
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement in Phase 4 (delegate to list 7).
			fmt.Println("week command is not yet implemented")
			return nil
		},
	}
}

func newMonthCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "month",
		Short: "Show prayer times for the next 30 days",
		Long:  "Alias for 'list 30'. Display a grid of prayer times for 30 days.",
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement in Phase 4 (delegate to list 30).
			fmt.Println("month command is not yet implemented")
			return nil
		},
	}
}

func newQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "query <prayer>",
		Short: "Query a specific prayer time",
		Long:  "Query a specific prayer time for today, or across multiple days with --days.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement in Phase 4.
			fmt.Printf("query command is not yet implemented (prayer: %s)\n", args[0])
			return nil
		},
	}

	cmd.Flags().String("days", "", "Number of days to show (or 'week'/'month')")

	return cmd
}

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Show or modify configuration",
		Long:  "Display current configuration, or use subcommands to modify it.\nWhen run without subcommands, shows the current configuration.",
		RunE:  runConfigShow,
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set a config value",
		Long: fmt.Sprintf("Set a configuration value. Valid keys: %s\n\nExamples:\n  prayer-times config set city Riyadh\n  prayer-times config set country \"Saudi Arabia\"\n  prayer-times config set method 4\n  prayer-times config set time_format 12h\n  prayer-times config set prayers Fajr,Dhuhr,Asr,Maghrib,Isha",
			strings.Join(config.ValidKeys, ", ")),
		Args: cobra.ExactArgs(2),
		RunE: runConfigSet,
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "reset",
		Short: "Reset config to defaults",
		Long:  "Delete the config file and restore all settings to defaults.",
		RunE:  runConfigReset,
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "path",
		Short: "Print config file path",
		RunE:  runConfigPath,
	})

	return cmd
}

// runConfigShow displays the current configuration.
func runConfigShow(cmd *cobra.Command, args []string) error {
	path, err := config.Path()
	if err != nil {
		return err
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	fmt.Printf("  Configuration (%s)\n\n", path)

	for _, key := range config.ValidKeys {
		val, _ := cfg.Get(key)
		display := val
		if display == "" {
			display = "(not set)"
		}
		// Add descriptive labels for method and school.
		if key == "method" && val != "" {
			display = formatMethodValue(val)
		}
		if key == "school" && val != "" {
			display = formatSchoolValue(val)
		}
		fmt.Printf("  %-14s %s\n", key, display)
	}
	return nil
}

// runConfigSet sets a config key to the given value.
func runConfigSet(cmd *cobra.Command, args []string) error {
	key, value := args[0], args[1]

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	if err := cfg.Set(key, value); err != nil {
		return err
	}

	if err := cfg.Save(); err != nil {
		return err
	}

	fmt.Printf("Set %s = %s\n", key, value)
	return nil
}

// runConfigReset deletes the config file.
func runConfigReset(cmd *cobra.Command, args []string) error {
	if err := config.Reset(); err != nil {
		return err
	}
	fmt.Println("Configuration reset to defaults.")
	return nil
}

// runConfigPath prints the config file path.
func runConfigPath(cmd *cobra.Command, args []string) error {
	path, err := config.Path()
	if err != nil {
		return err
	}
	fmt.Println(path)
	return nil
}

// formatMethodValue adds the method name to the numeric value.
func formatMethodValue(val string) string {
	for _, m := range CalculationMethods {
		if fmt.Sprintf("%d", m.ID) == val {
			return fmt.Sprintf("%s (%s)", val, m.Name)
		}
	}
	return val
}

// formatSchoolValue adds the school name to the numeric value.
func formatSchoolValue(val string) string {
	switch val {
	case "0":
		return "0 (Shafi)"
	case "1":
		return "1 (Hanafi)"
	default:
		return val
	}
}

// CalculationMethods lists all supported Al Adhan API calculation methods.
var CalculationMethods = []struct {
	ID   int
	Name string
}{
	{0, "Shia Ithna-Ashari (Jafari)"},
	{1, "University of Islamic Sciences, Karachi"},
	{2, "Islamic Society of North America (ISNA)"},
	{3, "Muslim World League (MWL)"},
	{4, "Umm Al-Qura University, Makkah"},
	{5, "Egyptian General Authority of Survey"},
	{7, "Institute of Geophysics, University of Tehran"},
	{8, "Gulf Region"},
	{9, "Kuwait"},
	{10, "Qatar"},
	{11, "Majlis Ugama Islam Singapura (Singapore)"},
	{12, "Union Organization Islamic de France"},
	{13, "Diyanet Isleri Baskanligi, Turkey (experimental)"},
	{14, "Spiritual Administration of Muslims of Russia"},
	{15, "Moonsighting Committee Worldwide"},
	{16, "Dubai (experimental)"},
	{17, "JAKIM (Malaysia)"},
	{18, "Tunisia"},
	{19, "Algeria"},
	{20, "KEMENAG (Indonesia)"},
	{21, "Morocco"},
	{22, "Comunidade Islamica de Lisboa (Portugal)"},
	{23, "Ministry of Awqaf, Jordan"},
}

func newMethodsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "methods",
		Short: "List all calculation methods",
		Long:  "Print the table of all supported Al Adhan API calculation methods.",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Supported calculation methods:")
			fmt.Println()
			fmt.Printf("  %-4s %s\n", "ID", "Name")
			fmt.Printf("  %-4s %s\n", "──", "────")
			for _, m := range CalculationMethods {
				fmt.Printf("  %-4d %s\n", m.ID, m.Name)
			}
			fmt.Println()
			fmt.Println("Use --method <ID> to select a calculation method.")
			fmt.Println("If omitted, the API picks a default based on your location.")
			return nil
		},
	}
}
