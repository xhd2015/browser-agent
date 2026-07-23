package browseragent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
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
}

// SessionNewTestHooks records injectable session-new hooks for CLI doctest.
type SessionNewTestHooks = inj.SessionNewHooks

// SessionNewConfig controls the session new operator flow (no agent-run).
type SessionNewConfig struct {
	BaseDir         string
	Addr            string
	SessionID       string
	NoOpenChrome    bool
	OpenChromeFn    func(sessionURL, extensionInstallPath string) error
	AgentRunProbeFn func(sessionID, systemPromptPath, workspaceDir string, env map[string]string) error
	Stdout          io.Writer
	Stderr          io.Writer
}

// EnsureDaemon returns daemon meta when the control plane at Addr is healthy and
// server.json matches BaseDir; otherwise invokes SpawnFn (default: detached serve)
// and polls until healthy. When client > daemon, may kill+respawn unless blocked
// by extension-connected sessions (Q1).
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
			sessions, _ := fetchDaemonSessions(baseURL)
			connected := connectedSessionIDs(sessions)
			if len(connected) > 0 {
				upgradeWarnConnected(stderr, connected, daemonVer, clientVer)
				return meta, nil
			}
			orphans := disconnectedSessionIDs(sessions)
			if len(orphans) > 0 {
				upgradeWarnOrphans(stderr, orphans)
			}
			if err := ensureDaemonKillAndRespawn(cfg, meta, stderr, orphans); err != nil {
				return DaemonMeta{}, err
			}
			return waitDaemonReady(cfg, wantAddr, timeout)
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
	killFn := cfg.KillFn
	if killFn == nil {
		killFn = func(m DaemonMeta) error {
			return KillExistingDaemon(cfg.BaseDir, defaultKillExistingTimeout)
		}
	}
	if err := killFn(meta); err != nil {
		return fmt.Errorf("kill daemon for upgrade: %w", err)
	}
	if len(orphanIDs) > 0 {
		removeSessionDirs(cfg.BaseDir, orphanIDs)
	}
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
// Chrome via OpenChromeFn, and prints operator-facing stdout. Never launches agent-run.
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

	extPath, _, err := EnsureCanonicalExtension()
	if err != nil {
		return fmt.Errorf("ensure canonical extension: %w", err)
	}

	result, err := postCreateSessionHTTP(baseURL, strings.TrimSpace(cfg.SessionID))
	if err != nil {
		return err
	}

	if !cfg.NoOpenChrome {
		openFn := cfg.OpenChromeFn
		if openFn == nil && inj.SessionNewTestHooks != nil && inj.SessionNewTestHooks.OpenChromeFn != nil {
			openFn = inj.SessionNewTestHooks.OpenChromeFn
		}
		if openFn != nil {
			if err := openFn(result.SessionURL, extPath); err != nil {
				fmt.Fprintf(stderr, "browser-agent: warning: open chrome: %v\n", err)
			}
		} else if err := openChrome(result.SessionURL, ""); err != nil {
			fmt.Fprintf(stderr, "browser-agent: warning: open chrome: %v\n", err)
		}
	}

	if err := formatSessionNewOutput(stdout, result, baseURL, extPath); err != nil {
		return err
	}
	return nil
}

type postCreateSessionResult struct {
	SessionID  string
	SessionURL string
	SessionDir string
	MetaPath   string
	SystemPath string
}

func postCreateSessionHTTP(baseURL, sessionID string) (*postCreateSessionResult, error) {
	body := map[string]string{}
	if sessionID != "" {
		body["session_id"] = sessionID
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

func formatSessionNewOutput(w io.Writer, result *postCreateSessionResult, baseURL, extPath string) error {
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
