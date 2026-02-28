package display

import (
	"strings"
	"testing"
)

func TestNewTable(t *testing.T) {
	tbl := NewTable([]string{"Name", "Value"})
	if tbl == nil {
		t.Fatal("NewTable returned nil")
	}
	if tbl.highlightRow != -1 {
		t.Errorf("highlightRow = %d, want -1", tbl.highlightRow)
	}
}

func TestTable_EmptyHeaders(t *testing.T) {
	tbl := NewTable([]string{})
	got := tbl.Render()
	if got != "" {
		t.Errorf("Render() with empty headers = %q, want empty", got)
	}
}

func TestTable_BasicRender(t *testing.T) {
	SetEnabled(false) // disable colors for predictable output

	tbl := NewTable([]string{"Date", "Fajr", "Isha"})
	tbl.AddRow([]string{"Mon 01 Mar", "05:06", "19:28"})
	tbl.AddRow([]string{"Tue 02 Mar", "05:05", "19:29"})

	got := tbl.Render()

	// Check header is present.
	if !strings.Contains(got, "Date") || !strings.Contains(got, "Fajr") || !strings.Contains(got, "Isha") {
		t.Errorf("Render() missing headers in:\n%s", got)
	}

	// Check separator exists (Unicode dashes).
	if !strings.Contains(got, "─") {
		t.Error("Render() missing separator line")
	}

	// Check data rows.
	if !strings.Contains(got, "Mon 01 Mar") {
		t.Error("Render() missing first data row")
	}
	if !strings.Contains(got, "Tue 02 Mar") {
		t.Error("Render() missing second data row")
	}
	if !strings.Contains(got, "05:06") || !strings.Contains(got, "19:28") {
		t.Error("Render() missing prayer time values")
	}
}

func TestTable_ColumnAlignment(t *testing.T) {
	SetEnabled(false)

	tbl := NewTable([]string{"A", "LongHeader"})
	tbl.AddRow([]string{"short", "x"})
	tbl.AddRow([]string{"y", "longer value"})

	got := tbl.Render()
	lines := strings.Split(strings.TrimSpace(got), "\n")

	// Should have 4 lines: header, separator, 2 data rows.
	if len(lines) != 4 {
		t.Errorf("expected 4 lines, got %d:\n%s", len(lines), got)
	}
}

func TestTable_HighlightRow(t *testing.T) {
	SetEnabled(true)
	defer SetEnabled(false)

	tbl := NewTable([]string{"Date", "Time"})
	tbl.AddRow([]string{"Mon", "05:00"})
	tbl.AddRow([]string{"Tue", "05:01"})
	tbl.SetHighlightRow(0)

	got := tbl.Render()

	// The highlighted row should contain ANSI codes.
	lines := strings.Split(got, "\n")
	// Line 0 is header, line 1 is separator, line 2 is first data row (highlighted).
	if len(lines) < 3 {
		t.Fatalf("expected at least 3 lines, got %d", len(lines))
	}
	if !strings.Contains(lines[2], "\033[") {
		t.Error("highlighted row should contain ANSI escape codes")
	}
}

func TestFormatRow(t *testing.T) {
	got := formatRow([]string{"abc", "de"}, []int{5, 4})
	want := "abc    de  "
	if got != want {
		t.Errorf("formatRow = %q, want %q", got, want)
	}
}

func TestFormatRow_MissingCells(t *testing.T) {
	// Fewer cells than widths should produce empty-padded columns.
	got := formatRow([]string{"a"}, []int{3, 5})
	// "a  " (3) + "  " (sep) + "     " (5) = "a         "
	want := "a         "
	if got != want {
		t.Errorf("formatRow = %q, want %q", got, want)
	}
}
