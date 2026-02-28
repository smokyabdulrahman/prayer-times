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
		{"config"},
		{"config", "path"},
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
