package browseragent

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"

	lessflags "github.com/xhd2015/less-flags"
)

const sessionListHelp = `Usage: browser-agent session list [flags]

List live sessions from the daemon registry via GET /v1/sessions.

Flags:
  --base-dir <path>          Session parent directory (default: ~/.tmp/browser-agent)
  --host <host>              Control server host (default: from server.json or 127.0.0.1)
  --server-port <port>       Control server port (default: from server.json or 43761)
  --json                     Emit raw JSON array of session snapshots (no ANSI)
  --color                    Force ANSI color on table output
  --no-color                 Disable ANSI color on table output
  -h, --help                 Show this help

Columns (human): Session ID, Created, Pages, Browser, Status.
Footer hints suggest session delete for sessions with 0 open session pages.

When the daemon is unreachable, prints a warning on stderr and exits 0 with empty
output (human table with 0 sessions) or [] with --json.
`

// FormatSessionListOptions controls human table rendering for session list.
type FormatSessionListOptions struct {
	Color serveColor
}

// FormatSessionList writes a human-readable session table to w.
// Columns: Session ID, Created, Pages, Browser, Status. Trailing line: "N sessions".
func FormatSessionList(w io.Writer, sessions []sessionSnapshot, opts FormatSessionListOptions) error {
	if w == nil {
		w = io.Discard
	}
	if sessions == nil {
		sessions = []sessionSnapshot{}
	}

	sorted := make([]sessionSnapshot, len(sessions))
	copy(sorted, sessions)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].SessionID < sorted[j].SessionID
	})

	colID := "Session ID"
	colCreated := "Created"
	colPages := "Pages"
	colBrowser := "Browser"
	colStatus := "Status"

	idW, createdW, pagesW, browserW, statusW := len(colID), len(colCreated), len(colPages), len(colBrowser), len(colStatus)
	for _, snap := range sorted {
		if n := len(snap.SessionID); n > idW {
			idW = n
		}
		created := formatSessionAge(snap.CreatedAt)
		if created == "" {
			created = "—"
		}
		if n := len(created); n > createdW {
			createdW = n
		}
		pages := formatPageCountDisplay(snap.SessionPageCount)
		if n := len(pages); n > pagesW {
			pagesW = n
		}
		browser := formatBrowserListDisplay(snap.Browsers)
		if n := len(browser); n > browserW {
			browserW = n
		}
		label := snap.StatusLabel
		if label == "" {
			label = snap.Status
		}
		if n := len(label); n > statusW {
			statusW = n
		}
	}
	idW += 2
	createdW += 2
	pagesW += 2
	browserW += 2
	statusW += 2

	if _, err := fmt.Fprintf(w, "%-*s %-*s %-*s %-*s %-*s\n",
		idW, colID, createdW, colCreated, pagesW, colPages, browserW, colBrowser, statusW, colStatus); err != nil {
		return err
	}

	colors := opts.Color
	for _, snap := range sorted {
		created := formatSessionAge(snap.CreatedAt)
		if created == "" {
			created = "—"
		}
		pages := formatPageCountDisplay(snap.SessionPageCount)
		browser := formatBrowserListDisplay(snap.Browsers)
		label := snap.StatusLabel
		if label == "" {
			label = snap.Status
		}
		statusOut := statusColor(colors, snap.Status, label)

		if _, err := fmt.Fprintf(w, "%-*s %-*s %-*s %-*s %-*s\n",
			idW, snap.SessionID, createdW, created, pagesW, pages, browserW, browser, statusW, statusOut); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "%d sessions\n", len(sorted)); err != nil {
		return err
	}

	for _, snap := range sorted {
		if snap.SessionPageCount != nil && *snap.SessionPageCount == 0 {
			if _, err := fmt.Fprintf(w, "Hint: %s has 0 session pages — run 'browser-agent session delete --session-id %s' to clean up.\n",
				snap.SessionID, snap.SessionID); err != nil {
				return err
			}
		}
	}
	return nil
}

type sessionListOptions struct {
	baseDir    string
	addr       string
	host       string
	serverPort int
	jsonMode   bool
	forceColor bool
	noColor    bool
	showHelp   bool
}

func cliSessionList(args []string, env map[string]string, stdout, stderr io.Writer) error {
	opts, err := parseSessionListOptions(args, stdout)
	if err != nil {
		return err
	}
	if opts.showHelp {
		return nil
	}

	if opts.forceColor && opts.noColor {
		return fmt.Errorf("--color and --no-color cannot be specified together")
	}

	baseDir := opts.baseDir
	if baseDir == "" {
		baseDir = defaultCLIBaseDir()
	}

	colors := newServeColor(stdout, env, opts.forceColor, opts.noColor)

	st, err := QueryDaemonStatus(baseDir)
	if err != nil {
		return err
	}
	if !st.Running {
		colors.writeWarning(stderr, fmt.Sprintf("daemon not running in %s", baseDir))
		if opts.jsonMode {
			_, err := fmt.Fprintln(stdout, "[]")
			return err
		}
		return FormatSessionList(stdout, nil, FormatSessionListOptions{Color: colors})
	}

	var baseURL string
	var errURL error
	if strings.TrimSpace(opts.addr) != "" {
		baseURL, errURL = ResolveControlBaseURL(baseDir, opts.addr)
	} else {
		baseURL, errURL = ResolveControlBaseURLWithHostPort(baseDir, opts.host, opts.serverPort)
	}
	if errURL != nil {
		return errURL
	}

	sessions, err := fetchDaemonSessions(baseURL)
	if err != nil {
		if !daemonHealthOK(baseURL) {
			colors.writeWarning(stderr, fmt.Sprintf("daemon not running in %s", baseDir))
			if opts.jsonMode {
				_, err := fmt.Fprintln(stdout, "[]")
				return err
			}
			return FormatSessionList(stdout, nil, FormatSessionListOptions{Color: colors})
		}
		return err
	}

	if opts.jsonMode {
		if sessions == nil {
			sessions = []sessionSnapshot{}
		}
		raw, err := json.Marshal(sessions)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintln(stdout, string(raw))
		return err
	}

	return FormatSessionList(stdout, sessions, FormatSessionListOptions{Color: colors})
}

func parseSessionListOptions(args []string, stdout io.Writer) (sessionListOptions, error) {
	var (
		baseDir    string
		addr       string
		host       string
		portStr    string
		jsonMode   bool
		forceColor bool
		noColor    bool
	)

	helpFn := func() {
		txt := strings.TrimPrefix(sessionListHelp, "\n")
		_, _ = io.WriteString(stdout, txt)
		if !strings.HasSuffix(txt, "\n") {
			_, _ = io.WriteString(stdout, "\n")
		}
	}

	_, err := lessflags.String("--base-dir", &baseDir).
		String("--addr", &addr).
		String("--host", &host).
		String("--server-port", &portStr).
		Bool("--json", &jsonMode).
		Bool("--color", &forceColor).
		Bool("--no-color", &noColor).
		HelpFunc("-h,--help", helpFn).
		HelpNoExit().
		Parse(args)
	if err == lessflags.ErrHelp {
		return sessionListOptions{showHelp: true}, nil
	}
	if err != nil {
		msg := err.Error()
		if strings.Contains(msg, "unrecognized flag") {
			return sessionListOptions{}, fmt.Errorf("%s\nRun 'browser-agent session list --help' for usage.", msg)
		}
		return sessionListOptions{}, err
	}

	serverPort := 0
	if strings.TrimSpace(portStr) != "" {
		if p, perr := strconv.Atoi(strings.TrimSpace(portStr)); perr == nil {
			serverPort = p
		}
	}

	return sessionListOptions{
		baseDir:    baseDir,
		addr:       addr,
		host:       host,
		serverPort: serverPort,
		jsonMode:   jsonMode,
		forceColor: forceColor,
		noColor:    noColor,
	}, nil
}