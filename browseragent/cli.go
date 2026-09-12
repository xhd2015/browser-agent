package browseragent

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	inj "github.com/xhd2015/browser-agent/browseragent/inject"
	lessflags "github.com/xhd2015/less-flags"
)

const briefUsage = `Usage: browser-agent <command> [flags]

Commands:
  serve       Blocking multi-session daemon host (default 127.0.0.1:43761)
  session     Session side-commands: session new|info|delete|eval|run|logs|screenshot|cdp|create-tab|har|list
  extension   Extension log: extension log [--browser chrome|firefox]
  har         Offline HAR inspect (summary|paths|entries|show); live capture is session har
  open-managed-chrome Open managed Chrome profile with embedded extension
  skill       Show/list/install embedded agent skills
  how-to      Print agent playbooks (e.g. how-to credentials/export)
  install-chrome-extension   Extract Chrome extension; TTY drives Load unpacked UI
  install-firefox-extension  Extract signed .xpi (open in Firefox) + temporary add-on help
  assets      Ensure/status hydrated session-page + extension assets

Flags:
  --version              Print version and exit
      --version --fetch-latest
                         Fetch and print latest published version from GitHub Releases
      --version --fetch-latest -v
                         Verbose: print local + latest and whether an update is available
  -h, --help             Show this help

Run 'browser-agent --help' for full help.
`

const fullHelp = `Usage: browser-agent <command> [flags]

browser-agent is the operator CLI for the browser-agent control plane.

Commands:
  serve [flags]              Blocking multi-session daemon host (default addr 127.0.0.1:43761)
  session <cmd> [flags]      Nested session side-commands
    session new [flags]                Ensure daemon, create session, open Chrome (no agent-run)
    session info [flags]               Print session snapshot (human default; --json for machine output)
    session delete [flags]             Remove session from registry and disk
    session list [flags]               List live sessions from daemon registry
    session eval [flags] <expr>        POST an eval job and print the result
    session run [flags] <path.js>      Read a JS file and POST a run job
    session log [flags]                Print persisted sessions/<id>/log.jsonl (attach/WS debug)
    session logs [flags]               POST a console logs job (optional --limit N)
    session screenshot [flags]         POST a screenshot job; optional -o file.png
    session cdp [flags] <Method> [json]
                                       POST a raw CDP job (method + optional params JSON)
    session create-tab [flags] [url]   POST a create_tab job (blank tab or optional URL)
    session har start [session-id]      Start Chrome HAR capture for all user tabs
    session har end [session-id]        End capture; export per-tab HARs and manifest.json
  extension <cmd> [flags]    Extension-scoped commands (unified SW logs)
    extension log [flags]              Print {baseDir}/extension-{chrome|firefox}.log.jsonl
  har inspect <cmd> <path>   Offline inspect of export dirs / .har files
    har inspect summary|paths|entries|show
  install-chrome-extension   Extract embedded extension; on TTY, Load unpacked via UI
  install-firefox-extension  Extract signed .xpi, print path, open in Firefox (install prompt)
  open-managed-chrome [url]  Open managed Chrome profile (isolated user-data-dir + extension)
  skill --list|--show|--install …
                             Embedded agent skills (see: browser-agent skill --help)
  how-to [<topic>]           Print agent playbooks (list topics, or how-to credentials/export)
  assets ensure|status       Hydrate / report session-page + extension assets

Global flags:
  -h, --help                 Show this help
      --version              Print local version and exit
      --version --fetch-latest
                             Fetch and print latest published version from GitHub Releases
      --version --fetch-latest -v
                             Verbose: print local + latest and whether an update is available

serve flags:
  --host <host>              Listen host (default: 127.0.0.1)
  --port <port>              Listen port (default: 43761)
  --base-dir <path>          Session parent directory (default: ~/.tmp/browser-agent)
  --stop                     Stop any running daemon in --base-dir and exit
  --status                   Read-only daemon status probe (prints JSON and exits)
  --kill-existing            Shutdown any existing daemon in --base-dir before starting
  --color                    Force ANSI color on operator stderr
  --no-color                 Disable ANSI color on operator stderr
  --session-id <id>          (deprecated) Fixed session id; prefer plain serve or session new
  --no-agent-run             Do not launch agent-run

serve --session-id is deprecated; start a blocking daemon host with plain serve, then
create sessions via session new or POST /v1/sessions.

session new flags:
  --session-id <id>          Session id (auto-generate sess-xxxxxx when omitted)
  --base-dir <path>          Session parent directory (default: ~/.tmp/browser-agent)
  --host <host>              Control server host (default: 127.0.0.1)
  --server-port <port>       Control server port (default: 43761; reads server.json when omitted)
  --browser chrome|firefox   Browser to open (default: chrome)
  --adhoc-browser-profile    Chrome only: isolated /tmp user-data-dir profile
                             (does not touch default Chrome). Prefer for
                             Firefox cookie import / HAR without polluting
                             the operator profile.
  --no-open-chrome           Do not launch the browser
  --no-wait                  Skip waiting for extension connection
  --no-wakeup-workaround     Do not open the background wakeup tab on attach stall.
                             Default: open a quiet background wakeup tab when needed.

open-managed-chrome flags:
  --root <dir>               Managed Chrome root (default: ~/.browser-agent/managed-chrome)
                             Layout: <root>/data (profile), <root>/extensions/browser-agent/{version}/
  [url]                      Optional navigation URL; omit for a blank new window

session list flags:
  --base-dir <path>          Session parent directory (default: ~/.tmp/browser-agent)
  --host <host>              Control server host (default: from server.json or 127.0.0.1)
  --server-port <port>       Control server port (default: from server.json or 43761)
  --json                     Emit raw JSON array of session snapshots (no ANSI)
  --color                    Force ANSI color on table output
  --no-color                 Disable ANSI color on table output

session info flags:
  --session-id <id>          Session id (or env BROWSER_AGENT_SESSION_ID)
  --base-dir <path>          Session parent directory (default: ~/.tmp/browser-agent)
  --host <host>              Control server host (default: from server.json or 127.0.0.1)
  --server-port <port>       Control server port (default: from server.json or 43761)
  --json                     Emit enriched session snapshot JSON (includes browser tabs when connected)

session delete / eval / run / logs / screenshot / cdp / create-tab flags:
  --session-id <id>          Session id (or env BROWSER_AGENT_SESSION_ID)
  --base-dir <path>          Session parent directory (default: ~/.tmp/browser-agent)
  --host <host>              Control server host (default: from server.json or 127.0.0.1)
  --server-port <port>       Control server port (default: from server.json or 43761)
  --tab-id <id>              Chrome tab id for job target (strongly recommended for agents)
  --tab-index <n>            1-based index of capturable tab in session window (unstable; prefer --tab-id)
  --timeout <ms|duration>    Job wait timeout (e.g. 60000, 60s, 1m). Defaults: eval 60s,
                             run/cdp 60s, logs/create-tab 30s, screenshot 30s

screenshot flags:
  -o, --output <file.png>    Write decoded PNG (from result base64) to path

session log flags:
  --session-id <id>          Session id (or env BROWSER_AGENT_SESSION_ID)
  --base-dir <path>          Session parent directory (default: ~/.tmp/browser-agent)
  --tail <N>                 Show only the last N lines
  --json                     Emit raw JSONL (default: pretty one line per event)

extension log flags:
  --browser chrome|firefox   Log file selector (default: chrome)
  --base-dir <path>          Daemon base directory (default: ~/.tmp/browser-agent)
  --tail <N>                 Show only the last N lines
  --json                     Emit raw JSONL (default: pretty one line per event)

logs flags (console job; requires connected extension):
  --limit <N>                Optional max log entries
  --level <level>            Optional log level filter

create-tab flags:
  --url <url>                Optional URL (positional [url] also accepted); omit for blank tab

session har flags:
  --session-id <id>          Session id (or positional id / BROWSER_AGENT_SESSION_ID)
  --timeout <ms|duration>    HAR start/end timeout (default 60s)
  -o, --output-dir <dir>     End destination (default /tmp/browser-agent-<session-slug>/)
                             Chrome only; captures all current and new user HTTP(S) tabs

har inspect flags:
  --json                     Machine-readable JSON (no ANSI)
  --host <host>              Include host (repeatable)
  --method <m>               Filter method
  --path-substr <s> / --match <s>
  --tab-id <id> / --index N  Tab / entry selection
  --no-noise (default) / --noise
  --redact (default) / --no-redact
  --limit N                  entries (default 100); optional for paths

install-chrome-extension flags:
  --open                     Force UI Load unpacked even when stdout is not a TTY
  --no-open                  Print path only; do not drive Chrome UI (default for pipes)
  --write-json-result FILE   Write open handoff JSON to FILE (implies --no-open)
  --dry-run                  Open chrome://extensions and report controls; no Load unpacked click
  --dump-tree                Dump extensions-page UI tree to stderr (no load)
  --extension-dir <path>     Unpacked folder for UI load (default: just-extracted path); still extracts embed
  --keep-old                 After load, keep other same-name extension cards (default: remove older best-effort)
  --debug-screenshot         Save a PNG after each UI step under a temp dir (path printed)
  --screenshot-dir <dir>     Save step PNGs under this directory (deterministic; enables shots)
  --color / --no-color       Force / disable ANSI on install help stdout
  (default: drive Load unpacked UI when stdout is a TTY; remove older same-name cards after load)

install-firefox-extension flags:
  --open                     Force open the signed .xpi in Firefox (install prompt)
  --no-open                  Do not open Firefox (print paths only; default for pipes)
  --color / --no-color       Force / disable ANSI on install help stdout
  (default: open .xpi when stdout is a TTY and a signed xpi is embedded)

Session resolution: --session-id flag, else BROWSER_AGENT_SESSION_ID env.
`

var cliStdin io.Reader

func cliStdinReader() io.Reader {
	if cliStdin != nil {
		return cliStdin
	}
	return strings.NewReader("")
}

// HandleCLIWithStdin dispatches CLI args with an explicit stdin (serve --stop confirm).
func HandleCLIWithStdin(args []string, env map[string]string, stdin io.Reader, stdout, stderr io.Writer) error {
	prev := cliStdin
	cliStdin = stdin
	defer func() { cliStdin = prev }()
	return HandleCLI(args, env, stdout, stderr)
}

// HandleCLI dispatches browser-agent CLI args (after the binary name).
// When env != nil, session id is resolved only from that map (no process env
// fallback). When env == nil, process env may be used.
func HandleCLI(args []string, env map[string]string, stdout, stderr io.Writer) error {
	if stdout == nil {
		stdout = io.Discard
	}
	if stderr == nil {
		stderr = io.Discard
	}
	if args == nil {
		args = []string{}
	}

	// bare → brief usage + error
	if len(args) == 0 {
		_, _ = io.WriteString(stdout, briefUsage)
		if !strings.HasSuffix(briefUsage, "\n") {
			_, _ = io.WriteString(stdout, "\n")
		}
		return fmt.Errorf("missing command; see usage")
	}

	// help
	if args[0] == "-h" || args[0] == "--help" || args[0] == "help" {
		_, _ = io.WriteString(stdout, fullHelp)
		if !strings.HasSuffix(fullHelp, "\n") {
			_, _ = io.WriteString(stdout, "\n")
		}
		return nil
	}

	// version
	if args[0] == "--version" {
		return cliVersion(args[1:], env, stdout, stderr)
	}

	cmd := args[0]
	rest := args[1:]

	switch cmd {
	case "serve":
		return cliServe(rest, env, stdout, stderr)
	case "session":
		return cliSession(rest, env, stdout, stderr)
	case "extension":
		return cliExtension(rest, env, stdout, stderr)
	case "har":
		return cliHAR(rest, env, stdout, stderr)
	case "install-chrome-extension":
		return cliInstallExt(rest, env, stdout, stderr)
	case "install-firefox-extension":
		return cliInstallFirefoxExt(rest, env, stdout, stderr)
	case "open-managed-chrome":
		return cliOpenManagedChrome(rest, env, stdout, stderr)
	case "skill":
		return cliSkill(rest, env, stdout, stderr)
	case "how-to":
		return cliHowTo(rest, env, stdout, stderr)
	case "assets":
		return cliAssets(rest, env, stdout, stderr)
	case "-h", "--help":
		_, _ = io.WriteString(stdout, fullHelp)
		return nil
	default:
		// Flat side-commands (info/eval/…) are not handlers after the nested refactor.
		_, _ = io.WriteString(stderr, briefUsage)
		return fmt.Errorf("unknown command %q; try serve, session, extension, har, open-managed-chrome, install-chrome-extension, install-firefox-extension, skill, how-to, or assets", cmd)
	}
}

// cliVersion handles --version with optional --fetch-latest and -v flags.
//
//	--version                     Print local embedded version (default).
//	--version --fetch-latest      Fetch and print latest published version from GitHub.
//	--version --fetch-latest -v   Verbose: print local + latest and comparison.
func cliVersion(args []string, env map[string]string, stdout, stderr io.Writer) error {
	fetchLatest := false
	verbose := false
	for _, a := range args {
		switch a {
		case "--fetch-latest":
			fetchLatest = true
		case "-v":
			verbose = true
		case "--help", "-h":
			_, _ = io.WriteString(stdout, fullHelp)
			if !strings.HasSuffix(fullHelp, "\n") {
				_, _ = io.WriteString(stdout, "\n")
			}
			return nil
		default:
			_, _ = io.WriteString(stderr, briefUsage)
			return fmt.Errorf("unknown flag %q for --version (try --fetch-latest or -v)", a)
		}
	}

	if !fetchLatest {
		_, _ = io.WriteString(stdout, ClientVersion())
		_, _ = io.WriteString(stdout, "\n")
		return nil
	}

	ctx := context.Background()
	latest, err := FetchLatestReleaseVersion(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch latest version: %w", err)
	}
	latest = strings.TrimSpace(latest)

	if !verbose {
		_, _ = io.WriteString(stdout, latest)
		_, _ = io.WriteString(stdout, "\n")
		return nil
	}

	local := ClientVersion()
	_, _ = fmt.Fprintf(stdout, "local:   %s\n", local)
	_, _ = fmt.Fprintf(stdout, "latest:  %s\n", latest)
	switch CompareVersion(local, latest) {
	case 0:
		_, _ = fmt.Fprintf(stdout, "up-to-date\n")
	case -1:
		_, _ = fmt.Fprintf(stdout, "update available: %s → %s\n", local, latest)
	default:
		_, _ = fmt.Fprintf(stdout, "local is newer: %s > %s\n", local, latest)
	}
	return nil
}

// cliSession dispatches nested session side-commands:
// session new|info|delete|eval|run|logs|screenshot|cdp|create-tab|har …
func cliSession(args []string, env map[string]string, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		_, _ = io.WriteString(stderr, briefUsage)
		return fmt.Errorf("session requires a subcommand: new|info|delete|eval|run|log|logs|screenshot|cdp|create-tab|har|list")
	}
	sub := args[0]
	rest := args[1:]
	switch sub {
	case "new":
		return cliSessionNew(rest, env, stdout, stderr)
	case "info":
		return cliInfo(rest, env, stdout, stderr)
	case "delete":
		return cliSessionDelete(rest, env, stdout, stderr)
	case "list":
		return cliSessionList(rest, env, stdout, stderr)
	case "eval":
		return cliEval(rest, env, stdout, stderr)
	case "run":
		return cliRun(rest, env, stdout, stderr)
	case "log":
		return cliSessionLog(rest, env, stdout, stderr)
	case "logs":
		return cliLogs(rest, env, stdout, stderr)
	case "screenshot":
		return cliScreenshot(rest, env, stdout, stderr)
	case "cdp":
		return cliCDP(rest, env, stdout, stderr)
	case "create-tab", "create_tab":
		return cliCreateTab(rest, env, stdout, stderr)
	case "har":
		return cliSessionHAR(rest, env, stdout, stderr)
	case "-h", "--help":
		_, _ = io.WriteString(stdout, fullHelp)
		if !strings.HasSuffix(fullHelp, "\n") {
			_, _ = io.WriteString(stdout, "\n")
		}
		return nil
	default:
		_, _ = io.WriteString(stderr, briefUsage)
		return fmt.Errorf("unknown session subcommand %q; try new, info, delete, list, eval, run, log, logs, screenshot, cdp, create-tab, or har", sub)
	}
}

func cliSessionDelete(args []string, env map[string]string, stdout, stderr io.Writer) error {
	if hasHelpFlag(args) {
		_, _ = io.WriteString(stdout, fullHelp)
		if !strings.HasSuffix(fullHelp, "\n") {
			_, _ = io.WriteString(stdout, "\n")
		}
		return nil
	}
	sessionID, err := resolveCLISession(args, env)
	if err != nil {
		return err
	}
	base, err := resolveCLIControlBase(args)
	if err != nil {
		return err
	}
	u := strings.TrimRight(base, "/") + "/v1/session?session=" + url.QueryEscape(sessionID)
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, u, nil)
	if err != nil {
		return err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("delete request failed: %w", err)
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	switch res.StatusCode {
	case http.StatusOK, http.StatusNoContent:
		_, err = fmt.Fprintf(stdout, "deleted session %s\n", sessionID)
		return err
	case http.StatusNotFound:
		return fmt.Errorf("session not found: %s", sessionID)
	case http.StatusConflict:
		return fmt.Errorf("cannot delete session: extension connected")
	default:
		return fmt.Errorf("delete failed: status %d: %s", res.StatusCode, strings.TrimSpace(string(body)))
	}
}

func cliSessionNew(args []string, env map[string]string, stdout, stderr io.Writer) error {
	if hasHelpFlag(args) {
		_, _ = io.WriteString(stdout, fullHelp)
		if !strings.HasSuffix(fullHelp, "\n") {
			_, _ = io.WriteString(stdout, "\n")
		}
		return nil
	}

	baseDir := flagString(args, "--base-dir")
	sessionID := flagString(args, "--session-id")
	noOpenChrome := flagBool(args, "--no-open-chrome")
	noWait := flagBool(args, "--no-wait")
	noWakeupWorkaround := flagBool(args, "--no-wakeup-workaround")
	adhocBrowserProfile := flagBool(args, "--adhoc-browser-profile")
	browserFlag := flagString(args, "--browser")
	addr := sessionNewAddrFromFlags(args)

	if baseDir == "" {
		baseDir = defaultCLIBaseDir()
	}

	browser := ""
	if browserFlag != "" {
		b, err := normalizeSessionNewBrowser(browserFlag)
		if err != nil {
			return fmt.Errorf("unknown browser %q", browserFlag)
		}
		browser = b
	}

	cfg := SessionNewConfig{
		BaseDir:             baseDir,
		Addr:                addr,
		SessionID:           sessionID,
		Browser:             browser,
		NoOpenChrome:        noOpenChrome,
		NoWait:              noWait,
		NoWakeupWorkaround:  noWakeupWorkaround,
		AdhocBrowserProfile: adhocBrowserProfile,
		Stdout:              stdout,
		Stderr:              stderr,
	}
	// Isolate ensure path when callers pass HOME via env (CLI doctests).
	if env != nil {
		if home := strings.TrimSpace(env["HOME"]); home != "" {
			cfg.Home = home
		}
	}
	// Snapshot CLI inject hooks onto cfg so SessionNew does not re-read globals.
	if cfg.OpenChromeFn == nil {
		cfg.OpenChromeFn = inj.SessionNewOpenChromeFn()
	}
	if cfg.OpenFirefoxFn == nil {
		cfg.OpenFirefoxFn = inj.SessionNewOpenFirefoxFn()
	}
	return SessionNew(cfg)
}

func cliInstallExt(args []string, env map[string]string, stdout, stderr io.Writer) error {
	opts, err := parseInstallChromeExtOptions(args, stdout)
	if err == lessflags.ErrHelp {
		// Help already printed via HelpFunc; subcommand-level --help.
		return nil
	}
	if err != nil {
		return err
	}
	if opts.forceColor && opts.noColor {
		return fmt.Errorf("--color and --no-color cannot be specified together")
	}
	if opts.open && opts.noOpen {
		return fmt.Errorf("--open and --no-open cannot be specified together")
	}
	if opts.noOpen && (opts.dryRun || opts.dumpTree || opts.open) {
		return fmt.Errorf("--no-open cannot be combined with --open, --dry-run, or --dump-tree")
	}
	jsonResult := strings.TrimSpace(opts.writeJSONResult)
	if writeJSONResultFlagPresent(args) && jsonResult == "" {
		return fmt.Errorf("--write-json-result requires a file path")
	}
	if jsonResult != "" && (opts.open || opts.dryRun || opts.dumpTree) {
		return fmt.Errorf("--write-json-result cannot be combined with --open, --dry-run, or --dump-tree")
	}

	colors := newServeColor(stdout, env, opts.forceColor, opts.noColor)

	// UI by default on TTY; --open / --no-open / dry-run / dump-tree override.
	// --write-json-result is extract-only (same as --no-open).
	openUI := false
	switch {
	case opts.noOpen || jsonResult != "":
		openUI = false
	case opts.open || opts.dryRun || opts.dumpTree:
		openUI = true
	default:
		openUI = stderrIsTTY(stdout)
	}

	shotDir, err := resolveChromeScreenshotDir(opts.debugScreenshot, opts.screenshotDir)
	if err != nil {
		return err
	}
	if shotDir != "" {
		// Always print effective dir so operators can open the folder.
		_, _ = fmt.Fprintf(stdout, "debug-screenshot dir: %s\n", shotDir)
	}

	ui := InstallChromeUIConfig{
		OpenUI:         openUI && !opts.dryRun && !opts.dumpTree,
		DryRun:         opts.dryRun,
		DumpTree:       opts.dumpTree,
		ExtensionDir:   opts.extensionDir,
		KeepOlder:      opts.keepOld,
		ScreenshotDir:  shotDir,
		JSONResultPath: jsonResult,
		Stderr:         stderr,
		Colors:         colors,
	}
	// dry-run / dump-tree still need wantUI path inside body.
	if opts.dryRun || opts.dumpTree {
		ui.OpenUI = false
		ui.DryRun = opts.dryRun
		ui.DumpTree = opts.dumpTree
	}

	home := ""
	if env != nil {
		home = strings.TrimSpace(env["HOME"])
	}
	if home != "" {
		return WithProcessEnv(map[string]string{"HOME": home}, func() error {
			return installChromeExtensionBody(stdout, opts.baseDir, ui)
		})
	}
	processEnvMu.Lock()
	defer processEnvMu.Unlock()
	return installChromeExtensionBody(stdout, opts.baseDir, ui)
}

type installChromeExtOptions struct {
	baseDir         string
	forceColor      bool
	noColor         bool
	open            bool
	noOpen          bool
	writeJSONResult string
	dryRun          bool
	dumpTree        bool
	extensionDir    string
	keepOld         bool
	debugScreenshot bool
	screenshotDir   string
}

func writeJSONResultFlagPresent(args []string) bool {
	for _, a := range args {
		if a == "--write-json-result" || strings.HasPrefix(a, "--write-json-result=") {
			return true
		}
	}
	return false
}

// resolveChromeScreenshotDir implements:
//   - neither flag → "" (no shots)
//   - --screenshot-dir PATH → PATH (deterministic; enables shots)
//   - --debug-screenshot only → temp dir browser-agent-chrome-debug-<timestamp>
//   - both → PATH wins
func resolveChromeScreenshotDir(debugScreenshot bool, screenshotDir string) (string, error) {
	dir := strings.TrimSpace(screenshotDir)
	if dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return "", fmt.Errorf("screenshot-dir: %w", err)
		}
		return dir, nil
	}
	if !debugScreenshot {
		return "", nil
	}
	dir = filepath.Join(os.TempDir(), fmt.Sprintf("browser-agent-chrome-debug-%s", time.Now().Format("20060102-150405")))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("debug-screenshot temp dir: %w", err)
	}
	return dir, nil
}

// installChromeExtHelp is subcommand-level help (flags-parsing: every level needs --help).
const installChromeExtHelp = `Usage: browser-agent install-chrome-extension [flags]

Extract the embedded Chrome extension to the canonical path and optionally
drive Load unpacked UI (macOS; TTY auto, or --open).

Flags:
  --open                     Force UI Load unpacked even when stdout is not a TTY
  --no-open                  Print path only; do not drive Chrome UI (default for pipes)
  --write-json-result FILE   Write open handoff JSON to FILE (implies --no-open)
  --dry-run                  Open chrome://extensions and report controls; no Load unpacked click
  --dump-tree                Dump extensions-page UI tree to stderr (no load)
  --extension-dir <path>     Unpacked folder for UI load (default: just-extracted path); still extracts embed
  --keep-old                 After load, keep other same-name extension cards (default: remove older best-effort)
  --debug-screenshot         Save a PNG after each UI step under a temp dir (path printed)
  --screenshot-dir <dir>     Save step PNGs under this directory (deterministic; enables shots)
  --base-dir <path>          Ignored for install path (kept for API compatibility)
  --color / --no-color       Force / disable ANSI on install help stdout
  -h, --help                 Show this help
`

func parseInstallChromeExtOptions(args []string, helpOut io.Writer) (installChromeExtOptions, error) {
	var opts installChromeExtOptions
	if helpOut == nil {
		helpOut = io.Discard
	}
	helpFn := func() {
		_, _ = io.WriteString(helpOut, installChromeExtHelp)
		if !strings.HasSuffix(installChromeExtHelp, "\n") {
			_, _ = io.WriteString(helpOut, "\n")
		}
	}
	// less-flags (go-best-practice/flags-parsing); HelpNoExit so tests do not os.Exit.
	_, err := lessflags.String("--base-dir", &opts.baseDir).
		String("--extension-dir", &opts.extensionDir).
		String("--screenshot-dir", &opts.screenshotDir).
		String("--write-json-result", &opts.writeJSONResult).
		Bool("--open", &opts.open).
		Bool("--no-open", &opts.noOpen).
		Bool("--dry-run", &opts.dryRun).
		Bool("--dump-tree", &opts.dumpTree).
		Bool("--keep-old", &opts.keepOld).
		Bool("--debug-screenshot", &opts.debugScreenshot).
		Bool("--color", &opts.forceColor).
		Bool("--no-color", &opts.noColor).
		HelpFunc("-h,--help", helpFn).
		HelpNoExit().
		Parse(args)
	if err == lessflags.ErrHelp {
		return opts, err
	}
	if err != nil {
		msg := err.Error()
		if strings.Contains(msg, "unrecognized flag") {
			return opts, fmt.Errorf("%s\nRun 'browser-agent install-chrome-extension --help' for usage.", msg)
		}
		return opts, err
	}
	return opts, nil
}

func cliInstallFirefoxExt(args []string, env map[string]string, stdout, stderr io.Writer) error {
	opts, err := parseInstallFirefoxExtOptions(args)
	if err != nil {
		return err
	}
	if opts.forceColor && opts.noColor {
		return fmt.Errorf("--color and --no-color cannot be specified together")
	}
	if opts.open && opts.noOpen {
		return fmt.Errorf("--open and --no-open cannot be specified together")
	}

	colors := newServeColor(stdout, env, opts.forceColor, opts.noColor)

	// Open .xpi in Firefox by default for interactive TTY; off for pipes/tests.
	// --open / --no-open override auto.
	openXPI := false
	switch {
	case opts.noOpen:
		openXPI = false
	case opts.open:
		openXPI = true
	default:
		openXPI = stderrIsTTY(stdout)
	}

	// Prefer HOME from env for parallel-safe isolation (matches WithHome).
	home := ""
	if env != nil {
		home = strings.TrimSpace(env["HOME"])
	}
	if home != "" {
		return WithProcessEnv(map[string]string{"HOME": home}, func() error {
			return installFirefoxExtensionBody(stdout, opts.baseDir, colors, openXPI)
		})
	}

	processEnvMu.Lock()
	defer processEnvMu.Unlock()
	return installFirefoxExtensionBody(stdout, opts.baseDir, colors, openXPI)
}

type installFirefoxExtOptions struct {
	baseDir    string
	forceColor bool
	noColor    bool
	open       bool
	noOpen     bool
}

func parseInstallFirefoxExtOptions(args []string) (installFirefoxExtOptions, error) {
	var opts installFirefoxExtOptions
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "--base-dir":
			if i+1 < len(args) {
				opts.baseDir = args[i+1]
				i++
			}
		case strings.HasPrefix(a, "--base-dir="):
			opts.baseDir = strings.TrimPrefix(a, "--base-dir=")
		case a == "--color":
			opts.forceColor = true
		case a == "--no-color":
			opts.noColor = true
		case a == "--open":
			opts.open = true
		case a == "--no-open":
			opts.noOpen = true
		case a == "-h" || a == "--help":
			// no dedicated help; ignore (parent --help covers command list)
		default:
			if strings.HasPrefix(a, "-") {
				return opts, fmt.Errorf("unrecognized flag %s\nRun 'browser-agent --help' for usage.", a)
			}
		}
	}
	return opts, nil
}

func cliOpenManagedChrome(args []string, env map[string]string, stdout, stderr io.Writer) error {
	if hasHelpFlag(args) {
		_, _ = io.WriteString(stdout, fullHelp)
		if !strings.HasSuffix(fullHelp, "\n") {
			_, _ = io.WriteString(stdout, "\n")
		}
		return nil
	}

	root := flagString(args, "--root")
	url := takeOpenManagedChromeURL(args)
	cfg := OpenManagedChromeConfig{
		Root:   root,
		URL:    url,
		Stdout: stdout,
		Stderr: stderr,
	}
	// Snapshot CLI inject hooks onto cfg so OpenManagedChrome uses call-local LaunchFn.
	if cfg.LaunchFn == nil {
		cfg.LaunchFn = inj.ManagedChromeLaunchFn()
	}
	_, err := OpenManagedChrome(cfg)
	return err
}

func takeOpenManagedChromeURL(args []string) string {
	skipNext := false
	for i := 0; i < len(args); i++ {
		if skipNext {
			skipNext = false
			continue
		}
		a := args[i]
		if a == "--root" {
			skipNext = true
			continue
		}
		if strings.HasPrefix(a, "--root=") {
			continue
		}
		if strings.HasPrefix(a, "--") {
			continue
		}
		return a
	}
	return ""
}

func cliServe(args []string, env map[string]string, stdout, stderr io.Writer) error {
	ctx := context.Background()
	return serveWithContextStdin(ctx, args, env, cliStdinReader(), stdout, stderr)
}

func cliInfo(args []string, env map[string]string, stdout, stderr io.Writer) error {
	if hasHelpFlag(args) {
		_, _ = io.WriteString(stdout, fullHelp)
		return nil
	}
	jsonMode := flagBool(args, "--json")
	sessionID, err := resolveCLISession(args, env)
	if err != nil {
		return err
	}
	base, err := resolveCLIControlBase(args)
	if err != nil {
		return err
	}
	u := strings.TrimRight(base, "/") + "/v1/session?session=" + url.QueryEscape(sessionID)
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("info request failed: %w", err)
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("info failed: status %d: %s", res.StatusCode, strings.TrimSpace(string(body)))
	}

	var snap sessionSnapshot
	if err := json.Unmarshal(body, &snap); err != nil {
		// Non-JSON control response: pass through pretty if possible.
		var buf bytes.Buffer
		if json.Valid(body) {
			_ = json.Indent(&buf, body, "", "  ")
			if buf.Len() == 0 {
				buf.Write(body)
			}
		} else {
			buf.Write(body)
		}
		out := buf.String()
		if !strings.HasSuffix(out, "\n") {
			out += "\n"
		}
		_, err = io.WriteString(stdout, out)
		return err
	}

	if jsonMode {
		return writeSessionInfoJSON(stdout, base, sessionID, snap)
	}

	colors := newServeColor(stdout, env, false, false)
	var browser map[string]any
	if snap.Extension.Connected {
		browser, _ = fetchInfoJobBrowser(base, sessionID, 15000)
	}
	var buf bytes.Buffer
	if err := FormatSessionInfo(&buf, snap, browser, FormatSessionInfoOptions{Color: colors}); err != nil {
		return err
	}
	out := buf.String()
	// CLI human output must end with a trailing newline (shell prompt safety).
	if out != "" && !strings.HasSuffix(out, "\n") {
		out += "\n"
	}
	_, err = io.WriteString(stdout, out)
	return err
}

func writeSessionInfoJSON(stdout io.Writer, base, sessionID string, snap sessionSnapshot) error {
	outObj := map[string]any{}
	raw, err := json.Marshal(snap)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(raw, &outObj); err != nil {
		return err
	}

	connected := snap.Extension.Connected
	if connected {
		browser, berr := fetchInfoJobBrowser(base, sessionID, 15000)
		if berr != nil {
			outObj["browser"] = nil
			outObj["browser_error"] = berr.Error()
		} else {
			outObj["browser"] = browser
			delete(outObj, "browser_error")
			mergeSessionInfoEnrichment(outObj, browser)
		}
	} else {
		outObj["browser"] = nil
		outObj["browser_error"] = "extension not connected"
	}

	pretty, err := json.MarshalIndent(outObj, "", "  ")
	if err != nil {
		return err
	}
	out := string(pretty)
	if !strings.HasSuffix(out, "\n") {
		out += "\n"
	}
	_, err = io.WriteString(stdout, out)
	return err
}

// extensionConnectedFromSnapshot reads extension.connected (preferred) or
// top-level connected from a GET /v1/session JSON object.
func extensionConnectedFromSnapshot(m map[string]any) bool {
	if m == nil {
		return false
	}
	if ext, ok := m["extension"].(map[string]any); ok {
		if c, ok := ext["connected"].(bool); ok {
			return c
		}
	}
	if c, ok := m["connected"].(bool); ok {
		return c
	}
	return false
}

// fetchInfoJobBrowser POSTs an info job and returns a browser map suitable for
// merging into session info stdout (tabs, version, features when present).
// On failure returns a non-nil error; caller should still print control data.
func fetchInfoJobBrowser(base, sessionID string, timeoutMS int64) (map[string]any, error) {
	if timeoutMS <= 0 {
		timeoutMS = 15000
	}
	u := strings.TrimRight(base, "/") + "/v1/jobs"
	payload := map[string]any{
		"session_id": sessionID,
		"type":       JobTypeInfo,
		"params":     map[string]any{},
		"timeout_ms": timeoutMS,
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	httpTO := time.Duration(timeoutMS)*time.Millisecond + 5*time.Second
	if httpTO < 20*time.Second {
		httpTO = 20 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), httpTO)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("info job request failed: %w", err)
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("info job failed: status %d: %s", res.StatusCode, strings.TrimSpace(string(body)))
	}
	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("info job: invalid JSON: %w", err)
	}
	if ok, _ := result["ok"].(bool); !ok {
		errMsg, _ := result["error"].(string)
		if errMsg == "" {
			errMsg = "info job failed"
		}
		return nil, fmt.Errorf("%s", errMsg)
	}
	// Prefer result.data as browser payload; otherwise use whole result sans meta.
	browser := map[string]any{}
	if data, ok := result["data"].(map[string]any); ok && data != nil {
		for k, v := range data {
			browser[k] = v
		}
	} else {
		// Fall back: surface tabs/version/features if top-level.
		for _, k := range []string{"tabs", "version", "features"} {
			if v, ok := result[k]; ok {
				browser[k] = v
			}
		}
	}
	return browser, nil
}

func cliEval(args []string, env map[string]string, stdout, stderr io.Writer) error {
	if hasHelpFlag(args) {
		_, _ = io.WriteString(stdout, fullHelp)
		return nil
	}
	sessionID, err := resolveCLISession(args, env)
	if err != nil {
		return err
	}
	expr := takePositional(args, 0)
	if strings.TrimSpace(expr) == "" {
		return fmt.Errorf("eval requires an expression argument")
	}
	return postJobAndPrint(args, sessionID, JobTypeEval, map[string]any{
		"expression": expr,
		"expr":       expr,
	}, jobTimeoutMS(args, 60000), stdout, stderr, nil)
}

func cliRun(args []string, env map[string]string, stdout, stderr io.Writer) error {
	if hasHelpFlag(args) {
		_, _ = io.WriteString(stdout, fullHelp)
		return nil
	}
	sessionID, err := resolveCLISession(args, env)
	if err != nil {
		return err
	}
	path := takePositional(args, 0)
	if strings.TrimSpace(path) == "" {
		return fmt.Errorf("run requires a .js file path argument")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		// try absolute from cwd-relative
		abs, aerr := filepath.Abs(path)
		if aerr == nil {
			data, err = os.ReadFile(abs)
			if err == nil {
				path = abs
			}
		}
		if err != nil {
			return fmt.Errorf("run: read %s: %w", path, err)
		}
	}
	source := string(data)
	return postJobAndPrint(args, sessionID, JobTypeRun, map[string]any{
		"source":     source,
		"expression": source,
		"path":       path,
	}, jobTimeoutMS(args, 60000), stdout, stderr, nil)
}

func cliExtension(args []string, env map[string]string, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		_, _ = io.WriteString(stderr, briefUsage)
		return fmt.Errorf("extension requires a subcommand: log")
	}
	sub := args[0]
	rest := args[1:]
	switch sub {
	case "log":
		return cliExtensionLog(rest, env, stdout, stderr)
	case "-h", "--help":
		_, _ = io.WriteString(stdout, fullHelp)
		if !strings.HasSuffix(fullHelp, "\n") {
			_, _ = io.WriteString(stdout, "\n")
		}
		return nil
	default:
		_, _ = io.WriteString(stderr, briefUsage)
		return fmt.Errorf("unknown extension subcommand %q; try log", sub)
	}
}

func cliExtensionLog(args []string, env map[string]string, stdout, stderr io.Writer) error {
	if hasHelpFlag(args) {
		_, _ = io.WriteString(stdout, fullHelp)
		if !strings.HasSuffix(fullHelp, "\n") {
			_, _ = io.WriteString(stdout, "\n")
		}
		return nil
	}
	baseDir := resolveCLIBaseDir(args)
	browser := flagString(args, "--browser")
	if browser == "" {
		browser = "chrome"
	}
	browser = NormalizeExtensionLogBrowser(browser)
	tail := 0
	if t := flagString(args, "--tail"); t != "" {
		n, err := strconv.Atoi(t)
		if err != nil || n < 0 {
			return fmt.Errorf("invalid --tail %q", t)
		}
		tail = n
	}
	asJSON := hasFlag(args, "--json")
	path := ExtensionLogPath(baseDir, browser)
	lines, err := ReadExtensionLogLines(baseDir, browser, tail)
	if err != nil {
		return err
	}
	fmt.Fprintf(stdout, "  path  %s\n", path)
	if len(lines) == 0 {
		fmt.Fprintln(stdout, "(empty)")
		return nil
	}
	fmt.Fprintln(stdout)
	for _, ln := range lines {
		if asJSON {
			fmt.Fprintln(stdout, ln)
		} else {
			fmt.Fprintln(stdout, FormatSessionLogPretty(ln))
		}
	}
	return nil
}

func cliSessionLog(args []string, env map[string]string, stdout, stderr io.Writer) error {
	if hasHelpFlag(args) {
		_, _ = io.WriteString(stdout, fullHelp)
		if !strings.HasSuffix(fullHelp, "\n") {
			_, _ = io.WriteString(stdout, "\n")
		}
		return nil
	}
	sessionID, err := resolveCLISession(args, env)
	if err != nil {
		return err
	}
	baseDir := strings.TrimSpace(flagString(args, "--base-dir"))
	if baseDir == "" {
		baseDir = defaultCLIBaseDir()
	}
	if !SessionDirExists(baseDir, sessionID) {
		return fmt.Errorf("session not found: %s", sessionID)
	}
	tail := 0
	if t := flagString(args, "--tail"); t != "" {
		n, err := strconv.Atoi(t)
		if err != nil || n < 0 {
			return fmt.Errorf("invalid --tail %q", t)
		}
		tail = n
	}
	asJSON := hasFlag(args, "--json")
	path := SessionLogPath(baseDir, sessionID)
	lines, err := ReadSessionLogLines(baseDir, sessionID, tail)
	if err != nil {
		return err
	}
	fmt.Fprintf(stdout, "  path  %s\n", path)
	if len(lines) == 0 {
		fmt.Fprintln(stdout, "(empty)")
		return nil
	}
	fmt.Fprintln(stdout)
	for _, ln := range lines {
		if asJSON {
			fmt.Fprintln(stdout, ln)
		} else {
			fmt.Fprintln(stdout, FormatSessionLogPretty(ln))
		}
	}
	return nil
}

func cliLogs(args []string, env map[string]string, stdout, stderr io.Writer) error {
	if hasHelpFlag(args) {
		_, _ = io.WriteString(stdout, fullHelp)
		return nil
	}
	sessionID, err := resolveCLISession(args, env)
	if err != nil {
		return err
	}
	params := map[string]any{}
	if lim := flagString(args, "--limit"); lim != "" {
		if n, err := strconv.Atoi(lim); err == nil {
			params["limit"] = n
		} else {
			params["limit"] = lim
		}
	}
	if level := flagString(args, "--level"); level != "" {
		params["level"] = level
	}
	return postJobAndPrint(args, sessionID, JobTypeLogs, params, jobTimeoutMS(args, 30000), stdout, stderr, nil)
}

func cliScreenshot(args []string, env map[string]string, stdout, stderr io.Writer) error {
	if hasHelpFlag(args) {
		_, _ = io.WriteString(stdout, fullHelp)
		return nil
	}
	sessionID, err := resolveCLISession(args, env)
	if err != nil {
		return err
	}
	params := map[string]any{
		"format":    "png",
		"full_page": false,
	}
	outPath := flagString(args, "-o")
	if outPath == "" {
		outPath = flagString(args, "--output")
	}
	return postJobAndPrint(args, sessionID, JobTypeScreenshot, params, jobTimeoutMS(args, 30000), stdout, stderr, func(result map[string]any) error {
		if outPath == "" {
			return nil
		}
		b64 := extractBase64(result)
		if b64 == "" {
			return nil
		}
		// Strip data-url prefix if present.
		if i := strings.Index(b64, ","); i >= 0 && strings.Contains(b64[:i], "base64") {
			b64 = b64[i+1:]
		}
		raw, err := base64.StdEncoding.DecodeString(b64)
		if err != nil {
			// try raw std without padding issues — still best-effort
			raw, err = base64.RawStdEncoding.DecodeString(b64)
			if err != nil {
				return fmt.Errorf("screenshot: decode base64 for -o: %w", err)
			}
		}
		if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil && filepath.Dir(outPath) != "." {
			return err
		}
		return os.WriteFile(outPath, raw, 0o644)
	})
}

func cliCDP(args []string, env map[string]string, stdout, stderr io.Writer) error {
	if hasHelpFlag(args) {
		_, _ = io.WriteString(stdout, fullHelp)
		return nil
	}
	sessionID, err := resolveCLISession(args, env)
	if err != nil {
		return err
	}
	method := takePositional(args, 0)
	if strings.TrimSpace(method) == "" {
		return fmt.Errorf("cdp requires a Method argument (e.g. Page.navigate)")
	}
	paramsJSON := takePositional(args, 1)
	params := map[string]any{
		"method": method,
	}
	if strings.TrimSpace(paramsJSON) != "" {
		var nested any
		if err := json.Unmarshal([]byte(paramsJSON), &nested); err != nil {
			return fmt.Errorf("cdp: invalid json params: %w", err)
		}
		params["params"] = nested
	}
	return postJobAndPrint(args, sessionID, JobTypeCDP, params, jobTimeoutMS(args, 60000), stdout, stderr, nil)
}

func cliCreateTab(args []string, env map[string]string, stdout, stderr io.Writer) error {
	if hasHelpFlag(args) {
		_, _ = io.WriteString(stdout, fullHelp)
		return nil
	}
	sessionID, err := resolveCLISession(args, env)
	if err != nil {
		return err
	}
	params := map[string]any{}
	url := flagString(args, "--url")
	if strings.TrimSpace(url) == "" {
		url = takePositional(args, 0)
	}
	if strings.TrimSpace(url) != "" {
		params["url"] = strings.TrimSpace(url)
	}
	// Extension defaults active:true when unspecified.
	return postJobAndPrint(args, sessionID, JobTypeCreateTab, params, jobTimeoutMS(args, 30000), stdout, stderr, nil)
}

// resolveCLITabTarget parses --tab-id / --tab-index (mutually exclusive).
// When --tab-index is set, resolves to tab_id via info job and emits a stderr warning.
func resolveCLITabTarget(args []string, stderr io.Writer, base, sessionID string) (int64, error) {
	tabIDStr, tabIDSet := flagStringSet(args, "--tab-id")
	tabIndexStr, tabIndexSet := flagStringSet(args, "--tab-index")
	if tabIDSet && tabIndexSet {
		return 0, fmt.Errorf("cannot use both --tab-id and --tab-index")
	}
	if tabIndexSet {
		if stderr != nil {
			_, _ = fmt.Fprintln(stderr, "warning: --tab-index is unstable; prefer --tab-id for job targeting")
		}
		idx, err := strconv.Atoi(strings.TrimSpace(tabIndexStr))
		if err != nil || idx < 1 {
			return 0, fmt.Errorf("invalid --tab-index %q (must be 1-based positive integer)", tabIndexStr)
		}
		browser, err := fetchInfoJobBrowser(base, sessionID, 15000)
		if err != nil {
			return 0, fmt.Errorf("resolve --tab-index: %w", err)
		}
		tabs, _ := browser["tabs"].([]any)
		for _, item := range tabs {
			tab, ok := item.(map[string]any)
			if !ok {
				continue
			}
			tabIndex := jsonNumberToInt64(tab["index"])
			if tabIndex == int64(idx) {
				id := jsonNumberToInt64(tab["id"])
				if id == 0 {
					id = jsonNumberToInt64(tab["tab_id"])
				}
				if id > 0 {
					return id, nil
				}
			}
		}
		return 0, fmt.Errorf("tab_index %d not found in session window tab list", idx)
	}
	if tabIDSet {
		id, err := strconv.ParseInt(strings.TrimSpace(tabIDStr), 10, 64)
		if err != nil || id <= 0 {
			return 0, fmt.Errorf("invalid --tab-id %q", tabIDStr)
		}
		return id, nil
	}
	return 0, nil
}

func jsonNumberToInt64(v any) int64 {
	switch n := v.(type) {
	case float64:
		return int64(n)
	case int:
		return int64(n)
	case int64:
		return n
	case json.Number:
		i, _ := n.Int64()
		return i
	default:
		return 0
	}
}

// postJobAndPrint POSTs /v1/jobs and writes the JSON result to stdout with trailing \n.
// afterOK is optional post-processing of a successful JSON result (e.g. write screenshot file).
func postJobAndPrint(args []string, sessionID, jobType string, params map[string]any, timeoutMS int64, stdout, stderr io.Writer, afterOK func(map[string]any) error) error {
	base, err := resolveCLIControlBase(args)
	if err != nil {
		return err
	}
	tabID, err := resolveCLITabTarget(args, stderr, base, sessionID)
	if err != nil {
		return err
	}
	u := strings.TrimRight(base, "/") + "/v1/jobs"

	if params == nil {
		params = map[string]any{}
	}
	if shouldAlwaysLogJob(jobType, params) && stderr != nil {
		_, _ = fmt.Fprintf(stderr, "browser-agent: cli submit session=%s type=%s tab_id=%d %s\n",
			sessionID, jobType, tabID, jobParamsSummary(jobType, params))
	}
	payload := map[string]any{
		"session_id": sessionID,
		"type":       jobType,
		"params":     params,
		"timeout_ms": timeoutMS,
	}
	if tabID > 0 {
		payload["tab_id"] = tabID
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	// HTTP client timeout slightly above job wait.
	httpTO := time.Duration(timeoutMS)*time.Millisecond + 5*time.Second
	if httpTO < 20*time.Second {
		httpTO = 20 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), httpTO)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(raw))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("%s request failed: %w", jobType, err)
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("%s failed: status %d: %s", jobType, res.StatusCode, strings.TrimSpace(string(body)))
	}

	var result map[string]any
	if err := json.Unmarshal(body, &result); err == nil {
		if ok, _ := result["ok"].(bool); !ok {
			if errMsg, _ := result["error"].(string); errMsg != "" {
				return fmt.Errorf("%s job failed: %s", jobType, errMsg)
			}
		}
		if afterOK != nil {
			if err := afterOK(result); err != nil {
				return err
			}
		}
		enc, _ := json.Marshal(result)
		out := string(enc)
		if !strings.HasSuffix(out, "\n") {
			out += "\n"
		}
		_, err = io.WriteString(stdout, out)
		return err
	}
	out := string(body)
	if !strings.HasSuffix(out, "\n") {
		out += "\n"
	}
	_, err = io.WriteString(stdout, out)
	return err
}

func extractBase64(result map[string]any) string {
	if result == nil {
		return ""
	}
	if s, ok := result["base64"].(string); ok && s != "" {
		return s
	}
	if data, ok := result["data"].(map[string]any); ok {
		if s, ok := data["base64"].(string); ok {
			return s
		}
		if s, ok := data["data"].(string); ok {
			return s
		}
	}
	return ""
}

func resolveCLISession(args []string, env map[string]string) (string, error) {
	flagVal, flagSet := flagStringSet(args, "--session-id")
	envVal, envSet := "", false
	if env != nil {
		// Injectable env: only use keys present in the map; missing = unset.
		if v, ok := env["BROWSER_AGENT_SESSION_ID"]; ok {
			envVal = v
			envSet = true
		}
	} else {
		// Process env fallback when env map is nil.
		if v, ok := os.LookupEnv("BROWSER_AGENT_SESSION_ID"); ok {
			envVal = v
			envSet = true
		}
	}
	return ResolveSessionID(flagVal, flagSet, envVal, envSet)
}

func sessionNewAddrFromFlags(args []string) string {
	if addr := flagString(args, "--addr"); strings.TrimSpace(addr) != "" {
		addr = strings.TrimSpace(addr)
		if strings.HasPrefix(addr, "http://") || strings.HasPrefix(addr, "https://") {
			addr = strings.TrimPrefix(strings.TrimPrefix(addr, "https://"), "http://")
		}
		return addr
	}
	host := flagString(args, "--host")
	port := flagInt(args, "--server-port")
	if strings.TrimSpace(host) == "" && port == 0 {
		return ""
	}
	return ComposeControlAddr(host, port)
}

func normalizeAddr(addr string) string {
	addr = strings.TrimSpace(addr)
	if addr == "" {
		return "http://127.0.0.1:43761"
	}
	if strings.HasPrefix(addr, "http://") || strings.HasPrefix(addr, "https://") {
		return addr
	}
	// host:port form
	return "http://" + addr
}

// jobTimeoutMS returns --timeout when set, otherwise defaultMS.
// Accepts integer milliseconds ("60000") or Go-style durations ("60s", "1m", "500ms").
func jobTimeoutMS(args []string, defaultMS int64) int64 {
	raw := strings.TrimSpace(flagString(args, "--timeout"))
	if raw == "" {
		return defaultMS
	}
	// Pure integer → milliseconds.
	if n, err := strconv.ParseInt(raw, 10, 64); err == nil {
		if n <= 0 {
			return defaultMS
		}
		return n
	}
	// Duration string (must include unit for time.ParseDuration).
	d, err := time.ParseDuration(raw)
	if err != nil || d <= 0 {
		return defaultMS
	}
	return d.Milliseconds()
}

func flagString(args []string, name string) string {
	v, _ := flagStringSet(args, name)
	return v
}

func flagStringSet(args []string, name string) (string, bool) {
	eqPrefix := name + "="
	for i := 0; i < len(args); i++ {
		a := args[i]
		if a == name {
			if i+1 < len(args) {
				return args[i+1], true
			}
			return "", true
		}
		if strings.HasPrefix(a, eqPrefix) {
			return strings.TrimPrefix(a, eqPrefix), true
		}
	}
	return "", false
}

func flagBool(args []string, name string) bool {
	for _, a := range args {
		if a == name || a == name+"=true" {
			return true
		}
		if a == name+"=false" {
			return false
		}
	}
	return false
}

func hasHelpFlag(args []string) bool {
	for _, a := range args {
		if a == "-h" || a == "--help" {
			return true
		}
	}
	return false
}

// takePositional returns the n-th non-flag positional argument (0-based).
func takePositional(args []string, n int) string {
	skipNext := false
	idx := 0
	for i := 0; i < len(args); i++ {
		if skipNext {
			skipNext = false
			continue
		}
		a := args[i]
		if a == "--session-id" || a == "--addr" || a == "--host" || a == "--server-port" ||
			a == "--base-dir" || a == "--root" ||
			a == "--tab-id" || a == "--tab-index" ||
			a == "--limit" || a == "--level" || a == "--output" || a == "--output-dir" || a == "-o" ||
			a == "--url" || a == "--timeout" {
			skipNext = true
			continue
		}
		if strings.HasPrefix(a, "--") {
			// --flag=value or bare boolean
			continue
		}
		// Short flags with value: -o path handled above when a == "-o".
		// Bare -o=file style:
		if strings.HasPrefix(a, "-o=") {
			continue
		}
		if strings.HasPrefix(a, "-") && a != "-" {
			// unknown short flag without value consumption
			continue
		}
		if idx == n {
			return a
		}
		idx++
	}
	return ""
}

// takeEvalExpr returns the first non-flag positional argument after "eval".
// Kept for backward compatibility with callers.
func takeEvalExpr(args []string) string {
	return takePositional(args, 0)
}
