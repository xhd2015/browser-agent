package browseragent

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// DaemonConfig controls a blocking multi-session daemon host (no auto-session).
type DaemonConfig struct {
	Addr    string // host:port; product default 127.0.0.1:43761
	BaseDir string // daemon / session root
	Stdout  io.Writer
	Stderr  io.Writer
	// DaemonVersion overrides the version reported in health/server.json (tests).
	DaemonVersion string
	// ShutdownGracePeriod delays drain before http.Server.Shutdown and sets its
	// timeout. Injectable for force-kill doctest leaves.
	ShutdownGracePeriod time.Duration
}

// Config controls a single browser-agent serve session.
type Config struct {
	Addr         string // host:port; product default 127.0.0.1:43761
	BaseDir      string // session / log root
	SessionID    string // fixed session id (required for tests)
	WorkspaceDir string // optional; passed to AgentRunFn / BuildAgentRunArgs
	NoOpenChrome bool
	NoAgentRun   bool // skip agent-run launch (tests)
	Stdout       io.Writer
	Stderr       io.Writer
	ReadyTimeout time.Duration

	// Optional injectors (nil → production best-effort; never used when flags true)
	OpenChromeFn func(sessionURL, extensionInstallPath string) error
	// AgentRunFn receives the control session id (disk / BuildAgentRunArgs input).
	// Session for agent-run is carried via BuildAgentRunArgs argv (--env / prefixed
	// --session-id); env map overlay is optional and not required for session resolve.
	AgentRunFn func(sessionID, systemPromptPath, workspaceDir string, env map[string]string) error
}

type controlHTTP struct {
	srv   *http.Server
	addr  string
	errCh chan error
}

func serveControlHTTP(ln net.Listener, handler http.Handler) *controlHTTP {
	bound := ln.Addr().String()
	srv := &http.Server{
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
	}
	errCh := make(chan error, 1)
	go func() {
		err := srv.Serve(ln)
		if err != nil && err != http.ErrServerClosed {
			errCh <- err
			return
		}
		errCh <- nil
	}()
	return &controlHTTP{srv: srv, addr: bound, errCh: errCh}
}

func (c *controlHTTP) shutdown(ctx context.Context) error {
	if err := c.srv.Shutdown(ctx); err != nil {
		return err
	}
	return <-c.errCh
}

// RunDaemon starts an empty registry control server and blocks until ctx is cancelled.
// Writes {BaseDir}/server.json on start and removes it on clean shutdown.
func RunDaemon(ctx context.Context, cfg DaemonConfig) (*Result, error) {
	cfg = applyDaemonDefaults(cfg)
	if cfg.BaseDir == "" {
		return nil, fmt.Errorf("BaseDir is required")
	}
	if err := os.MkdirAll(cfg.BaseDir, 0o755); err != nil {
		return nil, fmt.Errorf("mkdir base dir: %w", err)
	}
	info := func(format string, args ...any) {
		fmt.Fprintf(cfg.Stderr, "browser-agent: "+format+"\n", args...)
	}

	if err := CheckForeignControlPort(cfg.Addr); err != nil {
		return nil, err
	}

	ln, err := net.Listen("tcp", cfg.Addr)
	if err != nil {
		if portReachable("http://" + cfg.Addr) {
			host, portStr, _ := net.SplitHostPort(cfg.Addr)
			port, _ := strconv.Atoi(portStr)
			return nil, ForeignPortError(host, port)
		}
		low := strings.ToLower(err.Error())
		if strings.Contains(low, "in use") || strings.Contains(low, "bind") || strings.Contains(low, "address already") {
			return nil, fmt.Errorf("listen %s: port in use: %w", cfg.Addr, err)
		}
		return nil, fmt.Errorf("listen %s: %w", cfg.Addr, err)
	}
	addr := ln.Addr().String()
	cfg.Addr = addr

	registry := NewSessionRegistry(cfg.BaseDir, addr)

	baseURL := registry.BaseURL()
	daemonVer := strings.TrimSpace(cfg.DaemonVersion)
	if daemonVer == "" {
		daemonVer = ClientVersion()
	}
	metaPath := filepath.Join(cfg.BaseDir, "server.json")
	if err := WriteDaemonMeta(metaPath, DaemonMeta{
		PID:           os.Getpid(),
		Addr:          addr,
		BaseURL:       baseURL,
		BaseDir:       cfg.BaseDir,
		StartedAt:     time.Now(),
		DaemonVersion: daemonVer,
	}); err != nil {
		_ = ln.Close()
		return nil, fmt.Errorf("write server.json: %w", err)
	}

	shutdownReq := make(chan struct{}, 1)
	forceStop := make(chan struct{}, 1)
	httpSrv := serveControlHTTP(ln, NewRegistryControlHandlerConfig(registry, RegistryHandlerConfig{
		OnShutdown: func() {
			select {
			case shutdownReq <- struct{}{}:
			default:
			}
		},
		BaseDir:       cfg.BaseDir,
		DaemonVersion: daemonVer,
	}))

	registerRunningDaemon(cfg.BaseDir, &runningDaemon{
		forceStop: func() {
			select {
			case forceStop <- struct{}{}:
			default:
			}
		},
	})
	defer unregisterRunningDaemon(cfg.BaseDir)

	info("daemon started")
	info("  listen   %s", baseURL)
	info("  base-dir %s", cfg.BaseDir)
	info("  meta     %s", metaPath)
	info("create sessions via POST %s/v1/sessions", baseURL)
	info("Ctrl-C to stop")

	for {
		select {
		case <-ctx.Done():
			info("shutting down…")
			_ = shutdownDrain(0, forceStop, httpSrv)
			_ = RemoveDaemonMeta(metaPath)
			info("daemon stopped")
			return &Result{ExitCode: 0}, nil
		case <-shutdownReq:
			info("shutting down…")
			_ = shutdownDrain(cfg.ShutdownGracePeriod, forceStop, httpSrv)
			_ = RemoveDaemonMeta(metaPath)
			info("daemon stopped")
			return &Result{ExitCode: 0}, nil
		case <-forceStop:
			_ = httpSrv.srv.Close()
			_ = RemoveDaemonMeta(metaPath)
			info("daemon stopped")
			return &Result{ExitCode: 0}, nil
		case err := <-httpSrv.errCh:
			if err != nil {
				_ = RemoveDaemonMeta(metaPath)
				return &Result{ExitCode: 1}, err
			}
			_ = RemoveDaemonMeta(metaPath)
			info("daemon exited")
			return &Result{ExitCode: 0}, nil
		}
	}
}

// Result is the outcome of Run (serve until cancel).
type Result struct {
	ExitCode   int
	SessionDir string
	Stdout     string
	Stderr     string
}

// Run starts the control server and blocks until ctx is cancelled.
// Returns (*Result, error). On clean cancel, error may be context.Canceled
// or nil depending on shutdown path; harness only needs a healthy server.
//
// Progress milestones go to cfg.Stderr (discarded when nil / tests).
func Run(ctx context.Context, cfg Config) (*Result, error) {
	cfg = applyDefaults(cfg)
	if cfg.SessionID == "" {
		return nil, fmt.Errorf("SessionID is required")
	}
	if cfg.BaseDir == "" {
		return nil, fmt.Errorf("BaseDir is required")
	}

	info := func(format string, args ...any) {
		fmt.Fprintf(cfg.Stderr, "browser-agent: "+format+"\n", args...)
	}
	warn := func(format string, args ...any) {
		fmt.Fprintf(cfg.Stderr, "browser-agent: warning: "+format+"\n", args...)
	}

	// Canonical extension install path (independent of daemon base-dir).
	extPath, extVer, err := EnsureCanonicalExtension()
	if err != nil {
		return nil, fmt.Errorf("ensure canonical extension: %w", err)
	}

	// Ensure bundle-sum.js exists (write if missing from extract tree) and
	// capture embedded identity for hello match + session snapshot.
	embeddedSum, err := EnsureExtensionBundleSum(extPath, extVer)
	if err != nil {
		return nil, fmt.Errorf("ensure extension bundle-sum: %w", err)
	}
	if embeddedSum.Version == "" {
		embeddedSum.Version = extVer
	}

	ln, err := net.Listen("tcp", cfg.Addr)
	if err != nil {
		return nil, fmt.Errorf("listen %s: %w", cfg.Addr, err)
	}
	// Prefer actual bound address (handles :0).
	addr := ln.Addr().String()
	cfg.Addr = addr

	registry := NewSessionRegistry(cfg.BaseDir, addr)
	createResult, err := registry.Create(cfg.SessionID)
	if err != nil {
		_ = ln.Close()
		return nil, fmt.Errorf("create session: %w", err)
	}

	sess, ok := registry.Get(cfg.SessionID)
	if !ok {
		_ = ln.Close()
		return nil, fmt.Errorf("session %q missing after create", cfg.SessionID)
	}
	sess.setExtensionInstallPath(extPath)
	sess.setEmbeddedIdentity(embeddedSum.Version, embeddedSum.MD5)

	sessionDir := createResult.SessionDir
	absSessionDir := createResult.SessionDir
	absSysPath := createResult.SystemPath
	absMetaPath := createResult.MetaPath
	baseURL := registry.BaseURL()
	sessionURL := createResult.SessionURL
	controlPort := controlPortFromAddr(addr)

	// Enrich meta.json with extension install identity for CLI discovery.
	meta := map[string]any{
		"session_id":             cfg.SessionID,
		"addr":                   addr,
		"base_url":               baseURL,
		"session_url":            sessionURL,
		"system_prompt_path":     absSysPath,
		"extension_install_path": extPath,
		"extension_version":      embeddedSum.Version,
		"extension_md5":          embeddedSum.MD5,
		"product":                ProductName,
		"control_port":           controlPort,
	}
	metaBytes, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		_ = ln.Close()
		return nil, fmt.Errorf("marshal meta.json: %w", err)
	}
	if err := os.WriteFile(absMetaPath, append(metaBytes, '\n'), 0o644); err != nil {
		_ = ln.Close()
		return nil, fmt.Errorf("write meta.json: %w", err)
	}

	httpSrv := serveControlHTTP(ln, newRunControlHandler(registry))

	// --- Operator-facing lifecycle (stderr) ---
	info("serve started")
	info("  listen      %s", baseURL)
	info("  session-id  %s", cfg.SessionID)
	info("  session-url %s", sessionURL)
	info("  session-dir %s", absSessionDir)
	info("  meta        %s", absMetaPath)
	info("  system      %s", absSysPath)
	if extVer != "" {
		info("  extension   %s (v%s)", extPath, extVer)
	} else {
		info("  extension   %s", extPath)
	}
	info("    embedded  version=%s  md5=%s", embeddedSum.Version, embeddedSum.MD5)
	info("  ws          %s/v1/ws", baseURL)

	// Launch agent-run (unless skipped). Injector preferred; production best-effort.
	// Control session-id stays in operator logs; agent-run id + BROWSER_AGENT_SESSION_ID
	// travel via BuildAgentRunArgs argv (--session-id=prefixed, --env K=V), not process env.
	if !cfg.NoAgentRun {
		info("agent-run: launching (draft open, --no-submit) session-id=%s", cfg.SessionID)
		// Optional env map for injectors only; production launchAgentRun uses argv --env.
		env := map[string]string{
			"BROWSER_AGENT_SESSION_ID": cfg.SessionID,
		}
		if cfg.AgentRunFn != nil {
			if err := cfg.AgentRunFn(cfg.SessionID, absSysPath, cfg.WorkspaceDir, env); err != nil {
				warn("agent-run: %v", err)
			} else {
				info("agent-run: launched (argv --env BROWSER_AGENT_SESSION_ID)")
			}
		} else {
			if err := launchAgentRun(cfg.SessionID, absSysPath, cfg.WorkspaceDir); err != nil {
				warn("agent-run: %v", err)
			} else {
				info("agent-run: launched (argv --env BROWSER_AGENT_SESSION_ID)")
			}
		}
	} else {
		info("agent-run: skipped (--no-agent-run)")
	}

	info("waiting for extension WebSocket hello (Load unpacked if needed)")
	info("next:")
	info("  browser-agent session info --session-id %s", cfg.SessionID)
	info("  browser-agent session eval --session-id %s 'document.title'", cfg.SessionID)
	info("  browser-agent install-chrome-extension")
	info("Ctrl-C to stop")

	// Mismatch / md5_unknown: warn immediately on hello (orange when stderr is a TTY).
	stderrTTY := writerIsTTY(cfg.Stderr)
	sess.setOnHello(func(match string, embedded, loaded BundleSum, installPath string) {
		switch match {
		case ExtensionMatchOK, ExtensionMatchNotConnected:
			return
		default:
			msg := FormatExtensionMismatchWarning(embedded, loaded, installPath)
			msg = msg + " (" + match + ")"
			if stderrTTY {
				warn("%s", ColorOrangeIfTTY(msg, true))
			} else {
				warn("%s", msg)
			}
		}
	})
	// Log once when extension connects (hello / supports / identity match).
	go watchExtensionConnect(ctx, sess, info)

	select {
	case <-ctx.Done():
		info("shutting down…")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = httpSrv.shutdown(shutdownCtx)
		info("stopped session-id=%s dir=%s", cfg.SessionID, absSessionDir)
		return &Result{
			ExitCode:   0,
			SessionDir: sessionDir,
		}, nil
	case err := <-httpSrv.errCh:
		if err != nil {
			warn("server exited: %v", err)
			return &Result{ExitCode: 1, SessionDir: sessionDir}, err
		}
		info("server exited")
		return &Result{ExitCode: 0, SessionDir: sessionDir}, nil
	}
}

// watchExtensionConnect polls session snapshot until extension hello or ctx done.
func watchExtensionConnect(ctx context.Context, sess *session, info func(string, ...any)) {
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			snap := sess.snapshot()
			if !snap.Extension.Connected {
				continue
			}
			if snap.Extension.SupportsBrowserAgent {
				info("extension connected version=%s features=%v supports=true match=%s",
					snap.Extension.Version, snap.Extension.Features, snap.ExtensionMatch)
			} else {
				info("extension connected version=%s features=%v supports=false match=%s (need feature %q + version ≥ %s)",
					snap.Extension.Version, snap.Extension.Features, snap.ExtensionMatch, FeatureBrowserAgent, MinBrowserAgentVersion)
			}
			return
		}
	}
}

func applyDaemonDefaults(cfg DaemonConfig) DaemonConfig {
	if cfg.Addr == "" {
		cfg.Addr = DefaultAddr
	}
	if cfg.Stdout == nil {
		cfg.Stdout = io.Discard
	}
	if cfg.Stderr == nil {
		cfg.Stderr = io.Discard
	}
	return cfg
}

func applyDefaults(cfg Config) Config {
	if cfg.Addr == "" {
		cfg.Addr = DefaultAddr
	}
	if cfg.Stdout == nil {
		cfg.Stdout = io.Discard
	}
	if cfg.Stderr == nil {
		cfg.Stderr = io.Discard
	}
	if cfg.ReadyTimeout <= 0 {
		cfg.ReadyTimeout = 5 * time.Second
	}
	if cfg.BaseDir == "" {
		home, err := os.UserHomeDir()
		if err == nil {
			cfg.BaseDir = filepath.Join(home, ".tmp", "browser-agent")
		}
	}
	return cfg
}

func controlPortFromAddr(addr string) int {
	// Prefer product default when listening on the product port; otherwise parse.
	if host, portStr, err := net.SplitHostPort(addr); err == nil {
		_ = host
		if p, err := strconv.Atoi(portStr); err == nil && p > 0 {
			return p
		}
	}
	return DefaultControlPort
}
