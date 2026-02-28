package display

import (
	"fmt"
	"strings"
)

// Table renders an aligned text table with optional color support.
type Table struct {
	headers []string
	rows    [][]string
	// highlightRow is the 0-based row index to highlight (typically "today"). -1 = none.
	highlightRow int
}

// NewTable creates a new table with the given column headers.
func NewTable(headers []string) *Table {
	return &Table{
		headers:      headers,
		highlightRow: -1,
	}
}

// AddRow appends a row of values. The number of values should match the number of headers.
func (t *Table) AddRow(values []string) {
	t.rows = append(t.rows, values)
}

// SetHighlightRow sets which row index (0-based) should be highlighted.
func (t *Table) SetHighlightRow(idx int) {
	t.highlightRow = idx
}

// Render produces the formatted table string with leading indent.
func (t *Table) Render() string {
	if len(t.headers) == 0 {
		return ""
	}

	// Calculate column widths.
	widths := make([]int, len(t.headers))
	for i, h := range t.headers {
		widths[i] = len(h)
	}
	for _, row := range t.rows {
		for i, cell := range row {
			if i < len(widths) && len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}

	var sb strings.Builder

	// Header row.
	headerLine := formatRow(t.headers, widths)
	sb.WriteString("  " + Bold(headerLine) + "\n")

	// Separator row using Unicode box-drawing dashes.
	sepParts := make([]string, len(widths))
	for i, w := range widths {
		sepParts[i] = strings.Repeat("─", w)
	}
	sepLine := "  " + strings.Join(sepParts, "  ")
	sb.WriteString(Dim(sepLine) + "\n")

	// Data rows.
	for i, row := range t.rows {
		line := formatRow(row, widths)
		if i == t.highlightRow {
			sb.WriteString("  " + Accent(line) + "\n")
		} else {
			sb.WriteString("  " + line + "\n")
		}
	}

	return sb.String()
}

// formatRow formats a row of cells using the given column widths.
func formatRow(cells []string, widths []int) string {
	parts := make([]string, len(widths))
	for i, w := range widths {
		cell := ""
		if i < len(cells) {
			cell = cells[i]
		}
		parts[i] = fmt.Sprintf("%-*s", w, cell)
	}
	return strings.Join(parts, "  ")
}
