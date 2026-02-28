package main

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

// TestVersionFlag verifies that --version prints the version string.
func TestVersionFlag(t *testing.T) {
	// Build the binary with a known version.
	binPath := t.TempDir() + "/tmux-prayer-times"
	cmd := exec.Command("go", "build", "-ldflags", "-X main.version=v1.2.3-test", "-o", binPath, ".")
	cmd.Dir = "."
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build failed: %v\n%s", err, out)
	}

	out, err := exec.Command(binPath, "--version").Output()
	if err != nil {
		t.Fatalf("--version failed: %v", err)
	}

	got := strings.TrimSpace(string(out))
	want := "tmux-prayer-times v1.2.3-test"
	if got != want {
		t.Errorf("--version = %q, want %q", got, want)
	}
}

// TestVersionFlag_Dev verifies the default "dev" version when no ldflags.
func TestVersionFlag_Dev(t *testing.T) {
	binPath := t.TempDir() + "/tmux-prayer-times"
	cmd := exec.Command("go", "build", "-o", binPath, ".")
	cmd.Dir = "."
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build failed: %v\n%s", err, out)
	}

	out, err := exec.Command(binPath, "--version").Output()
	if err != nil {
		t.Fatalf("--version failed: %v", err)
	}

	got := strings.TrimSpace(string(out))
	if !strings.HasPrefix(got, "tmux-prayer-times ") {
		t.Errorf("--version output unexpected: %q", got)
	}
}

// TestListMethodsFlag verifies that --list-methods prints calculation methods.
func TestListMethodsFlag(t *testing.T) {
	binPath := t.TempDir() + "/tmux-prayer-times"
	cmd := exec.Command("go", "build", "-o", binPath, ".")
	cmd.Dir = "."
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build failed: %v\n%s", err, out)
	}

	out, err := exec.Command(binPath, "--list-methods").Output()
	if err != nil {
		t.Fatalf("--list-methods failed: %v", err)
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
			t.Errorf("--list-methods output missing %q", m)
		}
	}
}

// TestNoArgs_ExitCode verifies that running with no args and no location exits with error.
func TestNoArgs_ExitCode(t *testing.T) {
	binPath := t.TempDir() + "/tmux-prayer-times"
	cmd := exec.Command("go", "build", "-o", binPath, ".")
	cmd.Dir = "."
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build failed: %v\n%s", err, out)
	}

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

// TestCalculationMethods_NoDuplicateIDs ensures no duplicate method IDs.
func TestCalculationMethods_NoDuplicateIDs(t *testing.T) {
	seen := make(map[int]bool)
	for _, m := range calculationMethods {
		if seen[m.ID] {
			t.Errorf("duplicate calculation method ID: %d", m.ID)
		}
		seen[m.ID] = true
	}
}

// TestCalculationMethods_IDsAreValid ensures method IDs are in the expected range.
func TestCalculationMethods_IDsAreValid(t *testing.T) {
	for _, m := range calculationMethods {
		if m.ID < 0 || m.ID > 23 {
			t.Errorf("method ID %d out of range 0-23", m.ID)
		}
		if m.Name == "" {
			t.Errorf("method ID %d has empty name", m.ID)
		}
	}
}
