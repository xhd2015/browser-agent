package browseragent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	inj "github.com/xhd2015/browser-agent/browseragent/inject"
)

// EnsureDaemonConfig controls EnsureDaemon health reuse vs detached spawn.
type EnsureDaemonConfig struct {
	BaseDir       string
	Addr          string
	ClientVersion string
	SpawnFn       func() error
	KillFn        func(meta DaemonMeta) error
	Stderr        io.Writer
	WaitTimeout   time.Duration

	// PrepareUpgradeFn, when non-nil, is called instead of HTTP POST
	// /v1/admin/prepare-upgrade. Errors are ignored (best-effort).
	PrepareUpgradeFn func(baseURL string, delayMS int) error
	// FetchSessionsFn, when non-nil, supplies session connectivity for waitList
	// and reattach polling instead of GET /v1/sessions.
	FetchSessionsFn func(baseURL string) ([]SessionConnView, error)
	// SleepFn, when non-nil, replaces time.Sleep for prepare delay+slack.
	SleepFn func(d time.Duration)
	// ReattachWait is the max wait after waitDaemonReady for waitList reattach.
	// Default DefaultReattachWait (30s) when <= 0.
	ReattachWait time.Duration
	// PrepareDelayMS is delay_ms for prepare-upgrade and the pre-kill sleep base.
	// Default DefaultPrepareReconnectDelayMS (1000) when <= 0.
	PrepareDelayMS int
	// PrepareSlackMS is extra pre-kill sleep after delay. Default 200 when the
	// zero-value product config applies both prepare defaults; explicit 0 with a
	// positive PrepareDelayMS means no slack.
	PrepareSlackMS int
}

// SessionNewTestHooks records injectable session-new hooks for CLI doctest.
type SessionNewTestHooks = inj.SessionNewHooks

// SessionNewConfig controls the session new operator flow (no agent-run).
type SessionNewConfig struct {
	BaseDir      string
	Addr         string
	SessionID    string
	NoOpenChrome bool
	OpenChromeFn func(sessionURL, extensionInstallPath string) error
	// OpenFirefoxFn opens Firefox with the session URL only (no extension path).
	// Used when Browser == "firefox". Tests inject this; production uses openFirefox.
	OpenFirefoxFn   func(sessionURL string) error
	AgentRunProbeFn func(sessionID, systemPromptPath, workspaceDir string, env map[string]string) error
	Stdout          io.Writer
	Stderr          io.Writer

	// Browser selects the operator browser path: "" / "chrome" (default) or "firefox".
	// Invalid values return an error.
	Browser string

	// Home, when non-empty, extracts the canonical extension under
	// {Home}/.browser-agent/... without mutating process HOME (parallel-safe
	// for doctests). Empty means use the process HOME via EnsureCanonicalExtension
	// (chrome) or EnsureCanonicalFirefoxExtension (firefox).
	Home string

	// WaitExtensionTimeout is the hard cap for extension wait.
	// Zero means adaptive default: base/cap 70s (1m+10s, covers chrome.alarms
	// period), +10s per attach stage advance (still capped).
	// When set (e.g. doctests), that duration is the hard cap (and shrinks the base).
	WaitExtensionTimeout time.Duration

	// NoWait skips waiting entirely (new --no-wait flag).
	NoWait bool

	// NoWakeupWorkaround, when true, skips the background wakeup tab on attach
	// stall. Default false: open a quiet background wakeup tab when needed.
	NoWakeupWorkaround bool
}

// EnsureDaemon returns daemon meta when the control plane at Addr is healthy and
// server.json matches BaseDir; otherwise invokes SpawnFn (default: detached serve)
// and polls until healthy. When client > daemon, upgrades via prepare-upgrade
// (best-effort), kill+respawn (session dirs kept), and wait for reattach (default
// 30s). Extension-connected sessions no longer block upgrade.
func EnsureDaemon(cfg EnsureDaemonConfig) (DaemonMeta, error) {
	if strings.TrimSpace(cfg.BaseDir) == "" {
		return DaemonMeta{}, fmt.Errorf("BaseDir is required")
	}
	timeout := cfg.WaitTimeout
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	stderr := cfg.Stderr
	if stderr == nil {
		stderr = io.Discard
	}

	wantAddr := ResolveEnsureAddr(cfg.Addr)
	if err := CheckForeignControlPort(wantAddr); err != nil {
		return DaemonMeta{}, err
	}

	clientVer := ensureDaemonClientVersion(cfg)

	meta, ok, err := tryReuseDaemon(cfg.BaseDir, wantAddr)
	if err != nil {
		return DaemonMeta{}, err
	}

	if ok {
		baseURL := daemonMetaBaseURL(meta)
		daemonVer := strings.TrimSpace(meta.DaemonVersion)
		if daemonVer == "" {
			daemonVer = fetchDaemonVersion(baseURL)
		}
		daemonVer = EffectiveDaemonVersion(daemonVer)
		meta.DaemonVersion = daemonVer

		cmp := CompareVersion(clientVer, daemonVer)
		switch {
		case cmp > 0:
			return ensureDaemonUpgradeClientNewer(cfg, meta, baseURL, daemonVer, clientVer, wantAddr, timeout, stderr)
		case cmp < 0:
			warnOlderClient(stderr, clientVer, daemonVer)
			return meta, nil
		default:
			return meta, nil
		}
	}

	spawnFn := cfg.SpawnFn
	if spawnFn == nil {
		addr := wantAddr
		spawnFn = func() error {
			return defaultSpawnDaemon(cfg.BaseDir, addr)
		}
	}
	if err := spawnFn(); err != nil {
		return DaemonMeta{}, fmt.Errorf("spawn daemon: %w", err)
	}
	return waitDaemonReady(cfg, wantAddr, timeout)
}

func ensureDaemonKillAndRespawn(cfg EnsureDaemonConfig, meta DaemonMeta, stderr io.Writer, orphanIDs []string) error {
	// orphanIDs are warned by the caller; session dirs are intentionally kept
	// so a respawned daemon can RestoreSessionsFromDisk.
	_ = orphanIDs
	_ = stderr
	killFn := cfg.KillFn
	if killFn == nil {
		killFn = func(m DaemonMeta) error {
			return KillExistingDaemon(cfg.BaseDir, defaultKillExistingTimeout)
		}
	}
	if err := killFn(meta); err != nil {
		return fmt.Errorf("kill daemon for upgrade: %w", err)
	}
	// Phase 1: do NOT removeSessionDirs — preserve orphan session directories.
	spawnFn := cfg.SpawnFn
	if spawnFn == nil {
		addr := ResolveEnsureAddr(cfg.Addr)
		spawnFn = func() error {
			return defaultSpawnDaemon(cfg.BaseDir, addr)
		}
	}
	if err := spawnFn(); err != nil {
		return fmt.Errorf("respawn daemon: %w", err)
	}
	return nil
}

// TestExported_ensureDaemonKillAndRespawn exposes ensureDaemonKillAndRespawn for
// doctests that exercise the upgrade wipe policy without exporting the helper.
func TestExported_ensureDaemonKillAndRespawn(cfg EnsureDaemonConfig, meta DaemonMeta, stderr io.Writer, orphanIDs []string) error {
	return ensureDaemonKillAndRespawn(cfg, meta, stderr, orphanIDs)
}

func waitDaemonReady(cfg EnsureDaemonConfig, wantAddr string, timeout time.Duration) (DaemonMeta, error) {
	metaPath := filepath.Join(cfg.BaseDir, "server.json")
	deadline := time.Now().Add(timeout)
	var lastErr error
	for time.Now().Before(deadline) {
		meta, err := ReadDaemonMeta(metaPath)
		if err == nil {
			if meta.BaseDir != "" && meta.BaseDir != cfg.BaseDir {
				lastErr = fmt.Errorf("server.json base_dir mismatch")
			} else {
				addr := strings.TrimSpace(meta.Addr)
				if addr == "" {
					addr = strings.TrimSpace(wantAddr)
				}
				baseURL := daemonMetaBaseURL(meta)
				if baseURL != "" && daemonHealthOK(baseURL) {
					if meta.BaseDir == "" {
						meta.BaseDir = cfg.BaseDir
					}
					if meta.Addr == "" {
						meta.Addr = addr
					}
					if meta.BaseURL == "" {
						meta.BaseURL = baseURL
					}
					if strings.TrimSpace(meta.DaemonVersion) == "" {
						meta.DaemonVersion = fetchDaemonVersion(baseURL)
					}
					return meta, nil
				}
				lastErr = fmt.Errorf("daemon not healthy at %s", baseURL)
			}
		} else {
			lastErr = err
		}
		time.Sleep(20 * time.Millisecond)
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("timeout")
	}
	return DaemonMeta{}, fmt.Errorf("daemon not ready within %s: %w", timeout, lastErr)
}

func daemonMetaBaseURL(meta DaemonMeta) string {
	baseURL := strings.TrimRight(strings.TrimSpace(meta.BaseURL), "/")
	if baseURL != "" {
		return baseURL
	}
	addr := strings.TrimSpace(meta.Addr)
	if addr == "" {
		return ""
	}
	return "http://" + addr
}

func tryReuseDaemon(baseDir, wantAddr string) (DaemonMeta, bool, error) {
	metaPath := filepath.Join(baseDir, "server.json")
	meta, ok, err := readDaemonMetaIfPresent(metaPath)
	if err != nil {
		return DaemonMeta{}, false, err
	}
	if !ok {
		return DaemonMeta{}, false, nil
	}
	if meta.BaseDir != "" && meta.BaseDir != baseDir {
		return DaemonMeta{}, false, nil
	}

	addr := strings.TrimSpace(wantAddr)
	if addr == "" {
		addr = strings.TrimSpace(meta.Addr)
	}
	if addr == "" {
		return DaemonMeta{}, false, nil
	}
	if want := strings.TrimSpace(wantAddr); want != "" {
		if recorded := strings.TrimSpace(meta.Addr); recorded != "" && recorded != want {
			return DaemonMeta{}, false, nil
		}
	}
	if meta.PID > 0 && !IsProcessAlive(meta.PID) {
		return DaemonMeta{}, false, nil
	}

	baseURL := daemonMetaBaseURL(meta)
	if baseURL == "" || !daemonHealthOK(baseURL) {
		return DaemonMeta{}, false, nil
	}

	if meta.BaseDir == "" {
		meta.BaseDir = baseDir
	}
	if meta.Addr == "" {
		meta.Addr = addr
	}
	if meta.BaseURL == "" {
		meta.BaseURL = baseURL
	}
	if strings.TrimSpace(meta.DaemonVersion) == "" {
		meta.DaemonVersion = fetchDaemonVersion(baseURL)
	}
	return meta, true, nil
}

func defaultSpawnDaemon(baseDir, addr string) error {
	addr = ResolveEnsureAddr(addr)
	if err := CheckForeignControlPort(addr); err != nil {
		return err
	}
	if bin, ok := resolveServeBinary(); ok {
		return startDetachedServe(bin, baseDir, addr)
	}
	return spawnInProcessDaemon(baseDir, addr)
}

func resolveServeBinary() (string, bool) {
	base := filepath.Base(os.Args[0])
	if strings.HasSuffix(base, ".test") || strings.Contains(base, ".test") {
		return "", false
	}
	if bin, err := exec.LookPath("browser-agent"); err == nil {
		return bin, true
	}
	if base == "browser-agent" || strings.HasPrefix(base, "browser-agent") {
		return os.Args[0], true
	}
	return "", false
}

func startDetachedServe(bin, baseDir, addr string) error {
	if err := os.MkdirAll(baseDir, 0o755); err != nil {
		return fmt.Errorf("mkdir base dir: %w", err)
	}
	logPath := filepath.Join(baseDir, "serve.log")
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return fmt.Errorf("open serve.log: %w", err)
	}

	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		_ = logFile.Close()
		return fmt.Errorf("parse addr %q: %w", addr, err)
	}

	args := []string{"serve", "--base-dir", baseDir, "--host", host, "--port", portStr}
	cmd := exec.Command(bin, args...)
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	setDetachProcAttr(cmd)
	if err := cmd.Start(); err != nil {
		_ = logFile.Close()
		return fmt.Errorf("start detached serve: %w", err)
	}
	_ = logFile.Close()
	return nil
}

func spawnInProcessDaemon(baseDir, addr string) error {
	addr = ResolveEnsureAddr(addr)
	if err := os.MkdirAll(baseDir, 0o755); err != nil {
		return fmt.Errorf("mkdir base dir: %w", err)
	}
	logPath := filepath.Join(baseDir, "serve.log")
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return fmt.Errorf("open serve.log: %w", err)
	}
	go func() {
		defer logFile.Close()
		ctx := context.Background()
		_, _ = RunDaemon(ctx, DaemonConfig{
			Addr:    addr,
			BaseDir: baseDir,
			Stdout:  logFile,
			Stderr:  logFile,
		})
	}()
	return nil
}

// SessionNew ensures the daemon, creates a session via POST /v1/sessions, opens
// the selected browser (Chrome default, or Firefox when Browser=firefox), and
// prints operator-facing stdout. Never launches agent-run.
func SessionNew(cfg SessionNewConfig) error {
	if strings.TrimSpace(cfg.BaseDir) == "" {
		return fmt.Errorf("BaseDir is required")
	}
	stdout := cfg.Stdout
	if stdout == nil {
		stdout = io.Discard
	}
	stderr := cfg.Stderr
	if stderr == nil {
		stderr = io.Discard
	}

	browser, err := normalizeSessionNewBrowser(cfg.Browser)
	if err != nil {
		return err
	}

	meta, err := EnsureDaemon(EnsureDaemonConfig{
		BaseDir:     cfg.BaseDir,
		Addr:        cfg.Addr,
		WaitTimeout: 5 * time.Second,
		Stderr:      stderr,
	})
	if err != nil {
		return err
	}

	baseURL := daemonMetaBaseURL(meta)
	if baseURL == "" {
		return fmt.Errorf("daemon meta missing base URL")
	}

	// Agents often spam session new when attach is stuck; warn before creating another.
	warnIfManyWaitingSessions(baseURL, stderr)

	extPath, err := ensureSessionNewExtension(browser, cfg.Home)
	if err != nil {
		return err
	}

	result, err := postCreateSessionHTTP(baseURL, strings.TrimSpace(cfg.SessionID), browser)
	if err != nil {
		return err
	}

	if !cfg.NoOpenChrome {
		if browser == "firefox" {
			openFn := cfg.OpenFirefoxFn
			if openFn == nil {
				openFn = inj.SessionNewOpenFirefoxFn()
			}
			var openErr error
			if openFn != nil {
				openErr = openFn(result.SessionURL)
			} else {
				openErr = openFirefox(result.SessionURL)
			}
			if openErr != nil {
				fmt.Fprintf(stderr, "browser-agent: warning: open firefox: %v\n", openErr)
				_ = AppendSessionLog(cfg.BaseDir, result.SessionID, "warn", "firefox_open", map[string]any{
					"url": result.SessionURL, "error": openErr.Error(),
				})
			} else {
				_ = AppendSessionLog(cfg.BaseDir, result.SessionID, "info", "firefox_open", map[string]any{
					"url": result.SessionURL,
				})
			}
		} else {
			openFn := cfg.OpenChromeFn
			if openFn == nil {
				openFn = inj.SessionNewOpenChromeFn()
			}
			var openErr error
			if openFn != nil {
				openErr = openFn(result.SessionURL, extPath)
			} else {
				// extPath is unused by BuildChromeArgs (no --load-extension on default
				// profile). Operator Load-unpacked Chrome is the attach target.
				openErr = openChrome(result.SessionURL, extPath)
			}
			if openErr != nil {
				fmt.Fprintf(stderr, "browser-agent: warning: open chrome: %v\n", openErr)
				_ = AppendSessionLog(cfg.BaseDir, result.SessionID, "warn", "chrome_open", map[string]any{
					"url": result.SessionURL, "error": openErr.Error(),
				})
			} else {
				_ = AppendSessionLog(cfg.BaseDir, result.SessionID, "info", "chrome_open", map[string]any{
					"url": result.SessionURL,
				})
			}
		}
	}

	// Always print operator instructions first so URL / install path are visible
	// before any wait (or if the wait times out).
	if err := formatSessionNewOutput(stdout, result, baseURL, extPath, browser); err != nil {
		return err
	}

	// Then wait for extension connection (unless NoOpenChrome or NoWait).
	if !cfg.NoOpenChrome && !cfg.NoWait {
		if err := waitForExtensionConnection(cfg.BaseDir, baseURL, result.SessionID, browser, extPath, cfg.WaitExtensionTimeout, stderr, !cfg.NoWakeupWorkaround, nil); err != nil {
			return err
		}
	}
	return nil
}

// normalizeSessionNewBrowser returns "chrome" or "firefox".
// Empty / omitted browser defaults to chrome.
func normalizeSessionNewBrowser(browser string) (string, error) {
	b := strings.ToLower(strings.TrimSpace(browser))
	switch b {
	case "", "chrome":
		return "chrome", nil
	case "firefox":
		return "firefox", nil
	default:
		return "", fmt.Errorf("unknown browser %q", browser)
	}
}

// ensureSessionNewExtension extracts the canonical extension for the selected browser.
func ensureSessionNewExtension(browser, home string) (string, error) {
	home = strings.TrimSpace(home)
	if browser == "firefox" {
		if home != "" {
			p, _, err := EnsureCanonicalFirefoxExtensionWithHome(home)
			if err != nil {
				return "", fmt.Errorf("ensure canonical firefox extension: %w", err)
			}
			return p, nil
		}
		p, _, err := EnsureCanonicalFirefoxExtension()
		if err != nil {
			return "", fmt.Errorf("ensure canonical firefox extension: %w", err)
		}
		return p, nil
	}
	if home != "" {
		p, _, err := EnsureCanonicalExtensionWithHome(home)
		if err != nil {
			return "", fmt.Errorf("ensure canonical extension: %w", err)
		}
		return p, nil
	}
	p, _, err := EnsureCanonicalExtension()
	if err != nil {
		return "", fmt.Errorf("ensure canonical extension: %w", err)
	}
	return p, nil
}

type postCreateSessionResult struct {
	SessionID  string
	SessionURL string
	SessionDir string
	MetaPath   string
	SystemPath string
}

func postCreateSessionHTTP(baseURL, sessionID, browser string) (*postCreateSessionResult, error) {
	body := map[string]string{}
	if sessionID != "" {
		body["session_id"] = sessionID
	}
	// Stamp browser on the daemon Create so meta/snap use the matching extension tree.
	// Empty / chrome omit the field (daemon defaults to Chrome).
	if b := strings.ToLower(strings.TrimSpace(browser)); b == "firefox" {
		body["browser"] = "firefox"
	} else if b == "chrome" {
		// Explicit chrome is fine; optional. Leave out to preserve empty-body chrome default.
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	u := strings.TrimRight(baseURL, "/") + "/v1/sessions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("POST /v1/sessions: %w", err)
	}
	defer res.Body.Close()
	out, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if res.StatusCode == http.StatusConflict {
		msg := strings.TrimSpace(string(out))
		return nil, fmt.Errorf("POST /v1/sessions: session already exists (409): %s", msg)
	}
	if res.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("POST /v1/sessions: status %d: %s", res.StatusCode, strings.TrimSpace(string(out)))
	}

	var parsed map[string]string
	if err := json.Unmarshal(out, &parsed); err != nil {
		return nil, fmt.Errorf("POST /v1/sessions: parse response: %w", err)
	}
	result := &postCreateSessionResult{
		SessionID:  parsed["session_id"],
		SessionURL: parsed["session_url"],
		SessionDir: parsed["session_dir"],
		MetaPath:   parsed["meta_path"],
		SystemPath: parsed["system_path"],
	}
	if result.SessionID == "" {
		return nil, fmt.Errorf("POST /v1/sessions: missing session_id in response")
	}
	if result.SessionURL == "" {
		result.SessionURL = strings.TrimRight(baseURL, "/") + "/go?session=" + result.SessionID
	}
	return result, nil
}

func formatSessionNewOutput(w io.Writer, result *postCreateSessionResult, baseURL, extPath, browser string) error {
	if w == nil {
		w = io.Discard
	}
	if result == nil {
		return fmt.Errorf("missing create session result")
	}
	if strings.EqualFold(strings.TrimSpace(browser), "firefox") {
		return formatSessionNewFirefoxOutput(w, result, baseURL, extPath)
	}
	lines := []string{
		fmt.Sprintf("session-id: %s", result.SessionID),
		"",
		fmt.Sprintf("export BROWSER_AGENT_SESSION_ID=%s", result.SessionID),
		"",
		fmt.Sprintf("Session URL: %s", result.SessionURL),
		fmt.Sprintf("Control:     %s", baseURL),
		"",
		"Extension:",
		fmt.Sprintf("  path    %s", extPath),
		"  install browser-agent install-chrome-extension",
		"",
		"Note:",
		"  Chrome 137+ cannot auto-load extensions. Load unpacked once in your Chrome",
		"  (chrome://extensions → Developer mode → Load unpacked → path above).",
		"",
		"Next:",
		fmt.Sprintf("  browser-agent session info --session-id %s", result.SessionID),
		fmt.Sprintf("  browser-agent session eval --session-id %s 'document.title'", result.SessionID),
		fmt.Sprintf("  browser-agent session run --session-id %s script.js", result.SessionID),
		fmt.Sprintf("  browser-agent session log --session-id %s", result.SessionID),
		fmt.Sprintf("  browser-agent session logs --session-id %s", result.SessionID),
		fmt.Sprintf("  browser-agent session screenshot --session-id %s -o out.png", result.SessionID),
		fmt.Sprintf("  browser-agent session cdp --session-id %s Page.navigate '{\"url\":\"https://example.com\"}'", result.SessionID),
		"",
	}
	for _, line := range lines {
		if _, err := fmt.Fprintln(w, strings.TrimRight(line, "\n")); err != nil {
			return err
		}
	}
	return nil
}

func formatSessionNewFirefoxOutput(w io.Writer, result *postCreateSessionResult, baseURL, extPath string) error {
	if w == nil {
		w = io.Discard
	}
	if result == nil {
		return fmt.Errorf("missing create session result")
	}
	lines := []string{
		fmt.Sprintf("session-id: %s", result.SessionID),
		"",
		fmt.Sprintf("export BROWSER_AGENT_SESSION_ID=%s", result.SessionID),
		"",
		fmt.Sprintf("Session URL: %s", result.SessionURL),
		fmt.Sprintf("Control:     %s", baseURL),
		fmt.Sprintf("browser:     firefox"),
		"",
		"Extension:",
		fmt.Sprintf("  path    %s", extPath),
		"  install browser-agent install-firefox-extension",
		"",
		"Install / load the temporary add-on:",
		"",
		"  1. Open about:debugging#/runtime/this-firefox",
		"     (or type about:debugging in the address bar → This Firefox)",
		"  2. Click Load Temporary Add-on…",
		"  3. Select this folder's manifest.json:",
		"",
		fmt.Sprintf("     %s", extPath),
		"",
		"Note: temporary add-ons unload when Firefox restarts — re-run",
		"install-firefox-extension and Load Temporary Add-on after a restart.",
		"",
		"Next:",
		fmt.Sprintf("  browser-agent session info --session-id %s", result.SessionID),
		fmt.Sprintf("  browser-agent session eval --session-id %s 'document.title'", result.SessionID),
		fmt.Sprintf("  browser-agent session run --session-id %s script.js", result.SessionID),
		fmt.Sprintf("  browser-agent session log --session-id %s", result.SessionID),
		fmt.Sprintf("  browser-agent session logs --session-id %s", result.SessionID),
		fmt.Sprintf("  browser-agent session screenshot --session-id %s -o out.png", result.SessionID),
		fmt.Sprintf("  browser-agent session cdp --session-id %s Page.navigate '{\"url\":\"https://example.com\"}'", result.SessionID),
		"",
	}
	for _, line := range lines {
		if _, err := fmt.Fprintln(w, strings.TrimRight(line, "\n")); err != nil {
			return err
		}
	}
	return nil
}

// extensionTimeoutUserHandling is the fixed stderr banner after a soft wait
// timeout so agents/operators treat install as manual user handling.
const extensionTimeoutUserHandling = "Please run or ask user to run manually: this needs user handling"

// waitingExtensionWarnThreshold: session new warns when this many sessions are
// already waiting for extension connection (agents tend to spam session new).
const waitingExtensionWarnThreshold = 3

// isReceivingEndError reports Chrome's classic "service worker not listening"
// last_error from attach stage tracking.
func isReceivingEndError(err string) bool {
	return strings.Contains(strings.ToLower(err), "receiving end does not exist")
}

// formatExtensionTimeoutHelp returns stderr lines after the soft wait timeout
// warning. browser is "chrome" or "firefox"; extPath, sessionID, and lastError
// may be empty. When lastError is Receiving end, Chrome guidance leads with
// reload/popup (SW asleep) before install steps.
func formatExtensionTimeoutHelp(browser, extPath, sessionID, lastError string) string {
	browser = strings.ToLower(strings.TrimSpace(browser))
	extPath = strings.TrimSpace(extPath)
	sessionID = strings.TrimSpace(sessionID)

	var b strings.Builder
	b.WriteString(extensionTimeoutUserHandling)
	b.WriteByte('\n')
	b.WriteByte('\n')

	if browser != "firefox" && isReceivingEndError(lastError) {
		b.WriteString("  Extension service worker looks asleep (Receiving end does not exist).\n")
		b.WriteString("  Open the Browser Agent toolbar popup, or chrome://extensions → Browser Agent → Reload.\n")
		b.WriteString("  Or omit --no-wakeup-workaround (background wakeup is on by default).\n")
		b.WriteString("  Keep the /go?session= tab open; do not run session new again.\n")
		b.WriteByte('\n')
	}

	if browser == "firefox" {
		b.WriteString("  browser-agent install-firefox-extension\n")
		b.WriteByte('\n')
		b.WriteString("  Or temporary add-on: about:debugging#/runtime/this-firefox → Load Temporary Add-on…\n")
		if extPath != "" {
			b.WriteString("    ")
			b.WriteString(extPath)
			b.WriteByte('\n')
		}
	} else {
		b.WriteString("  browser-agent install-chrome-extension\n")
		b.WriteByte('\n')
		b.WriteString("  Or load unpacked: chrome://extensions → Developer mode → Load unpacked →\n")
		if extPath != "" {
			b.WriteString("    ")
			b.WriteString(extPath)
			b.WriteByte('\n')
		}
	}

	b.WriteByte('\n')
	b.WriteString("  Then reuse the same session (do not session new again):\n")
	if sessionID != "" {
		b.WriteString("    browser-agent session info --session-id ")
		b.WriteString(sessionID)
		b.WriteByte('\n')
	} else {
		b.WriteString("    browser-agent session info --session-id <session-id>\n")
	}
	return b.String()
}

// warnIfManyWaitingSessions prints a stderr warning when many daemon sessions
// are already disconnected / waiting_extension. Best-effort; ignores fetch errors.
func warnIfManyWaitingSessions(baseURL string, stderr io.Writer) {
	if stderr == nil {
		stderr = io.Discard
	}
	list, err := fetchDaemonSessions(strings.TrimRight(baseURL, "/"))
	if err != nil {
		return
	}
	waiting := 0
	for _, snap := range list {
		if !snap.Extension.Connected {
			waiting++
		}
	}
	if waiting < waitingExtensionWarnThreshold {
		return
	}
	fmt.Fprintf(stderr, "warning: %d sessions are already waiting for extension connection; prefer reusing an open /go tab (session info / session list) instead of session new. If attach is stuck with Receiving end does not exist, open the Browser Agent toolbar popup or reload the extension.\n", waiting)
}

// SW wakeup during attach wait: open chrome-extension://…/wakeup.html once when
// still stuck at register_sent (MV3 SW often ignores page sendMessage/Port).
const (
	swWakeupMinElapsed  = 2 * time.Second
	swWakeupMinAttempts = 5
)

// shouldOpenSWWakeup reports whether the one-shot extension wakeup page should open.
func shouldOpenSWWakeup(elapsed time.Duration, stage string, attempts int, alreadyOpened bool) bool {
	if alreadyOpened {
		return false
	}
	if stage != AttachStageRegisterSent {
		return false
	}
	if elapsed < swWakeupMinElapsed {
		return false
	}
	if attempts < swWakeupMinAttempts {
		return false
	}
	return true
}

// waitForExtensionConnection polls GET /v1/session?session=<id> every 500ms until
// the extension connects with browser-agent support, connects without support, or
// the adaptive deadline is reached. Deadline starts at base (70s default) and
// extends by grant (10s) on each attach stage advance, hard-capped at cap (70s
// default, or explicitTimeout when > 0). Progress ticks go to stderr. On stall,
// a warning plus install help is printed and nil is returned (session usable).
//
// When enableWakeup is true (default for session new), on attach stall threshold
// opens a background wakeup tab once (openWakeup, default openExtensionWakeup).
// When false (--no-wakeup-workaround), never opens wakeup.
// Tests pass a recorder / no-op for openWakeup to avoid launching Chrome.
func waitForExtensionConnection(baseDir, baseURL, sessionID, browser, extPath string, explicitTimeout time.Duration, stderr io.Writer, enableWakeup bool, openWakeup func(extensionInstallPath string) error) error {
	if stderr == nil {
		stderr = io.Discard
	}
	if openWakeup == nil {
		openWakeup = openExtensionWakeup
	}
	base, grant, cap := attachDeadline(explicitTimeout)
	pollInterval := 500 * time.Millisecond
	progressEvery := 5 * time.Second
	started := time.Now()
	nextProgress := started.Add(progressEvery)
	deadline := started.Add(base)
	lastStageRank := -1
	lastStage := ""
	wakeupOpened := false

	capLabel := cap.Truncate(time.Second)
	if capLabel < time.Second {
		capLabel = cap
	}
	baseLabel := base.Truncate(time.Second)
	if baseLabel < time.Second {
		baseLabel = base
	}
	fmt.Fprintf(stderr, "Waiting for extension (adaptive, base %s, cap %s)…\n", baseLabel, capLabel)
	_ = AppendSessionLog(baseDir, sessionID, "info", "wait_start", map[string]any{
		"base_ms": base.Milliseconds(),
		"cap_ms":  cap.Milliseconds(),
	})

	for {
		snap, err := pollSessionSnapshot(baseURL, sessionID)
		if err != nil {
			return fmt.Errorf("daemon unreachable during extension wait: %w", err)
		}
		stage := ""
		attempts := 0
		lastErr := ""
		if snap.Attach != nil {
			stage = snap.Attach.Stage
			attempts = snap.Attach.RegisterAttempts
			lastErr = snap.Attach.LastError
		}

		if snap.Extension.Connected {
			if snap.Extension.SupportsBrowserAgent {
				fmt.Fprintln(stderr, "Extension connected ✓")
				_ = AppendSessionLog(baseDir, sessionID, "info", "wait_connected", map[string]any{
					"elapsed_ms": time.Since(started).Milliseconds(),
					"stage":      stage,
				})
				return nil
			}
			return fmt.Errorf("Error: extension does not support browser-agent. Please install the bundled extension.")
		}

		elapsedNow := time.Since(started)
		if enableWakeup && shouldOpenSWWakeup(elapsedNow, stage, attempts, wakeupOpened) {
			wakeupOpened = true
			fmt.Fprintln(stderr, "  … waking extension service worker (background wakeup tab)")
			fields := map[string]any{
				"stage":      stage,
				"attempts":   attempts,
				"elapsed_ms": elapsedNow.Milliseconds(),
			}
			if err := openWakeup(extPath); err != nil {
				fields["error"] = err.Error()
				_ = AppendSessionLog(baseDir, sessionID, "warn", "sw_wakeup_open", fields)
				fmt.Fprintf(stderr, "warning: extension wakeup open failed: %v\n", err)
			} else {
				_ = AppendSessionLog(baseDir, sessionID, "info", "sw_wakeup_open", fields)
			}
		}

		rank := attachStageRank(stage)
		if rank > lastStageRank {
			lastStageRank = rank
			lastStage = stage
			// Extend deadline on stage advance; never past hard cap.
			extended := time.Now().Add(grant)
			hardCap := started.Add(cap)
			if extended.After(hardCap) {
				extended = hardCap
			}
			if extended.After(deadline) {
				deadline = extended
			}
			elapsed := time.Since(started).Truncate(time.Second)
			if stage != "" {
				if attempts > 0 {
					fmt.Fprintf(stderr, "  … stage=%s register_attempts=%d (%s)\n", stage, attempts, elapsed)
				} else {
					fmt.Fprintf(stderr, "  … stage=%s (%s)\n", stage, elapsed)
				}
				nextProgress = time.Now().Add(progressEvery)
			}
		} else if stage != "" {
			lastStage = stage
		}

		now := time.Now()
		if now.After(deadline) {
			elapsed := now.Sub(started).Truncate(time.Second)
			if elapsed < time.Second {
				elapsed = now.Sub(started)
			}
			if lastStage != "" && lastStage != AttachStageNoPage {
				fmt.Fprintf(stderr, "warning: extension attach stalled at stage=%s after %s\n", lastStage, elapsed)
				if lastErr != "" {
					fmt.Fprintf(stderr, "  last_error: %s\n", lastErr)
				}
				if attempts > 0 {
					fmt.Fprintf(stderr, "  register_attempts: %d\n", attempts)
				}
				// Keep legacy substring for existing doctests / agents.
				fmt.Fprintf(stderr, "warning: extension did not connect within %s\n", elapsed)
			} else {
				fmt.Fprintf(stderr, "warning: extension did not connect within %s\n", elapsed)
			}
			fmt.Fprint(stderr, formatExtensionTimeoutHelp(browser, extPath, sessionID, lastErr))
			fields := map[string]any{
				"elapsed_ms": elapsed.Milliseconds(),
				"stage":      lastStage,
				"attempts":   attempts,
			}
			if lastErr != "" {
				fields["last_error"] = lastErr
			}
			_ = AppendSessionLog(baseDir, sessionID, "warn", "wait_timeout", fields)
			return nil
		}
		if !now.Before(nextProgress) {
			elapsed := now.Sub(started).Truncate(time.Second)
			if stage != "" {
				msg := fmt.Sprintf("  … stage=%s (%s)", stage, elapsed)
				if attempts > 0 {
					msg = fmt.Sprintf("  … stage=%s register_attempts=%d (%s)", stage, attempts, elapsed)
				}
				fmt.Fprintln(stderr, msg)
			} else {
				fmt.Fprintf(stderr, "  … still waiting (%s)\n", elapsed)
			}
			nextProgress = now.Add(progressEvery)
		}

		time.Sleep(pollInterval)
	}
}

// pollSessionSnapshot fetches GET /v1/session?session=<id>.
func pollSessionSnapshot(baseURL, sessionID string) (sessionSnapshot, error) {
	u := strings.TrimRight(baseURL, "/") + "/v1/session?session=" + url.QueryEscape(sessionID)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return sessionSnapshot{}, err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return sessionSnapshot{}, err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return sessionSnapshot{}, err
	}
	if res.StatusCode != http.StatusOK {
		return sessionSnapshot{}, fmt.Errorf("GET /v1/session status %d: %s", res.StatusCode, strings.TrimSpace(string(body)))
	}
	var snap sessionSnapshot
	if err := json.Unmarshal(body, &snap); err != nil {
		return sessionSnapshot{}, fmt.Errorf("parse session snapshot: %w", err)
	}
	return snap, nil
}

// pollSessionExtension fetches a session snapshot and returns extension.connected
// and extension.supports_browser_agent.
func pollSessionExtension(baseURL, sessionID string) (connected, supports bool, err error) {
	snap, err := pollSessionSnapshot(baseURL, sessionID)
	if err != nil {
		return false, false, err
	}
	return snap.Extension.Connected, snap.Extension.SupportsBrowserAgent, nil
}
