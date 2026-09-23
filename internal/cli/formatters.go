package cli

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strings"
)

// FormatJSON renders an arbitrary struct or slice as pretty-printed JSON.
func FormatJSON(v any) (string, error) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// FormatCSV renders tabular data as standard comma-separated values.
func FormatCSV(headers []string, rows [][]string) (string, error) {
	buf := new(bytes.Buffer)
	writer := csv.NewWriter(buf)

	if len(headers) > 0 {
		if err := writer.Write(headers); err != nil {
			return "", err
		}
	}

	for _, r := range rows {
		if err := writer.Write(r); err != nil {
			return "", err
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// FormatTable renders an aligned terminal ASCII table with clean borders.
func FormatTable(headers []string, rows [][]string) string {
	if len(headers) == 0 && len(rows) == 0 {
		return ""
	}

	numCols := len(headers)
	for _, r := range rows {
		if len(r) > numCols {
			numCols = len(r)
		}
	}

	colWidths := make([]int, numCols)
	for i, h := range headers {
		if len(h) > colWidths[i] {
			colWidths[i] = len(h)
		}
	}

	for _, r := range rows {
		for i, cell := range r {
			if len(cell) > colWidths[i] {
				colWidths[i] = len(cell)
			}
		}
	}

	var sb strings.Builder

	// Header row
	if len(headers) > 0 {
		for i, h := range headers {
			sb.WriteString(fmt.Sprintf("%-*s  ", colWidths[i], h))
		}
		sb.WriteString("\n")

		// Divider
		for i := 0; i < numCols; i++ {
			sb.WriteString(strings.Repeat("-", colWidths[i]))
			sb.WriteString("  ")
		}
		sb.WriteString("\n")
	}

	// Data rows
	for _, r := range rows {
		for i := 0; i < numCols; i++ {
			cell := ""
			if i < len(r) {
				cell = r[i]
			}
			sb.WriteString(fmt.Sprintf("%-*s  ", colWidths[i], cell))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}
