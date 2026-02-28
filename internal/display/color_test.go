package display

import (
	"testing"
)

func TestWrap_Enabled(t *testing.T) {
	// Force colors on for testing.
	SetEnabled(true)
	defer SetEnabled(false)

	got := Bold("hello")
	if got != "\033[1mhello\033[0m" {
		t.Errorf("Bold(\"hello\") = %q, want ANSI bold wrapped", got)
	}
}

func TestWrap_Disabled(t *testing.T) {
	SetEnabled(false)

	got := Bold("hello")
	if got != "hello" {
		t.Errorf("Bold(\"hello\") with colors disabled = %q, want plain \"hello\"", got)
	}
}

func TestDim(t *testing.T) {
	SetEnabled(true)
	defer SetEnabled(false)

	got := Dim("text")
	if got != "\033[2mtext\033[0m" {
		t.Errorf("Dim(\"text\") = %q, want ANSI dim wrapped", got)
	}
}

func TestGreen(t *testing.T) {
	SetEnabled(true)
	defer SetEnabled(false)

	got := Green("ok")
	if got != "\033[32mok\033[0m" {
		t.Errorf("Green(\"ok\") = %q, want ANSI green wrapped", got)
	}
}

func TestYellow(t *testing.T) {
	SetEnabled(true)
	defer SetEnabled(false)

	got := Yellow("warn")
	if got != "\033[33mwarn\033[0m" {
		t.Errorf("Yellow(\"warn\") = %q, want ANSI yellow wrapped", got)
	}
}

func TestCyan(t *testing.T) {
	SetEnabled(true)
	defer SetEnabled(false)

	got := Cyan("info")
	if got != "\033[36minfo\033[0m" {
		t.Errorf("Cyan(\"info\") = %q, want ANSI cyan wrapped", got)
	}
}

func TestGray(t *testing.T) {
	SetEnabled(true)
	defer SetEnabled(false)

	got := Gray("muted")
	if got != "\033[90mmuted\033[0m" {
		t.Errorf("Gray(\"muted\") = %q, want ANSI gray wrapped", got)
	}
}

func TestAccent_Enabled(t *testing.T) {
	SetEnabled(true)
	defer SetEnabled(false)

	got := Accent("next")
	want := "\033[1m\033[36mnext\033[0m"
	if got != want {
		t.Errorf("Accent(\"next\") = %q, want %q", got, want)
	}
}

func TestAccent_Disabled(t *testing.T) {
	SetEnabled(false)

	got := Accent("next")
	if got != "next" {
		t.Errorf("Accent(\"next\") with colors disabled = %q, want plain \"next\"", got)
	}
}

func TestBoldf(t *testing.T) {
	SetEnabled(true)
	defer SetEnabled(false)

	got := Boldf("count: %d", 42)
	want := "\033[1mcount: 42\033[0m"
	if got != want {
		t.Errorf("Boldf = %q, want %q", got, want)
	}
}

func TestEnabled_ReportsState(t *testing.T) {
	SetEnabled(true)
	if !Enabled() {
		t.Error("Enabled() should return true after SetEnabled(true)")
	}

	SetEnabled(false)
	if Enabled() {
		t.Error("Enabled() should return false after SetEnabled(false)")
	}
}

func TestAllColors_Disabled_ReturnPlainText(t *testing.T) {
	SetEnabled(false)

	funcs := []struct {
		name string
		fn   func(string) string
	}{
		{"Bold", Bold},
		{"Dim", Dim},
		{"Green", Green},
		{"Yellow", Yellow},
		{"Cyan", Cyan},
		{"Gray", Gray},
		{"Accent", Accent},
	}

	for _, f := range funcs {
		t.Run(f.name, func(t *testing.T) {
			got := f.fn("plain")
			if got != "plain" {
				t.Errorf("%s(\"plain\") with colors disabled = %q, want \"plain\"", f.name, got)
			}
		})
	}
}
