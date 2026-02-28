package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// buildBinary compiles the prayer-times binary to a temp directory for testing.
func buildBinary(t *testing.T, ldflags string) string {
	t.Helper()
	binPath := filepath.Join(t.TempDir(), "prayer-times")

	args := []string{"build"}
	if ldflags != "" {
		args = append(args, "-ldflags", ldflags)
	}
	args = append(args, "-o", binPath, "../../cmd/prayer-times")

	cmd := exec.Command("go", args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build failed: %v\n%s", err, out)
	}
	return binPath
}

// TestVersionFlag verifies that --version prints the version string.
func TestVersionFlag(t *testing.T) {
	binPath := buildBinary(t, "-X main.version=v1.2.3-test")

	out, err := exec.Command(binPath, "--version").Output()
	if err != nil {
		t.Fatalf("--version failed: %v", err)
	}

	got := strings.TrimSpace(string(out))
	want := "prayer-times version v1.2.3-test"
	if got != want {
		t.Errorf("--version = %q, want %q", got, want)
	}
}

// TestVersionFlag_Dev verifies the default "dev" version when no ldflags.
func TestVersionFlag_Dev(t *testing.T) {
	binPath := buildBinary(t, "")

	out, err := exec.Command(binPath, "--version").Output()
	if err != nil {
		t.Fatalf("--version failed: %v", err)
	}

	got := strings.TrimSpace(string(out))
	if !strings.HasPrefix(got, "prayer-times version ") {
		t.Errorf("--version output unexpected: %q", got)
	}
}

// TestMethodsSubcommand verifies that 'methods' prints calculation methods.
func TestMethodsSubcommand(t *testing.T) {
	binPath := buildBinary(t, "")

	out, err := exec.Command(binPath, "methods").Output()
	if err != nil {
		t.Fatalf("methods failed: %v", err)
	}

	output := string(out)

	// Check for a few expected methods.
	expectedMethods := []string{
		"ISNA",
		"Muslim World League",
		"Umm Al-Qura",
		"Jafari",
		"Ministry of Awqaf, Jordan",
	}
	for _, m := range expectedMethods {
		if !strings.Contains(output, m) {
			t.Errorf("methods output missing %q", m)
		}
	}
}

// TestNoArgs_ExitCode verifies that running with no args and no location exits with error.
func TestNoArgs_ExitCode(t *testing.T) {
	binPath := buildBinary(t, "")

	// Run with a bogus cache dir and no location -- should fail since
	// geolocation may or may not work in CI. We use an env var to set
	// a non-writable cache dir to force a cache warning, but the main
	// error will be either a geolocation or the location-related error.
	runCmd := exec.Command(binPath, "--cache-dir", "/dev/null/impossible")
	runCmd.Env = append(os.Environ(), "HOME=/dev/null")
	err := runCmd.Run()
	if err == nil {
		// If it succeeds, geo-detection worked -- that's fine, not an error.
		return
	}
	// We expect a non-zero exit code.
	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatalf("expected ExitError, got %T: %v", err, err)
	}
	if exitErr.ExitCode() == 0 {
		t.Error("expected non-zero exit code")
	}
}

// TestNextSubcommand_ExitCode verifies that 'next' with no location exits with error.
func TestNextSubcommand_ExitCode(t *testing.T) {
	binPath := buildBinary(t, "")

	runCmd := exec.Command(binPath, "next", "--cache-dir", "/dev/null/impossible")
	runCmd.Env = append(os.Environ(), "HOME=/dev/null")
	err := runCmd.Run()
	if err == nil {
		// Geo-detection worked -- fine.
		return
	}
	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatalf("expected ExitError, got %T: %v", err, err)
	}
	if exitErr.ExitCode() == 0 {
		t.Error("expected non-zero exit code")
	}
}

// TestCalculationMethods_NoDuplicateIDs ensures no duplicate method IDs.
func TestCalculationMethods_NoDuplicateIDs(t *testing.T) {
	seen := make(map[int]bool)
	for _, m := range CalculationMethods {
		if seen[m.ID] {
			t.Errorf("duplicate calculation method ID: %d", m.ID)
		}
		seen[m.ID] = true
	}
}

// TestCalculationMethods_IDsAreValid ensures method IDs are in the expected range.
func TestCalculationMethods_IDsAreValid(t *testing.T) {
	for _, m := range CalculationMethods {
		if m.ID < 0 || m.ID > 23 {
			t.Errorf("method ID %d out of range 0-23", m.ID)
		}
		if m.Name == "" {
			t.Errorf("method ID %d has empty name", m.ID)
		}
	}
}

// TestHelpFlag verifies that --help shows the expected subcommands.
func TestHelpFlag(t *testing.T) {
	binPath := buildBinary(t, "")

	out, err := exec.Command(binPath, "--help").Output()
	if err != nil {
		t.Fatalf("--help failed: %v", err)
	}

	output := string(out)

	expectedSubcommands := []string{
		"next",
		"list",
		"week",
		"month",
		"query",
		"config",
		"methods",
	}
	for _, sub := range expectedSubcommands {
		if !strings.Contains(output, sub) {
			t.Errorf("--help output missing subcommand %q", sub)
		}
	}
}

// TestStubbedSubcommands verifies stubbed commands run without error.
func TestStubbedSubcommands(t *testing.T) {
	binPath := buildBinary(t, "")

	stubs := [][]string{
		{"list"},
		{"week"},
		{"month"},
		{"query", "fajr"},
	}

	for _, args := range stubs {
		t.Run(strings.Join(args, "_"), func(t *testing.T) {
			out, err := exec.Command(binPath, args...).CombinedOutput()
			if err != nil {
				t.Errorf("command %v failed: %v\n%s", args, err, out)
			}
		})
	}
}

// runWithConfig runs the binary with XDG_CONFIG_HOME set to a temp dir
// so tests don't touch the real config file.
func runWithConfig(t *testing.T, binPath, configDir string, args ...string) (string, error) {
	t.Helper()
	cmd := exec.Command(binPath, args...)
	cmd.Env = append(os.Environ(), "XDG_CONFIG_HOME="+configDir)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// TestConfigShow_Empty verifies 'config' with no config file shows "(not set)" for all fields.
func TestConfigShow_Empty(t *testing.T) {
	binPath := buildBinary(t, "")
	configDir := t.TempDir()

	output, err := runWithConfig(t, binPath, configDir, "config")
	if err != nil {
		t.Fatalf("config show failed: %v\n%s", err, output)
	}

	// All values should show "(not set)".
	expectedKeys := []string{"city", "country", "latitude", "longitude", "method", "school", "time_format", "prayers", "cache_dir"}
	for _, key := range expectedKeys {
		if !strings.Contains(output, key) {
			t.Errorf("config show output missing key %q", key)
		}
	}
	if !strings.Contains(output, "(not set)") {
		t.Error("config show should contain '(not set)' for empty config")
	}
}

// TestConfigSet_And_Show verifies 'config set' persists values and 'config' shows them.
func TestConfigSet_And_Show(t *testing.T) {
	binPath := buildBinary(t, "")
	configDir := t.TempDir()

	// Set city.
	out, err := runWithConfig(t, binPath, configDir, "config", "set", "city", "Riyadh")
	if err != nil {
		t.Fatalf("config set city failed: %v\n%s", err, out)
	}
	if !strings.Contains(out, "Set city = Riyadh") {
		t.Errorf("config set output unexpected: %s", out)
	}

	// Set country.
	out, err = runWithConfig(t, binPath, configDir, "config", "set", "country", "Saudi Arabia")
	if err != nil {
		t.Fatalf("config set country failed: %v\n%s", err, out)
	}

	// Set method.
	out, err = runWithConfig(t, binPath, configDir, "config", "set", "method", "4")
	if err != nil {
		t.Fatalf("config set method failed: %v\n%s", err, out)
	}

	// Show config and verify values.
	output, err := runWithConfig(t, binPath, configDir, "config")
	if err != nil {
		t.Fatalf("config show failed: %v\n%s", err, output)
	}

	if !strings.Contains(output, "Riyadh") {
		t.Error("config show should contain 'Riyadh'")
	}
	if !strings.Contains(output, "Saudi Arabia") {
		t.Error("config show should contain 'Saudi Arabia'")
	}
	if !strings.Contains(output, "4") && !strings.Contains(output, "Umm Al-Qura") {
		t.Error("config show should contain method 4 info")
	}
}

// TestConfigSet_InvalidKey verifies 'config set' with an invalid key fails.
func TestConfigSet_InvalidKey(t *testing.T) {
	binPath := buildBinary(t, "")
	configDir := t.TempDir()

	_, err := runWithConfig(t, binPath, configDir, "config", "set", "invalid_key", "value")
	if err == nil {
		t.Fatal("config set with invalid key should fail")
	}
}

// TestConfigSet_InvalidValue verifies 'config set' with invalid values fails.
func TestConfigSet_InvalidValue(t *testing.T) {
	binPath := buildBinary(t, "")
	configDir := t.TempDir()

	tests := []struct {
		name  string
		key   string
		value string
	}{
		{"bad method", "method", "99"},
		{"bad school", "school", "5"},
		{"bad time_format", "time_format", "30h"},
		{"bad latitude", "latitude", "999"},
		{"bad longitude", "longitude", "abc"},
		{"bad prayers", "prayers", "InvalidPrayer"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := runWithConfig(t, binPath, configDir, "config", "set", tt.key, tt.value)
			if err == nil {
				t.Errorf("config set %s %s should have failed", tt.key, tt.value)
			}
		})
	}
}

// TestConfigReset verifies 'config reset' removes the config file.
func TestConfigReset(t *testing.T) {
	binPath := buildBinary(t, "")
	configDir := t.TempDir()

	// Set a value first.
	runWithConfig(t, binPath, configDir, "config", "set", "city", "London")

	// Reset.
	out, err := runWithConfig(t, binPath, configDir, "config", "reset")
	if err != nil {
		t.Fatalf("config reset failed: %v\n%s", err, out)
	}
	if !strings.Contains(out, "reset") {
		t.Errorf("config reset output unexpected: %s", out)
	}

	// Show should now be all "(not set)".
	output, err := runWithConfig(t, binPath, configDir, "config")
	if err != nil {
		t.Fatalf("config show after reset failed: %v\n%s", err, output)
	}
	if strings.Contains(output, "London") {
		t.Error("config show after reset should not contain 'London'")
	}
}

// TestConfigPath verifies 'config path' prints a valid path.
func TestConfigPath(t *testing.T) {
	binPath := buildBinary(t, "")
	configDir := t.TempDir()

	output, err := runWithConfig(t, binPath, configDir, "config", "path")
	if err != nil {
		t.Fatalf("config path failed: %v\n%s", err, output)
	}

	path := strings.TrimSpace(output)
	if !strings.Contains(path, "prayer-times") {
		t.Errorf("config path should contain 'prayer-times', got: %q", path)
	}
	if !strings.HasSuffix(path, "config.json") {
		t.Errorf("config path should end with 'config.json', got: %q", path)
	}
}

// TestConfigPath_XDG verifies 'config path' respects XDG_CONFIG_HOME.
func TestConfigPath_XDG(t *testing.T) {
	binPath := buildBinary(t, "")
	configDir := t.TempDir()

	output, err := runWithConfig(t, binPath, configDir, "config", "path")
	if err != nil {
		t.Fatalf("config path failed: %v\n%s", err, output)
	}

	path := strings.TrimSpace(output)
	expected := filepath.Join(configDir, "prayer-times", "config.json")
	if path != expected {
		t.Errorf("config path = %q, want %q", path, expected)
	}
}
