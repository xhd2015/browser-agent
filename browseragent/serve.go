package browseragent

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	inj "github.com/xhd2015/browser-agent/browseragent/inject"
	lessflags "github.com/xhd2015/less-flags"
)

const serveHelp = `Usage: browser-agent serve [flags]

Blocking multi-session daemon host (default 127.0.0.1:43761).

Flags:
  --host <host>              Listen host (default: 127.0.0.1)
  --port <port>              Listen port (default: 43761)
  --base-dir <path>          Session parent directory (default: ~/.tmp/browser-agent)
  --stop                     Stop any running daemon in --base-dir and exit
  --status                   Read-only daemon status probe (human table; exits without starting daemon)
  --kill-existing            Shutdown any existing daemon in --base-dir before starting
  --session-id <id>          (deprecated) Fixed session id; prefer plain serve or session new
  --no-agent-run             Do not launch agent-run
  --color                    Force ANSI color on operator stderr
  --no-color                 Disable ANSI color on operator stderr
  -h, --help                 Show this help

serve --session-id is deprecated; start a blocking daemon host with plain serve, then
create sessions via session new or POST /v1/sessions.
`

// serveOptions holds parsed serve CLI flags.
type serveOptions struct {
	host         string
	port         int
	baseDir      string
	sessionID    string
	sessionIDSet bool
	noOpenChrome bool
	noAgentRun   bool
	stop         bool
	status       bool
	killExisting bool
	forceColor   bool
	noColor      bool
	showHelp     bool
}

// ServeWithContext runs the serve command with shared flag parsing used by HandleCLI
// and the browser-agent binary main path.
func ServeWithContext(ctx context.Context, args []string, env map[string]string, stdout, stderr io.Writer) error {
	return serveWithContextStdin(ctx, args, env, cliStdinReader(), stdout, stderr)
}

func serveWithContextStdin(ctx context.Context, args []string, env map[string]string, stdin io.Reader, stdout, stderr io.Writer) error {
	if stdout == nil {
		stdout = io.Discard
	}
	if stderr == nil {
		stderr = io.Discard
	}
	if args == nil {
		args = []string{}
	}
	if stdin == nil {
		stdin = strings.NewReader("")
	}

	opts, err := parseServeOptions(args, stdout)
	if err != nil {
		return err
	}
	if opts.showHelp {
		return nil
	}

	colors := newServeColor(stderr, env, opts.forceColor, opts.noColor)

	if opts.forceColor && opts.noColor {
		return fmt.Errorf("--color and --no-color cannot be specified together")
	}

	modeCount := 0
	if opts.stop {
		modeCount++
	}
	if opts.status {
		modeCount++
	}
	if opts.killExisting {
		modeCount++
	}
	if modeCount > 1 {
		colors.writeError(stderr, "--stop, --status, and --kill-existing are mutually exclusive")
		return fmt.Errorf("--stop, --status, and --kill-existing are mutually exclusive")
	}

	baseDir := opts.baseDir
	if baseDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			home = os.TempDir()
		}
		baseDir = filepath.Join(home, ".tmp", "browser-agent")
	}

	if opts.status {
		st, err := QueryDaemonStatus(baseDir)
		if err != nil {
			return err
		}
		return FormatDaemonStatus(stdout, st)
	}

	if opts.stop {
		return serveStop(opts, baseDir, colors, stdin, stderr)
	}

	listenAddr := ComposeControlAddr(opts.host, opts.port)

	if opts.sessionIDSet && opts.sessionID != "" {
		fmt.Fprintf(stderr, "browser-agent: warning: serve --session-id is deprecated; start the daemon without --session-id and create sessions via POST /v1/sessions\n")
		cfg := Config{
			Addr:         listenAddr,
			BaseDir:      baseDir,
			SessionID:    opts.sessionID,
			NoOpenChrome: opts.noOpenChrome,
			NoAgentRun:   opts.noAgentRun,
			Stdout:       stdout,
			Stderr:       stderr,
		}
		_, err := Run(ctx, cfg)
		return err
	}

	if opts.killExisting {
		fmt.Fprintf(stderr, "browser-agent: --kill-existing: shutdown/kill existing daemon until stopped in %s…\n", baseDir)
		if err := serveKillExisting(baseDir, stderr); err != nil {
			fmt.Fprintf(stderr, "browser-agent: warning: kill-existing shutdown failed: %v\n", err)
			return err
		}
		fmt.Fprintf(stderr, "browser-agent: --kill-existing: existing daemon stopped\n")
	}

	daemonCfg := DaemonConfig{
		Addr:    listenAddr,
		BaseDir: baseDir,
		Stdout:  stdout,
		Stderr:  stderr,
	}
	_, err = RunDaemon(ctx, daemonCfg)
	return err
}

func serveKillExisting(baseDir string, stderr io.Writer) error {
	st, err := QueryDaemonStatus(baseDir)
	if err != nil {
		return err
	}
	if st.Running {
		connected := connectedSessionIDs(st.Sessions)
		orphans := disconnectedSessionIDs(st.Sessions)
		if len(connected) > 0 {
			warnKillExistingConnected(stderr, connected)
		}
		for _, id := range orphans {
			fmt.Fprintf(stderr, "browser-agent: warning: disconnected session will be removed: %s\n", id)
		}
		allIDs := append(append([]string{}, connected...), orphans...)
		if err := KillExistingDaemon(baseDir, defaultKillExistingTimeout); err != nil {
			return err
		}
		if len(allIDs) > 0 {
			removeSessionDirs(baseDir, allIDs)
		} else {
			removeAllSessionDirs(baseDir)
		}
	}
	return nil
}

func serveStop(opts serveOptions, baseDir string, colors serveColor, stdin io.Reader, stderr io.Writer) error {
	st, err := QueryDaemonStatus(baseDir)
	if err != nil {
		return err
	}
	if !st.Running {
		colors.writeWarning(stderr, fmt.Sprintf("no daemon running in %s", baseDir))
		return nil
	}

	connected := connectedSessionIDs(st.Sessions)
	if len(connected) > 0 {
		n := len(connected)
		label := "sessions"
		if n == 1 {
			label = "session"
		}
		fmt.Fprintf(stderr, "browser-agent: %d extension-connected %s: %s\n", n, label, formatSessionList(connected))
	}

	if isStdinTTY(stdin) {
		fmt.Fprintf(stderr, "Stop daemon? [Y/n]\n")
		answer, _ := readStdinLine(stdin)
		if strings.HasPrefix(strings.ToLower(strings.TrimSpace(answer)), "n") {
			return nil
		}
	} else if len(connected) > 0 {
		fmt.Fprintf(stderr, "browser-agent: warning: stopping daemon with %d extension-connected session(s): %s\n",
			len(connected), formatSessionList(connected))
	}

	pid := st.PID
	if err := KillExistingDaemon(baseDir, defaultKillExistingTimeout); err != nil {
		return err
	}
	colors.writeStopped(stderr, pid, baseDir)
	return nil
}

func isStdinTTY(stdin io.Reader) bool {
	if inj.IsTerminalFn != nil {
		return inj.IsTerminalFn(stdin)
	}
	if f, ok := stdin.(*os.File); ok {
		return fileIsTTY(f)
	}
	return false
}

func readStdinLine(stdin io.Reader) (string, error) {
	if stdin == nil {
		return "", nil
	}
	sc := bufio.NewScanner(stdin)
	if sc.Scan() {
		return sc.Text(), nil
	}
	return "", sc.Err()
}

func parseServeOptions(args []string, stdout io.Writer) (serveOptions, error) {
	var (
		host           string
		port           int
		addrLegacy     string
		baseDir        string
		sessionIDPtr   *string
		noOpenChrome   bool
		noAgentRun     bool
		stop           bool
		status         bool
		killExisting   bool
		forceColor     bool
		noColor        bool
		portStr        string
	)

	helpFn := func() {
		txt := strings.TrimPrefix(serveHelp, "\n")
		_, _ = io.WriteString(stdout, txt)
		if !strings.HasSuffix(txt, "\n") {
			_, _ = io.WriteString(stdout, "\n")
		}
	}

	_, err := lessflags.String("--host", &host).
		String("--port", &portStr).
		String("--addr", &addrLegacy).
		String("--base-dir", &baseDir).
		String("--session-id", &sessionIDPtr).
		Bool("--no-open-chrome", &noOpenChrome).
		Bool("--no-agent-run", &noAgentRun).
		Bool("--stop", &stop).
		Bool("--status", &status).
		Bool("--kill-existing", &killExisting).
		Bool("--color", &forceColor).
		Bool("--no-color", &noColor).
		HelpFunc("-h,--help", helpFn).
		HelpNoExit().
		Parse(args)
	if err == lessflags.ErrHelp {
		return serveOptions{showHelp: true}, nil
	}
	if err != nil {
		msg := err.Error()
		if strings.Contains(msg, "unrecognized flag") {
			return serveOptions{}, fmt.Errorf("%s\nRun 'browser-agent serve --help' for usage.", msg)
		}
		return serveOptions{}, err
	}

	if strings.TrimSpace(portStr) != "" {
		p, perr := strconv.Atoi(strings.TrimSpace(portStr))
		if perr != nil {
			return serveOptions{}, fmt.Errorf("invalid --port %q", portStr)
		}
		port = p
	}
	if strings.TrimSpace(addrLegacy) != "" {
		legacyHost, legacyPort := parseHostPortFromAddr(addrLegacy)
		if strings.TrimSpace(host) == "" {
			host = legacyHost
		}
		if port == 0 {
			port = legacyPort
		}
	}

	sessionID := ""
	sessionIDSet := sessionIDPtr != nil
	if sessionIDSet {
		sessionID = *sessionIDPtr
	}

	opts := serveOptions{
		host:         host,
		port:         port,
		baseDir:      baseDir,
		sessionID:    sessionID,
		sessionIDSet: sessionIDSet,
		noOpenChrome: noOpenChrome,
		noAgentRun:   noAgentRun,
		stop:         stop,
		status:       status,
		killExisting: killExisting,
		forceColor:   forceColor,
		noColor:      noColor,
	}
	return opts, nil
}