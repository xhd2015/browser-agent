package server

import (
	"fmt"
	"io"
	"strings"
)

func formatSize(bytes int) string {
	if bytes < 0 {
		return "—"
	}
	if bytes < 1024 {
		return fmt.Sprintf("%d B", bytes)
	}
	if bytes < 1024*1024 {
		return fmt.Sprintf("%.1f KB", float64(bytes)/1024)
	}
	return fmt.Sprintf("%.1f MB", float64(bytes)/(1024*1024))
}

func formatTime(ms float64) string {
	if ms < 0 {
		return "—"
	}
	if ms < 1 {
		return fmt.Sprintf("%.0f µs", ms*1000)
	}
	if ms < 1000 {
		return fmt.Sprintf("%.1f ms", ms)
	}
	return fmt.Sprintf("%.2f s", ms/1000)
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func PrintEntriesTable(w io.Writer, entries []EntrySummary) {
	if len(entries) == 0 {
		fmt.Fprintln(w, "No entries found.")
		return
	}

	// Column widths
	const (
		colIdx    = 5
		colStatus = 6
		colMethod = 7
		colType   = 10
		colSize   = 10
		colTime   = 10
		colHost   = 30
		colPath   = 60
	)

	header := fmt.Sprintf("%-*s %-*s %-*s %-*s %-*s %-*s %*s %*s",
		colIdx, "#",
		colStatus, "Status",
		colMethod, "Method",
		colHost, "Host",
		colPath, "Path",
		colType, "Type",
		colSize, "Size",
		colTime, "Time",
	)
	fmt.Fprintln(w, header)
	fmt.Fprintln(w, strings.Repeat("─", len(header)))

	for _, e := range entries {
		fmt.Fprintf(w, "%-*d %-*d %-*s %-*s %-*s %-*s %*s %*s\n",
			colIdx, e.Index,
			colStatus, e.Status,
			colMethod, e.Method,
			colHost, truncate(e.Host, colHost),
			colPath, truncate(e.Path, colPath),
			colType, truncate(e.ResourceType, colType),
			colSize, formatSize(e.Size),
			colTime, formatTime(e.Time),
		)
	}

	fmt.Fprintf(w, "\nTotal: %d requests\n", len(entries))
}
