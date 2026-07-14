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
	"time"
)

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

	sessionDir := filepath.Join(cfg.BaseDir, "sessions", cfg.SessionID)
	absSessionDir, _ := filepath.Abs(sessionDir)
	if absSessionDir == "" {
		absSessionDir = sessionDir
	}
	if err := os.MkdirAll(sessionDir, 0o755); err != nil {
		return nil, fmt.Errorf("create session dir: %w", err)
	}

	// Write SYSTEM.md playbook for the agent.
	sysPath := filepath.Join(sessionDir, "SYSTEM.md")
	if err := os.WriteFile(sysPath, []byte(FormatSystemPrompt(cfg.SessionID)), 0o644); err != nil {
		return nil, fmt.Errorf("write SYSTEM.md: %w", err)
	}
	absSysPath, err := filepath.Abs(sysPath)
	if err != nil {
		absSysPath = sysPath
	}

	// Extract embedded Chrome extension under {BaseDir}/extension/{version}/.
	extPath, extVer, err := ExtractEmbeddedExtension(cfg.BaseDir)
	if err != nil {
		return nil, fmt.Errorf("extract extension: %w", err)
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

	sess := newSession(cfg.SessionID, cfg.BaseDir)
	sess.setExtensionInstallPath(extPath)
	sess.setEmbeddedIdentity(embeddedSum.Version, embeddedSum.MD5)
	cs := newControlServer(sess)

	ln, err := net.Listen("tcp", cfg.Addr)
	if err != nil {
		return nil, fmt.Errorf("listen %s: %w", cfg.Addr, err)
	}
	// Prefer actual bound address (handles :0).
	addr := ln.Addr().String()
	cfg.Addr = addr

	baseURL := "http://" + addr
	sessionURL := baseURL + "/go?session=" + cfg.SessionID
	controlPort := controlPortFromAddr(addr)

	// Write meta.json for CLI discovery / operator tooling.
	metaPath := filepath.Join(sessionDir, "meta.json")
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
	if err := os.WriteFile(metaPath, append(metaBytes, '\n'), 0o644); err != nil {
		_ = ln.Close()
		return nil, fmt.Errorf("write meta.json: %w", err)
	}
	absMetaPath, _ := filepath.Abs(metaPath)

	srv := &http.Server{
		Handler:           cs.handler(),
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

	// Launch Chrome (unless skipped). Injector preferred; production best-effort.
	if !cfg.NoOpenChrome {
		info("chrome: opening session window…")
		if cfg.OpenChromeFn != nil {
			if err := cfg.OpenChromeFn(sessionURL, extPath); err != nil {
				warn("open chrome: %v", err)
			} else {
				info("chrome: launch requested")
			}
		} else {
			if err := openChrome(sessionURL, extPath); err != nil {
				warn("open chrome: %v", err)
			} else {
				info("chrome: launch requested")
			}
		}
	} else {
		info("chrome: skipped (--no-open-chrome)")
	}

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
		_ = srv.Shutdown(shutdownCtx)
		<-errCh
		info("stopped session-id=%s dir=%s", cfg.SessionID, absSessionDir)
		return &Result{
			ExitCode:   0,
			SessionDir: sessionDir,
		}, nil
	case err := <-errCh:
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
	if p, err := strconv.Atoi(DefaultControlPort); err == nil {
		return p
	}
	return 43761
}
