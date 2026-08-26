package browseragent

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
)

const harRootHelp = `Usage: browser-agent har <command> [flags]

Offline HAR helpers for Browser Agent exports (no live session required).

Commands:
  inspect <subcommand> <path>   Inspect an export directory or .har file

Capture (live) remains under:
  browser-agent session har start|end

Run 'browser-agent har inspect --help' for inspect subcommands.
`

const harInspectHelp = `Usage: browser-agent har inspect <command> <path> [flags]

Inspect exported Browser Agent HAR directories or .har files (offline).

Commands:
  summary   Manifest + per-tab/file counts + body coverage
  paths     Unique METHOD + path inventory
  entries   Filtered request list
  show      One entry with request/response bodies

<path>:
  export directory containing manifest.json, or a single .har file

Shared flags:
  --json              Machine-readable JSON on stdout (no ANSI)
  --host <host>       Include only this host (repeatable; substring/suffix match)
  --method <m>        Filter by HTTP method (e.g. POST)
  --path-substr <s>   Substring match on path (+ query)
  --match <s>         Substring match on URL or path (preferred for show)
  --tab-id <id>       Limit to one tab (directory mode)
  --no-noise          Drop static/analytics (default: true)
  --noise             Include noise entries
  --redact            Redact secrets (default: true)
  --no-redact         Disable redaction
  --limit N           Max rows for entries (default 100); optional for paths
  --index N           show: 0-based entry index (directory mode needs --tab-id)
  -h, --help
`

const harInspectSummaryHelp = `Usage: browser-agent har inspect summary <path> [flags]

Print capture summary: manifest fields (if directory), per-tab counts, body coverage.

Flags: --json, --host, --method, --path-substr, --match, --tab-id, --no-noise/--noise, -h
`

const harInspectPathsHelp = `Usage: browser-agent har inspect paths <path> [flags]

Print unique METHOD + path inventory sorted by count.

Flags: --json, --host, --method, --path-substr, --match, --tab-id, --no-noise/--noise, --limit, -h
`

const harInspectEntriesHelp = `Usage: browser-agent har inspect entries <path> [flags]

Print filtered HAR entries (index, method, status, path, body lengths).

Flags: --json, --host, --method, --path-substr, --match, --tab-id, --no-noise/--noise, --limit, -h
`

const harInspectShowHelp = `Usage: browser-agent har inspect show <path> [flags]

Show one HAR entry with redacted request/response bodies.

Select with --match/--path-substr (must be unique) or --index (with --tab-id for directories).

Flags: --json, --match, --path-substr, --method, --host, --tab-id, --index,
       --no-noise/--noise, --redact/--no-redact, -h
`

func cliHAR(args []string, env map[string]string, stdout, stderr io.Writer) error {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" || args[0] == "help" {
		_, _ = io.WriteString(stdout, harRootHelp)
		return nil
	}
	switch args[0] {
	case "inspect":
		return cliHARInspect(args[1:], env, stdout, stderr)
	default:
		return fmt.Errorf("unknown har subcommand %q; try inspect", args[0])
	}
}

func cliHARInspect(args []string, env map[string]string, stdout, stderr io.Writer) error {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" || args[0] == "help" {
		_, _ = io.WriteString(stdout, harInspectHelp)
		return nil
	}
	cmd := args[0]
	rest := args[1:]
	switch cmd {
	case "summary":
		if len(rest) > 0 && (rest[0] == "-h" || rest[0] == "--help" || rest[0] == "help") {
			_, _ = io.WriteString(stdout, harInspectSummaryHelp)
			return nil
		}
		return cliHARInspectSummary(rest, env, stdout, stderr)
	case "paths":
		if len(rest) > 0 && (rest[0] == "-h" || rest[0] == "--help" || rest[0] == "help") {
			_, _ = io.WriteString(stdout, harInspectPathsHelp)
			return nil
		}
		return cliHARInspectPaths(rest, env, stdout, stderr)
	case "entries":
		if len(rest) > 0 && (rest[0] == "-h" || rest[0] == "--help" || rest[0] == "help") {
			_, _ = io.WriteString(stdout, harInspectEntriesHelp)
			return nil
		}
		return cliHARInspectEntries(rest, env, stdout, stderr)
	case "show":
		if len(rest) > 0 && (rest[0] == "-h" || rest[0] == "--help" || rest[0] == "help") {
			_, _ = io.WriteString(stdout, harInspectShowHelp)
			return nil
		}
		return cliHARInspectShow(rest, env, stdout, stderr)
	default:
		return fmt.Errorf("unknown har inspect subcommand %q; try summary, paths, entries, or show", cmd)
	}
}

type harInspectCLIOptions struct {
	asJSON    bool
	filter    harInspectFilter
	redact    bool
	limit     int
	limitSet  bool
	index     int
	indexSet  bool
	forceColor bool
	noColor    bool
}

func parseHARInspectOptions(args []string) (path string, opts harInspectCLIOptions, err error) {
	opts.redact = true
	opts.filter.DropNoise = true
	opts.limit = 100

	path, rest, err := takeInspectPath(args)
	if err != nil {
		return "", opts, err
	}
	opts.asJSON = hasFlag(rest, "--json")
	opts.forceColor = hasFlag(rest, "--color")
	opts.noColor = hasFlag(rest, "--no-color")
	if hasFlag(rest, "--noise") {
		opts.filter.DropNoise = false
	}
	if hasFlag(rest, "--no-noise") {
		opts.filter.DropNoise = true
	}
	if hasFlag(rest, "--no-redact") {
		opts.redact = false
	}
	if hasFlag(rest, "--redact") {
		opts.redact = true
	}
	opts.filter.Hosts = collectFlagValues(rest, "--host")
	opts.filter.Method = strings.TrimSpace(flagString(rest, "--method"))
	opts.filter.PathSubstr = flagString(rest, "--path-substr")
	opts.filter.Match = flagString(rest, "--match")
	if tabRaw, ok := flagStringSet(rest, "--tab-id"); ok {
		n, err := strconv.Atoi(strings.TrimSpace(tabRaw))
		if err != nil {
			return "", opts, fmt.Errorf("--tab-id must be an integer")
		}
		opts.filter.TabID = n
		opts.filter.TabIDSet = true
	}
	if limRaw, ok := flagStringSet(rest, "--limit"); ok {
		n, err := strconv.Atoi(strings.TrimSpace(limRaw))
		if err != nil || n < 0 {
			return "", opts, fmt.Errorf("--limit must be a non-negative integer")
		}
		opts.limit = n
		opts.limitSet = true
	}
	idx, idxSet, err := parseOptionalIndex(rest)
	if err != nil {
		return "", opts, err
	}
	opts.index, opts.indexSet = idx, idxSet
	return path, opts, nil
}

func cliHARInspectSummary(args []string, env map[string]string, stdout, stderr io.Writer) error {
	path, opts, err := parseHARInspectOptions(args)
	if err != nil {
		return fmt.Errorf("har inspect summary: %w", err)
	}
	src, err := loadHARInspectSource(path)
	if err != nil {
		return fmt.Errorf("har inspect summary: %w", err)
	}
	color := newServeColor(stderr, env, opts.forceColor, opts.noColor || opts.asJSON)

	type tabSummary struct {
		TabID            int    `json:"tab_id,omitempty"`
		Title            string `json:"title,omitempty"`
		File             string `json:"file"`
		Entries          int    `json:"entries"`
		FilteredEntries  int    `json:"filtered_entries"`
		WithRequestBody  int    `json:"with_request_body"`
		WithResponseBody int    `json:"with_response_body"`
	}
	tabsOut := make([]tabSummary, 0, len(src.Tabs))
	totalEntries := 0
	var allRows []harInspectRow
	for _, tab := range src.Tabs {
		totalEntries += len(tab.Entries)
		rows := (&harInspectSource{Tabs: []harInspectTab{tab}, IsDir: src.IsDir, Manifest: src.Manifest}).filteredRows(opts.filter)
		cov := bodyCoverageFromRows(rows)
		allRows = append(allRows, rows...)
		tabsOut = append(tabsOut, tabSummary{
			TabID:            tab.TabID,
			Title:            tab.Title,
			File:             tab.File,
			Entries:          len(tab.Entries),
			FilteredEntries:  cov.Entries,
			WithRequestBody:  cov.WithRequestBody,
			WithResponseBody: cov.WithResponseBody,
		})
	}
	overall := bodyCoverageFromRows(allRows)

	payload := map[string]any{
		"path":           src.Path,
		"is_dir":         src.IsDir,
		"total_entries":  totalEntries,
		"body_coverage":  overall,
		"tabs":           tabsOut,
	}
	if src.Manifest != nil {
		payload["session_id"] = src.Manifest.SessionID
		payload["capture_id"] = src.Manifest.CaptureID
		payload["partial"] = src.Manifest.Partial
		payload["request_count"] = src.Manifest.RequestCount
		payload["error_count"] = src.Manifest.ErrorCount
		payload["tab_count"] = src.Manifest.TabCount
	} else {
		payload["manifest"] = nil
	}

	if opts.asJSON {
		enc := json.NewEncoder(stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(payload); err != nil {
			return err
		}
	} else {
		_, _ = fmt.Fprintf(stdout, "HAR export  %s\n", src.Path)
		if src.Manifest != nil {
			_, _ = fmt.Fprintf(stdout, "session     %s\n", src.Manifest.SessionID)
			_, _ = fmt.Fprintf(stdout, "partial     %v\n", src.Manifest.Partial)
			_, _ = fmt.Fprintf(stdout, "requests    %d\n", src.Manifest.RequestCount)
			_, _ = fmt.Fprintf(stdout, "errors      %d\n", src.Manifest.ErrorCount)
		} else {
			_, _ = fmt.Fprintf(stdout, "manifest    (none — single .har)\n")
		}
		_, _ = fmt.Fprintf(stdout, "\n")
		_, _ = fmt.Fprintf(stdout, "%-10s %-18s %8s %8s %8s  %s\n", "Tab", "Title", "Entries", "ReqBody", "RespBody", "File")
		for _, t := range tabsOut {
			title := t.Title
			if len(title) > 18 {
				title = title[:15] + "..."
			}
			tabLabel := "-"
			if t.TabID != 0 || src.IsDir {
				tabLabel = strconv.Itoa(t.TabID)
			}
			_, _ = fmt.Fprintf(stdout, "%-10s %-18s %8d %8d %8d  %s\n",
				tabLabel, title, t.FilteredEntries, t.WithRequestBody, t.WithResponseBody, t.File)
		}
		_, _ = fmt.Fprintf(stdout, "\nbody coverage (filtered): request %d/%d, response %d/%d\n",
			overall.WithRequestBody, overall.Entries, overall.WithResponseBody, overall.Entries)
	}

	if src.Manifest != nil && src.Manifest.Partial {
		color.writeWarning(stderr, fmt.Sprintf("capture is partial (error_count=%d); treat API contracts as incomplete", src.Manifest.ErrorCount))
	}
	return nil
}

func cliHARInspectPaths(args []string, env map[string]string, stdout, stderr io.Writer) error {
	path, opts, err := parseHARInspectOptions(args)
	if err != nil {
		return fmt.Errorf("har inspect paths: %w", err)
	}
	src, err := loadHARInspectSource(path)
	if err != nil {
		return fmt.Errorf("har inspect paths: %w", err)
	}
	rows := src.filteredRows(opts.filter)
	inv := pathInventory(rows)
	if opts.limitSet && opts.limit < len(inv) {
		inv = inv[:opts.limit]
	}
	if opts.asJSON {
		enc := json.NewEncoder(stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(map[string]any{"path": src.Path, "paths": inv})
	}
	_, _ = fmt.Fprintf(stdout, "%-6s %5s  %s\n", "METHOD", "COUNT", "PATH")
	for _, p := range inv {
		_, _ = fmt.Fprintf(stdout, "%-6s %5d  %s\n", p.Method, p.Count, p.Path)
	}
	return nil
}

func cliHARInspectEntries(args []string, env map[string]string, stdout, stderr io.Writer) error {
	path, opts, err := parseHARInspectOptions(args)
	if err != nil {
		return fmt.Errorf("har inspect entries: %w", err)
	}
	src, err := loadHARInspectSource(path)
	if err != nil {
		return fmt.Errorf("har inspect entries: %w", err)
	}
	rows := src.filteredRows(opts.filter)
	limited := rows
	if opts.limit >= 0 && opts.limit < len(rows) {
		limited = rows[:opts.limit]
	}
	if opts.asJSON {
		enc := json.NewEncoder(stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(map[string]any{
			"path":    src.Path,
			"total":   len(rows),
			"limit":   opts.limit,
			"entries": limited,
		})
	}
	for _, r := range limited {
		tabPart := ""
		if src.IsDir {
			tabPart = fmt.Sprintf("tab=%d ", r.TabID)
		}
		_, _ = fmt.Fprintf(stdout, "[%3d] %s%s  %-4s  %3d  req=%d resp=%d  %s\n",
			r.Index, tabPart, r.Started, r.Method, r.Status, r.RequestBodyLen, r.ResponseBodyLen, r.Path)
	}
	if len(rows) > len(limited) {
		color := newServeColor(stderr, env, opts.forceColor, opts.noColor)
		color.writeWarning(stderr, fmt.Sprintf("showing %d of %d entries; raise --limit", len(limited), len(rows)))
	}
	return nil
}

func cliHARInspectShow(args []string, env map[string]string, stdout, stderr io.Writer) error {
	path, opts, err := parseHARInspectOptions(args)
	if err != nil {
		return fmt.Errorf("har inspect show: %w", err)
	}
	if !opts.indexSet && strings.TrimSpace(opts.filter.Match) == "" && strings.TrimSpace(opts.filter.PathSubstr) == "" {
		return fmt.Errorf("har inspect show: require --match/--path-substr or --index")
	}
	src, err := loadHARInspectSource(path)
	if err != nil {
		return fmt.Errorf("har inspect show: %w", err)
	}
	var indexPtr *int
	if opts.indexSet {
		indexPtr = &opts.index
	}
	tab, index, entry, err := findHAREntry(src, opts.filter, indexPtr, opts.indexSet)
	if err != nil {
		return fmt.Errorf("har inspect show: %w", err)
	}
	payload := buildShowPayload(tab, index, entry, opts.redact)
	enc := json.NewEncoder(stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(payload); err != nil {
		return err
	}
	// Human mode still JSON for bodies (structured); plan allowed --json always for show.
	// When not --json, still print JSON (clearest for agents); skip extra banners.
	_ = opts.asJSON
	return nil
}
