// Package browsertrace implements the browser-trace control server and session lifecycle.
package browsertrace

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	DefaultAddr            = "127.0.0.1:43759"
	DefaultReadyTimeout    = 30 * time.Second
	DefaultCompleteTimeout = 30 * time.Second
	// DefaultReadyHeartbeat is how often ready-wait progress is logged.
	DefaultReadyHeartbeat = 5 * time.Second
	// DefaultHeartbeatTimeout is how long the recording phase may go without
	// POST /v1/status or POST /v1/entries before heartbeat_lost (partial save).
	DefaultHeartbeatTimeout = 10 * time.Second
	// DefaultControlPort is the product control port used by ShouldCaptureURL
	// exclusion (independent of ephemeral test bind addresses).
	DefaultControlPort       = "43759"
	StatusWaitingExtension   = "waiting_extension"
	StatusExtensionConnected = "extension_connected"
	StatusRecording          = "recording"
	StatusStopping           = "stopping"
	StatusSaved              = "saved"
	StatusFailed             = "failed"

	// MinBrowserTraceVersion is the floor for supports_browser_trace (with feature).
	MinBrowserTraceVersion = "1.2.0"
	FeatureBrowserTrace    = "browser-trace"
)

// ShouldCaptureURL reports whether a request URL should be stored in the
// capture buffer. Returns false for traffic to the product control hosts
// 127.0.0.1:43759 and localhost:43759 (any path, query, or fragment).
// All other well-formed http(s) URLs return true.
func ShouldCaptureURL(raw string) bool {
	if strings.TrimSpace(raw) == "" {
		return true
	}
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" {
		// Non-parseable or relative URLs are not treated as control traffic.
		return true
	}
	host := strings.ToLower(u.Hostname())
	port := u.Port()
	if port == "" {
		switch strings.ToLower(u.Scheme) {
		case "https":
			port = "443"
		default:
			port = "80"
		}
	}
	if port != DefaultControlPort {
		return true
	}
	if host == "127.0.0.1" || host == "localhost" {
		return false
	}
	return true
}

// IsCapturableTabURL reports whether a tab's URL is eligible for chrome.debugger
// attach. Returns false for empty/whitespace, chrome://, chrome-extension://,
// devtools://, and about:blank. Returns true for http:// and https://
// (including the product control page on 127.0.0.1:43759).
//
// Distinct from ShouldCaptureURL: control-host *request* traffic is excluded
// from the capture buffer, but the control tab may still be debugger-attached.
func IsCapturableTabURL(raw string) bool {
	u := strings.TrimSpace(raw)
	if u == "" {
		return false
	}
	// Prefer scheme prefix checks (stable even if url.Parse is lenient).
	lower := strings.ToLower(u)
	if strings.HasPrefix(lower, "chrome://") ||
		strings.HasPrefix(lower, "chrome-extension://") ||
		strings.HasPrefix(lower, "devtools://") ||
		lower == "about:blank" ||
		strings.HasPrefix(lower, "about:blank?") ||
		strings.HasPrefix(lower, "about:blank#") {
		return false
	}
	// http(s) only — including control page http://127.0.0.1:43759/go
	if strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://") {
		return true
	}
	return false
}

// ShouldAttemptAttach reports whether the extension agent should call
// chrome.debugger.attach for a tab right now.
// true iff recording && windowMatch && !alreadyAttached && IsCapturableTabURL(url).
//
// Used by tabs.onCreated and tabs.onUpdated: create-time empty/chrome:// URLs
// skip attach without permanent give-up; later navigation to http(s) retries.
func ShouldAttemptAttach(recording, windowMatch, alreadyAttached bool, rawURL string) bool {
	return recording && windowMatch && !alreadyAttached && IsCapturableTabURL(rawURL)
}

// Config controls a single browser-trace session.
type Config struct {
	Addr            string // host:port, product default 127.0.0.1:43759
	BaseDir         string // default ~/.tmp/browser-trace
	ReadyTimeout    time.Duration
	CompleteTimeout time.Duration
	// HeartbeatTimeout is how long the recording phase may go without
	// POST /v1/status or POST /v1/entries before heartbeat_lost.
	// Zero → DefaultHeartbeatTimeout (10s). Injectable for tests (e.g. 200ms).
	// Distinct from ReadyHeartbeat (ready-phase progress logging only).
	HeartbeatTimeout time.Duration
	NoOpenChrome     bool
	SessionSuffix    string // optional fixed suffix for dir naming
	Stdout           io.Writer
	Stderr           io.Writer

	// Logging / verbosity
	Verbose bool // extra detail: hello/version, start recording, stop, complete
	Quiet   bool // suppress info milestones; errors still on stderr. Quiet wins over Verbose.
	// NoLogFile, when true, does not write {sessionDir}/browser-trace.log.
	NoLogFile bool
	// ReadyHeartbeat is the ready-wait heartbeat interval. Zero → DefaultReadyHeartbeat (5s).
	ReadyHeartbeat time.Duration
}

// Result is the outcome of Run.
type Result struct {
	ExitCode   int
	SessionDir string
	Stdout     string
	Stderr     string
}

// lifecycleLogger writes progress to stderr (unless Quiet) and optionally
// mirrors info+ lines into {sessionDir}/browser-trace.log.
type lifecycleLogger struct {
	stderr  io.Writer
	quiet   bool
	verbose bool // effective: Verbose && !Quiet
	logFile *os.File
	logMu   sync.Mutex
}

func newLifecycleLogger(cfg Config, sessionDir string) *lifecycleLogger {
	// Quiet wins over Verbose.
	quiet := cfg.Quiet
	verbose := cfg.Verbose && !quiet
	l := &lifecycleLogger{
		stderr:  cfg.Stderr,
		quiet:   quiet,
		verbose: verbose,
	}
	// Log file mirrors info+ unless Quiet or NoLogFile.
	if !quiet && !cfg.NoLogFile && sessionDir != "" {
		path := filepath.Join(sessionDir, "browser-trace.log")
		f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
		if err == nil {
			l.logFile = f
		}
	}
	return l
}

func (l *lifecycleLogger) Close() {
	if l == nil || l.logFile == nil {
		return
	}
	l.logMu.Lock()
	_ = l.logFile.Close()
	l.logFile = nil
	l.logMu.Unlock()
}

func (l *lifecycleLogger) emit(line string) {
	if l == nil {
		return
	}
	if !strings.HasSuffix(line, "\n") {
		line += "\n"
	}
	if l.stderr != nil {
		_, _ = io.WriteString(l.stderr, line)
	}
	l.logMu.Lock()
	if l.logFile != nil {
		_, _ = io.WriteString(l.logFile, line)
	}
	l.logMu.Unlock()
}

// Info emits a lifecycle milestone (suppressed when Quiet).
func (l *lifecycleLogger) Info(format string, args ...any) {
	if l == nil || l.quiet {
		return
	}
	msg := fmt.Sprintf(format, args...)
	l.emit("browser-trace: " + msg)
}

// Verbose emits extra detail when Verbose is on and Quiet is off.
func (l *lifecycleLogger) Verbose(format string, args ...any) {
	if l == nil || !l.verbose {
		return
	}
	msg := fmt.Sprintf(format, args...)
	l.emit("browser-trace: " + msg)
}

// Error always prints to stderr; mirrors to log file when open (not Quiet).
func (l *lifecycleLogger) Error(format string, args ...any) {
	if l == nil {
		return
	}
	msg := fmt.Sprintf(format, args...)
	// Errors do not use the info prefix pattern required for Quiet bans;
	// they always go to stderr.
	line := msg
	if !strings.HasSuffix(line, "\n") {
		line += "\n"
	}
	if l.stderr != nil {
		_, _ = io.WriteString(l.stderr, line)
	}
	// Mirror errors into log file only when logging is enabled (not Quiet).
	if !l.quiet {
		l.logMu.Lock()
		if l.logFile != nil {
			_, _ = io.WriteString(l.logFile, line)
		}
		l.logMu.Unlock()
	}
}

// Run binds cfg.Addr, runs one control-server session, and returns when the
// session is saved, failed, or a deadline expires.
//
// Context cancel is treated as CLI SIGINT/SIGTERM: queue stop and wait up to
// CompleteTimeout for POST /v1/complete.
func Run(ctx context.Context, cfg Config) (*Result, error) {
	cfg = applyDefaults(cfg)

	suffix := cfg.SessionSuffix
	if suffix == "" {
		suffix = randomSuffix(6)
	}
	sessionID := suffix
	startedAt := time.Now()
	dirName := startedAt.Format("2006-01-02-15-04-05") + "-" + suffix
	sessionDir := filepath.Join(cfg.BaseDir, dirName)

	if err := os.MkdirAll(sessionDir, 0o755); err != nil {
		msg := fmt.Sprintf("create session dir: %v", err)
		fmt.Fprintln(cfg.Stderr, msg)
		return &Result{ExitCode: 1, SessionDir: sessionDir}, fmt.Errorf("%s", msg)
	}

	log := newLifecycleLogger(cfg, sessionDir)
	defer log.Close()

	// Extract embedded extension early so session JSON/HTML can expose the path.
	extPath, extVersion, extErr := ExtractEmbeddedExtension(cfg.BaseDir)
	if extErr != nil {
		// Best-effort: session can still run; install help may be incomplete.
		// Always surface extract failures as warnings (not Quiet-suppressed info).
		fmt.Fprintf(cfg.Stderr, "warning: extract embedded extension: %v\n", extErr)
	}

	ln, err := net.Listen("tcp", cfg.Addr)
	if err != nil {
		msg := fmt.Sprintf("listen %s failed: address already in use or cannot bind: %v", cfg.Addr, err)
		log.Error("%s", msg)
		return &Result{ExitCode: 1, SessionDir: sessionDir}, fmt.Errorf("%s", msg)
	}

	log.Info("listening on %s", cfg.Addr)

	sess := newSession(sessionID, sessionDir, startedAt, cfg.ReadyTimeout)
	sess.extensionInstallPath = extPath
	sess.embeddedVersion = extVersion
	srv := newControlServer(sess)

	httpServer := &http.Server{Handler: srv.handler()}
	serveErrCh := make(chan error, 1)
	go func() {
		err := httpServer.Serve(ln)
		if err != nil && err != http.ErrServerClosed {
			serveErrCh <- err
		}
		close(serveErrCh)
	}()
	defer func() {
		shCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = httpServer.Shutdown(shCtx)
		_ = ln.Close()
	}()

	baseURL := "http://" + cfg.Addr
	sessionURL := baseURL + "/go?session=" + sessionID

	log.Info("session id=%s url=%s", sessionID, sessionURL)

	if !cfg.NoOpenChrome {
		if err := openChromeNewWindow(sessionURL, extPath); err != nil {
			fmt.Fprintf(cfg.Stderr, "warning: open chrome: %v\n", err)
		}
	}

	log.Info("ready wait: waiting for recording (timeout %s)", cfg.ReadyTimeout)

	// --- Ready phase: hello + status recording ---
	readyDeadline := time.After(cfg.ReadyTimeout)
	readyEnd := time.Now().Add(cfg.ReadyTimeout)
	heartbeat := cfg.ReadyHeartbeat
	hbTicker := time.NewTicker(heartbeat)
	defer hbTicker.Stop()

	loggedHello := false
	maybeLogHello := func() {
		if loggedHello || !sess.helloReceived() {
			return
		}
		loggedHello = true
		ver, feats := sess.helloInfo()
		log.Verbose("hello received version=%s features=%v", ver, feats)
		log.Verbose("start recording command queued")
	}

	var ready bool
	for !ready {
		maybeLogHello()
		select {
		case <-sess.recordingCh:
			ready = true
		case <-hbTicker.C:
			left := time.Until(readyEnd)
			if left < 0 {
				left = 0
			}
			stage := "no_hello"
			if sess.helloReceived() {
				stage = "no_recording"
			}
			// Heartbeat while still waiting (tokens: waiting, left, stage).
			log.Info("still waiting for ready: %ds left, stage=%s",
				int(left.Round(time.Second)/time.Second), stage)
		case <-readyDeadline:
			sess.setStatus(StatusFailed)
			msg := readyFailMessage(sess, sessionURL)
			log.Error("%s", msg)
			_ = writeMeta(sess, "timeout", nil)
			return &Result{ExitCode: 1, SessionDir: sessionDir}, fmt.Errorf("%s", msg)
		case <-ctx.Done():
			sess.setStatus(StatusFailed)
			msg := "session cancelled before ready (extension not recording)"
			log.Error("%s", msg)
			_ = writeMeta(sess, "cancelled", nil)
			return &Result{ExitCode: 1, SessionDir: sessionDir}, fmt.Errorf("%s", msg)
		case err := <-serveErrCh:
			if err != nil {
				msg := fmt.Sprintf("server error: %v", err)
				log.Error("%s", msg)
				return &Result{ExitCode: 1, SessionDir: sessionDir}, fmt.Errorf("%s", msg)
			}
		}
	}
	// Catch hello that arrived in the same wait as recording.
	maybeLogHello()

	log.Info("recording started")
	// Seed heartbeat clock if the first recording status already set it;
	// otherwise start the clock now so HeartbeatTimeout is well-defined.
	sess.ensureHeartbeatSeed()

	// --- Recording: wait for complete, heartbeat_lost, or CLI stop ---
	// Poll interval: small enough for short injectable timeouts (tests use ~200ms).
	// Ready-phase hbTicker continues until Run returns (harmless after ready).
	recHbPoll := cfg.HeartbeatTimeout / 4
	if recHbPoll < 25*time.Millisecond {
		recHbPoll = 25 * time.Millisecond
	}
	if recHbPoll > time.Second {
		recHbPoll = time.Second
	}
	recHbTicker := time.NewTicker(recHbPoll)
	defer recHbTicker.Stop()

	for {
		select {
		case payload := <-sess.completeCh:
			log.Verbose("complete received (stop_reason=%s)", payload.StopReason)
			return finishSuccess(cfg, sess, log, payload)
		case <-recHbTicker.C:
			if sess.heartbeatStale(cfg.HeartbeatTimeout) {
				return finishHeartbeatLost(cfg, sess, log)
			}
		case <-ctx.Done():
			// CLI SIGINT/SIGTERM: queue stop and wait for complete.
			sess.setStatus(StatusStopping)
			log.Verbose("stop requested (CLI signal)")
			sess.queueCommand("stop")
			return waitComplete(cfg, sess, log, cfg.CompleteTimeout)
		case err := <-serveErrCh:
			if err != nil {
				msg := fmt.Sprintf("server error: %v", err)
				log.Error("%s", msg)
				return &Result{ExitCode: 1, SessionDir: sessionDir}, fmt.Errorf("%s", msg)
			}
			// Server closed unexpectedly while recording.
			msg := "server closed before complete"
			log.Error("%s", msg)
			return &Result{ExitCode: 1, SessionDir: sessionDir}, fmt.Errorf("%s", msg)
		}
	}
}

func waitComplete(cfg Config, sess *session, log *lifecycleLogger, timeout time.Duration) (*Result, error) {
	select {
	case payload := <-sess.completeCh:
		log.Verbose("complete received (stop_reason=%s)", payload.StopReason)
		return finishSuccess(cfg, sess, log, payload)
	case <-time.After(timeout):
		sess.setStatus(StatusFailed)
		msg := fmt.Sprintf("complete timeout: final HAR not received within %s after stop", timeout)
		log.Error("%s", msg)
		_ = writeMeta(sess, "timeout", map[string]any{
			"error": "complete_timeout",
		})
		// Do not write a corrupt recording.har.
		return &Result{ExitCode: 1, SessionDir: sess.outDir}, fmt.Errorf("%s", msg)
	}
}

func finishSuccess(cfg Config, sess *session, log *lifecycleLogger, payload completePayload) (*Result, error) {
	if err := saveArtifacts(sess, payload); err != nil {
		sess.setStatus(StatusFailed)
		msg := fmt.Sprintf("save artifacts: %v", err)
		log.Error("%s", msg)
		return &Result{ExitCode: 1, SessionDir: sess.outDir}, fmt.Errorf("%s", msg)
	}
	sess.setStatus(StatusSaved)
	log.Info("saved recording to session dir")
	// User-facing stdout: session path with trailing newline only (no banners).
	line := sess.outDir + "\n"
	if _, err := io.WriteString(cfg.Stdout, line); err != nil {
		return &Result{ExitCode: 1, SessionDir: sess.outDir, Stdout: line}, err
	}
	return &Result{
		ExitCode:   0,
		SessionDir: sess.outDir,
		Stdout:     line,
	}, nil
}

// finishHeartbeatLost saves a partial HAR from the last /v1/entries snapshot
// (empty array OK), writes meta with stop_reason=heartbeat_lost + partial=true,
// emits a stderr warning, and exits 0 with session path on stdout.
func finishHeartbeatLost(cfg Config, sess *session, log *lifecycleLogger) (*Result, error) {
	if err := savePartialFromPreview(sess, "heartbeat_lost"); err != nil {
		sess.setStatus(StatusFailed)
		msg := fmt.Sprintf("save heartbeat_lost artifacts: %v", err)
		log.Error("%s", msg)
		return &Result{ExitCode: 1, SessionDir: sess.outDir}, fmt.Errorf("%s", msg)
	}
	sess.setStatus(StatusSaved)
	// Always surface on stderr (not Quiet-suppressed): tests and users need
	// warning / heartbeat / unusual|acceptable tokens.
	warn := "warning: heartbeat lost — browser closed or extension stopped without Stop; partial snapshot saved (unusual but acceptable)\n"
	if cfg.Stderr != nil {
		_, _ = io.WriteString(cfg.Stderr, warn)
	}
	// Also mirror via lifecycle logger when not Quiet (log file + stderr may duplicate; OK).
	log.Info("heartbeat_lost: partial snapshot saved (unusual but acceptable)")

	line := sess.outDir + "\n"
	if _, err := io.WriteString(cfg.Stdout, line); err != nil {
		return &Result{ExitCode: 1, SessionDir: sess.outDir, Stdout: line, Stderr: warn}, err
	}
	return &Result{
		ExitCode:   0,
		SessionDir: sess.outDir,
		Stdout:     line,
		Stderr:     warn,
	}, nil
}

// savePartialFromPreview builds recording.har from the last previewEntries
// snapshot and writes meta.json with partial=true and the given stop_reason.
func savePartialFromPreview(sess *session, stopReason string) error {
	sess.mu.Lock()
	entries := sess.previewEntries
	if entries == nil {
		entries = []json.RawMessage{}
	}
	copied := make([]json.RawMessage, len(entries))
	for i, e := range entries {
		if e == nil {
			copied[i] = json.RawMessage("null")
			continue
		}
		copied[i] = append(json.RawMessage(nil), e...)
	}
	entryCount := len(copied)
	windowID := sess.windowID
	sess.entryCount = entryCount
	sess.stopReason = stopReason
	sess.mu.Unlock()

	entryList := make([]any, 0, len(copied))
	for _, raw := range copied {
		var v any
		if err := json.Unmarshal(raw, &v); err != nil {
			// Keep a placeholder so entry_count still matches wire length.
			entryList = append(entryList, map[string]any{})
			continue
		}
		entryList = append(entryList, v)
	}
	doc := map[string]any{
		"log": map[string]any{
			"version": "1.2",
			"creator": map[string]any{"name": "browser-trace", "version": "1.0"},
			"entries": entryList,
		},
	}
	harBytes, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return err
	}
	harPath := filepath.Join(sess.outDir, "recording.har")
	tmpPath := harPath + ".tmp"
	if err := os.WriteFile(tmpPath, harBytes, 0o644); err != nil {
		return err
	}
	if err := os.Rename(tmpPath, harPath); err != nil {
		return err
	}

	return writeMeta(sess, stopReason, map[string]any{
		"entry_count": entryCount,
		"window_id":   windowID,
		"status":      StatusSaved,
		"partial":     true,
	})
}

func readyFailMessage(sess *session, sessionURL string) string {
	installHint := "Install/enable Chrome-Ext-Capture-API and grant host_permissions for http://127.0.0.1:43759/*"
	sess.mu.Lock()
	installPath := sess.extensionInstallPath
	sess.mu.Unlock()
	if installPath != "" {
		installHint = fmt.Sprintf(
			"Install/enable the extension: open chrome://extensions, enable Developer mode, Load unpacked from %s",
			installPath,
		)
	}
	if sessionURL == "" {
		sessionURL = "/go?session=" + sess.id
	}
	if !sess.helloReceived() {
		return fmt.Sprintf(
			"ready timeout: extension did not connect (POST /v1/hello) [stage=no_hello]. Session URL: %s. %s",
			sessionURL, installHint,
		)
	}
	return fmt.Sprintf(
		"ready timeout: extension connected but did not start recording within deadline [stage=no_recording]. Session URL: %s. %s",
		sessionURL, installHint,
	)
}

func applyDefaults(cfg Config) Config {
	if cfg.Addr == "" {
		cfg.Addr = DefaultAddr
	}
	if cfg.BaseDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			home = os.TempDir()
		}
		cfg.BaseDir = filepath.Join(home, ".tmp", "browser-trace")
	}
	if cfg.ReadyTimeout <= 0 {
		cfg.ReadyTimeout = DefaultReadyTimeout
	}
	if cfg.CompleteTimeout <= 0 {
		cfg.CompleteTimeout = DefaultCompleteTimeout
	}
	if cfg.ReadyHeartbeat <= 0 {
		cfg.ReadyHeartbeat = DefaultReadyHeartbeat
	}
	if cfg.HeartbeatTimeout <= 0 {
		cfg.HeartbeatTimeout = DefaultHeartbeatTimeout
	}
	if cfg.Stdout == nil {
		cfg.Stdout = os.Stdout
	}
	if cfg.Stderr == nil {
		cfg.Stderr = os.Stderr
	}
	return cfg
}

func randomSuffix(nBytes int) string {
	b := make([]byte, nBytes)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano()%1e9)
	}
	return hex.EncodeToString(b)
}

// --- session ---

type session struct {
	mu             sync.Mutex
	id             string
	outDir         string
	startedAt      time.Time
	readyTimeout   time.Duration
	status         string
	helloOK        bool
	startQueued    bool
	windowID       int
	entryCount     int
	stopReason     string
	errors         []string
	extensionVer   string
	extensionFeats []string
	supportsBT     bool

	// Install help from embedded extract (set before server serves traffic).
	extensionInstallPath string
	embeddedVersion      string

	cmdMu   sync.Mutex
	cmdCond *sync.Cond
	cmds    []command

	recordingOnce sync.Once
	recordingCh   chan struct{}
	completeOnce  sync.Once
	completeCh    chan completePayload

	// Live preview snapshot (last POST /v1/entries). Replace, not merge.
	previewEntries   []json.RawMessage
	previewUpdatedAt time.Time
	previewCount     int

	// lastHeartbeatAt is refreshed on POST /v1/status and POST /v1/entries
	// while recording. Zero until first recording-phase heartbeat.
	lastHeartbeatAt time.Time
}

type command struct {
	Type      string         `json:"type"`
	SessionID string         `json:"session_id,omitempty"`
	Options   map[string]any `json:"options,omitempty"`
}

type completePayload struct {
	HAR        json.RawMessage
	StopReason string
	WindowID   int
	Stats      map[string]any
	Raw        map[string]any
}

func newSession(id, outDir string, startedAt time.Time, readyTimeout time.Duration) *session {
	if readyTimeout <= 0 {
		readyTimeout = DefaultReadyTimeout
	}
	s := &session{
		id:           id,
		outDir:       outDir,
		startedAt:    startedAt,
		readyTimeout: readyTimeout,
		status:       StatusWaitingExtension,
		recordingCh:  make(chan struct{}),
		completeCh:   make(chan completePayload, 1),
	}
	s.cmdCond = sync.NewCond(&s.cmdMu)
	return s
}

func (s *session) setStatus(st string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.status = st
}

func (s *session) getStatus() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.status
}

func (s *session) helloReceived() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.helloOK
}

// helloInfo returns extension version and features after hello (may be empty).
func (s *session) helloInfo() (version string, features []string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	version = s.extensionVer
	if len(s.extensionFeats) > 0 {
		features = append([]string(nil), s.extensionFeats...)
	}
	return version, features
}

func (s *session) markHello(version string, features []string) {
	if features == nil {
		features = []string{}
	}
	// Copy so callers cannot mutate session state later.
	feats := append([]string(nil), features...)
	supports := computeSupportsBrowserTrace(version, feats)
	s.mu.Lock()
	s.helloOK = true
	s.extensionVer = version
	s.extensionFeats = feats
	s.supportsBT = supports
	// Intermediate UI phase: connected but not yet recording.
	if s.status == StatusWaitingExtension {
		s.status = StatusExtensionConnected
	}
	s.mu.Unlock()
	s.ensureStartQueued()
}

// computeSupportsBrowserTrace requires feature "browser-trace" and version ≥ 1.2.0.
// Version alone (no features) is not enough.
func computeSupportsBrowserTrace(version string, features []string) bool {
	hasFeature := false
	for _, f := range features {
		if f == FeatureBrowserTrace {
			hasFeature = true
			break
		}
	}
	if !hasFeature {
		return false
	}
	return versionGTE(version, MinBrowserTraceVersion)
}

// versionGTE compares simple major.minor.patch (extra suffixes ignored after numeric parts).
func versionGTE(v, min string) bool {
	vp := parseSemver(v)
	mp := parseSemver(min)
	for i := 0; i < 3; i++ {
		if vp[i] > mp[i] {
			return true
		}
		if vp[i] < mp[i] {
			return false
		}
	}
	return true
}

func parseSemver(s string) [3]int {
	var out [3]int
	s = strings.TrimSpace(s)
	if s == "" {
		return out
	}
	// Strip pre-release / build metadata.
	if i := strings.IndexAny(s, "-+"); i >= 0 {
		s = s[:i]
	}
	parts := strings.Split(s, ".")
	for i := 0; i < 3 && i < len(parts); i++ {
		n := 0
		for _, c := range parts[i] {
			if c < '0' || c > '9' {
				break
			}
			n = n*10 + int(c-'0')
		}
		out[i] = n
	}
	return out
}

func (s *session) ensureStartQueued() {
	s.mu.Lock()
	if s.startQueued {
		s.mu.Unlock()
		return
	}
	s.startQueued = true
	s.mu.Unlock()
	s.queueCommand("start")
}

func (s *session) queueCommand(typ string) {
	s.cmdMu.Lock()
	s.cmds = append(s.cmds, command{
		Type:      typ,
		SessionID: s.id,
		Options: map[string]any{
			"window_strategy": "pin_session_window",
			"capture":         "all_tabs_in_window",
		},
	})
	s.cmdCond.Broadcast()
	s.cmdMu.Unlock()
}

// waitCommand long-polls for the next command up to wait duration.
func (s *session) waitCommand(wait time.Duration) command {
	deadline := time.Now().Add(wait)
	s.cmdMu.Lock()
	defer s.cmdMu.Unlock()
	for {
		if len(s.cmds) > 0 {
			cmd := s.cmds[0]
			s.cmds = s.cmds[1:]
			return cmd
		}
		remaining := time.Until(deadline)
		if remaining <= 0 {
			return command{Type: ""} // noop
		}
		// Wait with timeout via timed Broadcast helper.
		timer := time.AfterFunc(remaining, func() {
			s.cmdMu.Lock()
			s.cmdCond.Broadcast()
			s.cmdMu.Unlock()
		})
		s.cmdCond.Wait()
		timer.Stop()
		if len(s.cmds) == 0 && time.Now().After(deadline) {
			return command{Type: ""}
		}
	}
}

func (s *session) applyStatus(state string, entryCount, windowID int) {
	s.mu.Lock()
	if windowID != 0 {
		s.windowID = windowID
	}
	s.entryCount = entryCount
	if state == StatusRecording || state == "recording" {
		// Promote to recording from waiting / connected intermediate phases.
		if s.status == StatusWaitingExtension ||
			s.status == StatusExtensionConnected ||
			s.status == StatusRecording {
			s.status = StatusRecording
		}
		// Recording-phase liveness: every successful status while recording
		// refreshes the heartbeat clock (including the first recording status).
		if s.status == StatusRecording {
			s.lastHeartbeatAt = time.Now()
		}
		s.mu.Unlock()
		s.recordingOnce.Do(func() { close(s.recordingCh) })
		return
	}
	s.mu.Unlock()
}

// touchHeartbeat refreshes lastHeartbeatAt (caller must hold s.mu when needed).
func (s *session) touchHeartbeat() {
	s.mu.Lock()
	s.lastHeartbeatAt = time.Now()
	s.mu.Unlock()
}

// ensureHeartbeatSeed sets lastHeartbeatAt if still zero (e.g. ready just fired).
func (s *session) ensureHeartbeatSeed() {
	s.mu.Lock()
	if s.lastHeartbeatAt.IsZero() {
		s.lastHeartbeatAt = time.Now()
	}
	s.mu.Unlock()
}

// heartbeatStale reports whether recording has been silent longer than timeout.
func (s *session) heartbeatStale(timeout time.Duration) bool {
	if timeout <= 0 {
		return false
	}
	s.mu.Lock()
	last := s.lastHeartbeatAt
	st := s.status
	s.mu.Unlock()
	if st != StatusRecording || last.IsZero() {
		return false
	}
	return time.Since(last) > timeout
}

func (s *session) deliverComplete(p completePayload) {
	s.completeOnce.Do(func() {
		s.completeCh <- p
		close(s.completeCh)
	})
}

// --- HTTP control server ---

type controlServer struct {
	sess *session
}

func newControlServer(sess *session) *controlServer {
	return &controlServer{sess: sess}
}

func (c *controlServer) handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/health", c.handleHealth)
	mux.HandleFunc("/v1/session", c.handleSession)
	mux.HandleFunc("/go", c.handleGo)
	mux.HandleFunc("/v1/hello", c.handleHello)
	mux.HandleFunc("/v1/commands", c.handleCommands)
	mux.HandleFunc("/v1/status", c.handleStatus)
	mux.HandleFunc("/v1/complete", c.handleComplete)
	mux.HandleFunc("/v1/entries", c.handleEntries)
	mux.HandleFunc("/preview", c.handlePreview)
	return mux
}

func (c *controlServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{"ok":true}`))
}

// ShouldExpandInstallPanel reports whether the session-page install panel
// should be open. Expand unless the extension is connected and supports
// browser-trace: !(connected && supports).
func ShouldExpandInstallPanel(connected, supports bool) bool {
	return !(connected && supports)
}

func (c *controlServer) handleGo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	session := r.URL.Query().Get("session")
	esc := htmlEscape(session)

	c.sess.mu.Lock()
	installPath := c.sess.extensionInstallPath
	embeddedVer := c.sess.embeddedVersion
	helloOK := c.sess.helloOK
	supportsBT := c.sess.supportsBT
	c.sess.mu.Unlock()

	// Always render the install panel (update/reload remains available even when
	// the extension is healthy). Expand when not fully working.
	// Note: when collapsed, omit data-default-open entirely — sealed tests match
	// \bopen\b on the details tag, which also hits the substring in
	// data-default-open="false". Expanded uses open + data-default-open="true".
	pathEsc := htmlEscape(installPath)
	verEsc := htmlEscape(embeddedVer)
	expand := ShouldExpandInstallPanel(helloOK, supportsBT)
	expandAttrs := ""
	if expand {
		expandAttrs = ` open data-default-open="true"`
	}
	installPanel := fmt.Sprintf(`
  <details id="browser-trace-install" class="browser-trace-install" data-browser-trace-install data-install-rerun-guidance data-extension-path="%s" data-install-path="%s" data-embedded-version="%s"%s>
    <summary>Install / update API Capture extension</summary>
    <div class="install-body">
      <p>Load or reload the unpacked package so browser-trace can record this window. Expand this panel anytime to update the extension.</p>
      <ol>
        <li>Open <strong>chrome://extensions</strong> (copy that address into the Chrome address bar)</li>
        <li>Enable <strong>Developer mode</strong> (top-right toggle)</li>
        <li>Click <strong>Load unpacked</strong> (or Reload if already loaded)</li>
        <li>Select this folder:
          <pre class="path">%s</pre>
        </li>
        <li data-install-rerun-guidance class="install-rerun-guidance">
          After install, <strong>Load unpacked</strong>, or <strong>Reload</strong>:
          <strong>close this Chrome window</strong> and <strong>re-run browser-trace</strong>
          so capture starts with the updated extension. Mid-session hot-reload is not supported.
        </li>
      </ol>
      <p class="muted">Embedded version: <code>%s</code>.</p>
    </div>
  </details>`, pathEsc, pathEsc, verEsc, expandAttrs, pathEsc, verEsc)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = fmt.Fprintf(w, `<!DOCTYPE html>
<html><head><meta charset="utf-8"><title>browser-trace session</title>
<style>
  body { font-family: system-ui, sans-serif; margin: 1.5rem; max-width: 40rem; }
  code, pre.path { background: #f4f4f4; padding: 0.1em 0.3em; }
  pre.path { padding: 0.6rem 0.8rem; overflow-x: auto; white-space: pre-wrap; word-break: break-all; }
  #browser-trace-status { border: 1px solid #ccc; border-radius: 8px; padding: 1rem; margin-top: 1rem; }
  #browser-trace-install { border: 1px solid #c9a227; background: #fffbeb; border-radius: 8px; padding: 0.75rem 1rem; margin-top: 1rem; }
  #browser-trace-install summary { cursor: pointer; font-weight: 600; }
  #browser-trace-install .install-body { margin-top: 0.75rem; }
  .install-rerun-guidance { margin-top: 0.35rem; color: #5c4a00; }
  .row { margin: 0.35rem 0; }
  .hint { color: #444; margin-top: 0.75rem; }
  .muted { color: #888; font-size: 0.9rem; }
</style>
</head>
<body>
  <h1>browser-trace</h1>
  <p>Session <code id="session-id">%s</code> is active.</p>
  <p class="muted">Keep this window open. The API Capture extension records all tabs in this window.</p>
  <div id="browser-trace-status" data-browser-trace-status>
    <div class="row"><strong>Phase:</strong> <span id="st-phase">…</span></div>
    <div class="row"><strong>Extension:</strong> <span id="st-ext">…</span></div>
    <div class="row"><strong>Recording:</strong> <span id="st-rec">…</span></div>
    <div class="row"><strong>Ready:</strong> <span id="st-ready">…</span></div>
    <div class="hint" id="st-hint">Loading status…</div>
  </div>
%s
  <script>
(function() {
  var sessionId = %q;
  function $(id) { return document.getElementById(id); }
  function shouldExpandInstallPanel(connected, supports) {
    return !(connected && supports);
  }
  var panel = document.getElementById('browser-trace-install');
  var syncingOpen = false;
  if (panel) {
    // Prefer user gesture on summary so programmatic open sync does not freeze.
    var summary = panel.querySelector('summary');
    if (summary) {
      summary.addEventListener('click', function() {
        panel.setAttribute('data-user-toggled', 'true');
      });
    }
    panel.addEventListener('toggle', function() {
      if (!syncingOpen) {
        panel.setAttribute('data-user-toggled', 'true');
      }
    });
  }
  function render(data) {
    if (!data) return;
    $('st-phase').textContent = data.phase || '';
    var ext = data.extension || {};
    var extBits = [];
    extBits.push(ext.connected ? 'connected' : 'not connected');
    if (ext.version) extBits.push('v' + ext.version);
    extBits.push(ext.supports_browser_trace ? 'supports browser-trace' : 'no browser-trace support');
    $('st-ext').textContent = extBits.join(' · ');
    var rec = data.recording || {};
    $('st-rec').textContent = (rec.active ? 'active' : 'inactive') +
      ' · entries ' + (rec.entry_count || 0);
    var ready = data.ready || {};
    var rem = typeof ready.remaining_ms === 'number' ? ready.remaining_ms : 0;
    $('st-ready').textContent = Math.max(0, Math.ceil(rem / 1000)) + 's remaining';
    $('st-hint').textContent = data.hint || '';
    var p = document.getElementById('browser-trace-install');
    if (p && p.getAttribute('data-user-toggled') !== 'true') {
      // Never display:none — keep summary visible; sync open from expand policy.
      var wantOpen = shouldExpandInstallPanel(!!ext.connected, !!ext.supports_browser_trace);
      if (p.open !== wantOpen) {
        syncingOpen = true;
        p.open = wantOpen;
        syncingOpen = false;
      }
    }
  }
  function poll() {
    fetch('/v1/session?session=' + encodeURIComponent(sessionId))
      .then(function(r) { return r.json().then(function(j) { return { ok: r.ok, body: j }; }); })
      .then(function(res) {
        if (res.ok) render(res.body);
        else $('st-hint').textContent = (res.body && (res.body.error || res.body.message)) || 'session error';
      })
      .catch(function() { $('st-hint').textContent = 'Waiting for control server…'; });
  }
  poll();
  setInterval(poll, 500);
})();
  </script>
</body></html>`, esc, installPanel, session)
}

func htmlEscape(s string) string {
	r := strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&quot;")
	return r.Replace(s)
}

func (c *controlServer) handleSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	id := r.URL.Query().Get("session")
	if id == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error":   "missing session",
			"message": "session query parameter is required",
		})
		return
	}
	if id != c.sess.id {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error":   "session not found",
			"message": "unknown session id",
		})
		return
	}
	snap := c.sess.snapshotUI()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(snap)
}

// sessionUISnapshot is the GET /v1/session JSON shape for the page poller.
type sessionUISnapshot struct {
	SessionID            string `json:"session_id"`
	Phase                string `json:"phase"`
	ExtensionInstallPath string `json:"extension_install_path,omitempty"`
	EmbeddedVersion      string `json:"embedded_version,omitempty"`
	Extension            struct {
		Connected            bool     `json:"connected"`
		Version              string   `json:"version"`
		Features             []string `json:"features"`
		SupportsBrowserTrace bool     `json:"supports_browser_trace"`
	} `json:"extension"`
	Recording struct {
		Active     bool `json:"active"`
		EntryCount int  `json:"entry_count"`
		WindowID   int  `json:"window_id"`
	} `json:"recording"`
	Ready struct {
		DeadlineMS  int64 `json:"deadline_ms"`
		ElapsedMS   int64 `json:"elapsed_ms"`
		RemainingMS int64 `json:"remaining_ms"`
	} `json:"ready"`
	Hint string `json:"hint"`
}

func (s *session) snapshotUI() sessionUISnapshot {
	s.mu.Lock()
	defer s.mu.Unlock()

	var snap sessionUISnapshot
	snap.SessionID = s.id
	snap.Phase = s.status
	snap.ExtensionInstallPath = s.extensionInstallPath
	snap.EmbeddedVersion = s.embeddedVersion
	snap.Extension.Connected = s.helloOK
	snap.Extension.Version = s.extensionVer
	if s.extensionFeats == nil {
		snap.Extension.Features = []string{}
	} else {
		snap.Extension.Features = append([]string(nil), s.extensionFeats...)
	}
	snap.Extension.SupportsBrowserTrace = s.supportsBT
	snap.Recording.Active = s.status == StatusRecording
	snap.Recording.EntryCount = s.entryCount
	snap.Recording.WindowID = s.windowID

	deadlineMS := s.readyTimeout.Milliseconds()
	if deadlineMS <= 0 {
		deadlineMS = DefaultReadyTimeout.Milliseconds()
	}
	elapsedMS := time.Since(s.startedAt).Milliseconds()
	if elapsedMS < 0 {
		elapsedMS = 0
	}
	remainingMS := deadlineMS - elapsedMS
	if remainingMS < 0 {
		remainingMS = 0
	}
	snap.Ready.DeadlineMS = deadlineMS
	snap.Ready.ElapsedMS = elapsedMS
	snap.Ready.RemainingMS = remainingMS
	snap.Hint = buildSessionHint(s.helloOK, s.supportsBT, s.status, s.extensionVer, s.extensionInstallPath)
	return snap
}

func buildSessionHint(helloOK, supports bool, status, version, installPath string) string {
	if !helloOK {
		// Install-oriented guidance (session page is primary when disconnected).
		base := "Waiting for API Capture extension… Install or enable Chrome-Ext-Capture-API if it is not loaded."
		if installPath != "" {
			return base + " Load unpacked from chrome://extensions (enable Developer mode) using folder: " + installPath
		}
		return base + " Open chrome://extensions, enable Developer mode, then Load unpacked the extension package."
	}
	if !supports {
		msg := "Extension connected but does not support browser-trace. Update the extension to version " +
			MinBrowserTraceVersion + "+ and ensure it advertises the browser-trace feature/capability."
		if installPath != "" {
			msg += " You can Load unpacked the bundled package from: " + installPath
		}
		return msg
	}
	switch status {
	case StatusRecording:
		return "Recording network traffic in this window."
	case StatusStopping:
		return "Stopping capture…"
	case StatusSaved:
		return "Capture saved."
	case StatusFailed:
		return "Session failed."
	default:
		// extension_connected or similar — operational, not install tutorial.
		return "Extension connected and supports browser-trace. Browse in this window to capture APIs. Ready to record."
	}
}

func (c *controlServer) handleHello(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		Version  string   `json:"version"`
		Features []string `json:"features"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	// features omitted → nil slice → supports_browser_trace false (version alone not enough)
	c.sess.markHello(body.Version, body.Features)
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{"ok":true}`))
}

func (c *controlServer) handleCommands(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	waitSec := 0
	if v := r.URL.Query().Get("wait"); v != "" {
		var n float64
		if _, err := fmt.Sscanf(v, "%f", &n); err == nil && n > 0 {
			waitSec = int(n)
		}
	}
	if waitSec <= 0 {
		waitSec = 0
	}
	if waitSec > 30 {
		waitSec = 30
	}
	// Honor request context cancel so server shutdown unblocks pollers.
	type result struct{ cmd command }
	ch := make(chan result, 1)
	go func() {
		ch <- result{cmd: c.sess.waitCommand(time.Duration(waitSec) * time.Second)}
	}()
	select {
	case <-r.Context().Done():
		// Client cancelled long-poll; return noop.
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(command{Type: ""})
		return
	case res := <-ch:
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(res.cmd)
	}
}

func (c *controlServer) handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		State      string `json:"state"`
		EntryCount int    `json:"entry_count"`
		WindowID   int    `json:"window_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}
	c.sess.applyStatus(body.State, body.EntryCount, body.WindowID)
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{"ok":true}`))
}

func (c *controlServer) handleComplete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var raw map[string]json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}
	payload := completePayload{Raw: map[string]any{}}
	if har, ok := raw["har"]; ok {
		payload.HAR = har
	}
	if sr, ok := raw["stop_reason"]; ok {
		var s string
		if json.Unmarshal(sr, &s) == nil {
			payload.StopReason = s
		}
	}
	if wid, ok := raw["window_id"]; ok {
		var n int
		if json.Unmarshal(wid, &n) == nil {
			payload.WindowID = n
		}
	}
	if st, ok := raw["stats"]; ok {
		var m map[string]any
		if json.Unmarshal(st, &m) == nil {
			payload.Stats = m
		}
	}
	// If status was recording and complete arrives without prior stop, treat as extension stop.
	if st := c.sess.getStatus(); st == StatusRecording {
		c.sess.setStatus(StatusStopping)
	}
	c.sess.deliverComplete(payload)
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{"ok":true}`))
}

// writeSessionNotFoundJSON writes a 404 JSON body for unknown sessions.
func writeSessionNotFoundJSON(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"error":   "session not found",
		"message": "unknown session id",
	})
}

// handleEntries serves POST (replace snapshot) and GET (read snapshot) for live preview.
func (c *controlServer) handleEntries(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		c.handleEntriesPOST(w, r)
	case http.MethodGet:
		c.handleEntriesGET(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (c *controlServer) handleEntriesPOST(w http.ResponseWriter, r *http.Request) {
	var body struct {
		SessionID string            `json:"session_id"`
		Entries   []json.RawMessage `json:"entries"`
		Count     int               `json:"count"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error":   "bad json",
			"message": err.Error(),
		})
		return
	}
	if body.SessionID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error":   "missing session_id",
			"message": "session_id is required",
		})
		return
	}
	if body.SessionID != c.sess.id {
		writeSessionNotFoundJSON(w)
		return
	}
	if body.Entries == nil {
		body.Entries = []json.RawMessage{}
	}
	count := body.Count
	if count == 0 && len(body.Entries) > 0 {
		count = len(body.Entries)
	}
	// Empty entries always mean clear / zero snapshot.
	if len(body.Entries) == 0 {
		count = 0
	}

	// Copy raw messages so subsequent callers cannot mutate session memory.
	copied := make([]json.RawMessage, len(body.Entries))
	for i, e := range body.Entries {
		if e == nil {
			copied[i] = json.RawMessage("null")
			continue
		}
		copied[i] = append(json.RawMessage(nil), e...)
	}

	c.sess.mu.Lock()
	c.sess.previewEntries = copied
	c.sess.previewCount = count
	c.sess.previewUpdatedAt = time.Now().UTC()
	// Refresh recording-phase heartbeat on successful entries push.
	if c.sess.status == StatusRecording {
		c.sess.lastHeartbeatAt = time.Now()
	}
	c.sess.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"ok":    true,
		"count": count,
	})
}

func (c *controlServer) handleEntriesGET(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("session")
	if id == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error":   "missing session",
			"message": "session query parameter is required",
		})
		return
	}
	if id != c.sess.id {
		writeSessionNotFoundJSON(w)
		return
	}

	c.sess.mu.Lock()
	entries := c.sess.previewEntries
	if entries == nil {
		entries = []json.RawMessage{}
	}
	// Shallow-copy slice of raw messages for encode.
	out := make([]json.RawMessage, len(entries))
	copy(out, entries)
	count := c.sess.previewCount
	if count == 0 && len(out) > 0 {
		count = len(out)
	}
	updatedAt := c.sess.previewUpdatedAt
	c.sess.mu.Unlock()

	updated := ""
	if !updatedAt.IsZero() {
		updated = updatedAt.Format(time.RFC3339Nano)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"entries":    out,
		"count":      count,
		"updated_at": updated,
	})
}

// handlePreview serves the live HTML viewer that polls GET /v1/entries.
func (c *controlServer) handlePreview(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	id := r.URL.Query().Get("session")
	if id == "" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`<!DOCTYPE html><html><body><p>missing session</p></body></html>`))
		return
	}
	if id != c.sess.id {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusNotFound)
		_, _ = fmt.Fprintf(w, `<!DOCTYPE html><html><body>
<p id="browser-trace-preview-error" data-error="session not found">Session not found: %s</p>
</body></html>`, htmlEscape(id))
		return
	}

	// Snapshot for optional server-side seed of the table (poll still updates).
	c.sess.mu.Lock()
	entries := c.sess.previewEntries
	count := c.sess.previewCount
	if count == 0 && len(entries) > 0 {
		count = len(entries)
	}
	if entries == nil {
		entries = []json.RawMessage{}
	}
	// Build simple seed rows from request.url when present.
	var seedRows strings.Builder
	for _, raw := range entries {
		var m map[string]any
		if json.Unmarshal(raw, &m) != nil {
			continue
		}
		method, urlStr := "", ""
		if req, ok := m["request"].(map[string]any); ok {
			method, _ = req["method"].(string)
			urlStr, _ = req["url"].(string)
		}
		if urlStr == "" {
			urlStr, _ = m["url"].(string)
		}
		if urlStr == "" {
			continue
		}
		status := ""
		if resp, ok := m["response"].(map[string]any); ok {
			switch n := resp["status"].(type) {
			case float64:
				status = fmt.Sprintf("%d", int(n))
			case json.Number:
				status = n.String()
			case string:
				status = n
			}
		}
		seedRows.WriteString(fmt.Sprintf(
			`<tr><td>%s</td><td class="url">%s</td><td>%s</td></tr>`,
			htmlEscape(method), htmlEscape(urlStr), htmlEscape(status),
		))
	}
	c.sess.mu.Unlock()

	emptyClass := ""
	tableClass := ""
	if count == 0 || seedRows.Len() == 0 {
		emptyClass = ""
		tableClass = "hidden"
	} else {
		emptyClass = "hidden"
		tableClass = ""
	}
	esc := htmlEscape(id)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = fmt.Fprintf(w, `<!DOCTYPE html>
<html><head><meta charset="utf-8"><title>browser-trace live preview</title>
<style>
  body { font-family: system-ui, sans-serif; margin: 1.25rem; max-width: 56rem; }
  h1 { font-size: 1.15rem; margin: 0 0 0.5rem; }
  .meta { color: #555; margin-bottom: 1rem; font-size: 0.9rem; }
  table { width: 100%%; border-collapse: collapse; font-size: 0.9rem; }
  th, td { text-align: left; padding: 0.4rem 0.5rem; border-bottom: 1px solid #e5e5e5; vertical-align: top; }
  th { background: #f6f8fa; position: sticky; top: 0; }
  td.url { word-break: break-all; font-family: ui-monospace, monospace; font-size: 0.85rem; }
  .empty-state { color: #666; padding: 1.5rem 0; }
  .hidden { display: none; }
  #status { color: #888; font-size: 0.85rem; }
</style>
</head>
<body>
<div id="browser-trace-preview" data-browser-trace-preview data-session="%s" data-entry-count="%d" data-count="%d">
  <h1>Live preview</h1>
  <div class="meta">Session <code id="session-id">%s</code> · <span id="entry-count">%d entries</span> · <span id="status">polling /v1/entries…</span></div>
  <div id="empty-state" class="empty-state %s">No requests captured yet (empty / cleared).</div>
  <table id="entries-table" class="preview-entries %s">
    <thead><tr><th>Method</th><th>URL</th><th>Status</th></tr></thead>
    <tbody id="entries-body">%s</tbody>
  </table>
</div>
<script>
(function() {
  var sessionId = %q;
  var body = document.getElementById('entries-body');
  var empty = document.getElementById('empty-state');
  var table = document.getElementById('entries-table');
  var countEl = document.getElementById('entry-count');
  var statusEl = document.getElementById('status');
  var root = document.getElementById('browser-trace-preview');
  function esc(s) {
    return String(s == null ? '' : s)
      .replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;');
  }
  function entryURL(e) {
    if (!e) return '';
    if (e.request && e.request.url) return e.request.url;
    return e.url || '';
  }
  function entryMethod(e) {
    if (e && e.request && e.request.method) return e.request.method;
    return e && e.method ? e.method : '';
  }
  function entryStatus(e) {
    if (e && e.response && e.response.status != null) return e.response.status;
    return '';
  }
  function render(data) {
    var list = (data && data.entries) ? data.entries : [];
    var count = (data && data.count != null) ? data.count : list.length;
    countEl.textContent = count + ' entries';
    root.setAttribute('data-entry-count', String(count));
    root.setAttribute('data-count', String(count));
    if (!list.length) {
      body.innerHTML = '';
      empty.classList.remove('hidden');
      table.classList.add('hidden');
      empty.textContent = 'No requests captured yet (empty / cleared).';
      return;
    }
    empty.classList.add('hidden');
    table.classList.remove('hidden');
    body.innerHTML = list.map(function(e) {
      return '<tr><td>' + esc(entryMethod(e)) + '</td><td class="url">' + esc(entryURL(e)) +
        '</td><td>' + esc(entryStatus(e)) + '</td></tr>';
    }).join('');
  }
  function poll() {
    fetch('/v1/entries?session=' + encodeURIComponent(sessionId))
      .then(function(r) {
        if (!r.ok) throw new Error('HTTP ' + r.status);
        return r.json();
      })
      .then(function(data) {
        statusEl.textContent = 'updated';
        render(data);
      })
      .catch(function(err) {
        statusEl.textContent = 'poll error: ' + (err && err.message ? err.message : err);
      });
  }
  poll();
  setInterval(poll, 1000);
})();
</script>
</body></html>`,
		esc, count, count, esc, count, emptyClass, tableClass, seedRows.String(), id)
}

// --- save ---

func saveArtifacts(sess *session, payload completePayload) error {
	harDoc, entryCount, err := normalizeHAR(payload.HAR)
	if err != nil {
		return fmt.Errorf("normalize HAR: %w", err)
	}

	// Atomic HAR write.
	harPath := filepath.Join(sess.outDir, "recording.har")
	tmpPath := harPath + ".tmp"
	harBytes, err := json.MarshalIndent(harDoc, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(tmpPath, harBytes, 0o644); err != nil {
		return err
	}
	if err := os.Rename(tmpPath, harPath); err != nil {
		return err
	}

	reason := payload.StopReason
	if reason == "" {
		reason = "extension"
	}
	sess.mu.Lock()
	if payload.WindowID != 0 {
		sess.windowID = payload.WindowID
	}
	sess.stopReason = reason
	sess.entryCount = entryCount
	if n, ok := payload.Stats["entry_count"].(float64); ok && entryCount == 0 {
		sess.entryCount = int(n)
	}
	windowID := sess.windowID
	sess.mu.Unlock()

	return writeMeta(sess, reason, map[string]any{
		"entry_count": entryCount,
		"window_id":   windowID,
		"status":      StatusSaved,
	})
}

func writeMeta(sess *session, stopReason string, extra map[string]any) error {
	sess.mu.Lock()
	meta := map[string]any{
		"session_id":  sess.id,
		"started_at":  sess.startedAt.UTC().Format(time.RFC3339Nano),
		"finished_at": time.Now().UTC().Format(time.RFC3339Nano),
		"status":      sess.status,
		"stop_reason": stopReason,
		"entry_count": sess.entryCount,
		"window_id":   sess.windowID,
		"out_dir":     sess.outDir,
	}
	if sess.extensionVer != "" {
		meta["extension_version"] = sess.extensionVer
	}
	if sess.helloOK {
		meta["supports_browser_trace"] = sess.supportsBT
		if len(sess.extensionFeats) > 0 {
			meta["extension_features"] = append([]string(nil), sess.extensionFeats...)
		}
	}
	if len(sess.errors) > 0 {
		meta["errors"] = append([]string(nil), sess.errors...)
	}
	sess.mu.Unlock()
	for k, v := range extra {
		meta[k] = v
	}
	// Prefer failed status when stop_reason is timeout.
	if stopReason == "timeout" || stopReason == "cancelled" {
		meta["status"] = StatusFailed
	}
	b, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(sess.outDir, "meta.json"), b, 0o644)
}

func normalizeHAR(raw json.RawMessage) (map[string]any, int, error) {
	if len(raw) == 0 {
		doc := map[string]any{
			"log": map[string]any{
				"version": "1.2",
				"creator": map[string]any{"name": "browser-trace", "version": "1.0"},
				"entries": []any{},
			},
		}
		return doc, 0, nil
	}
	var v any
	if err := json.Unmarshal(raw, &v); err != nil {
		return nil, 0, err
	}
	// Accept full HAR document or bare entries array / object.
	doc, ok := v.(map[string]any)
	if !ok {
		return nil, 0, fmt.Errorf("HAR root must be object")
	}
	logObj, _ := doc["log"].(map[string]any)
	if logObj == nil {
		// Maybe they sent entries at top level.
		if entries, ok := doc["entries"]; ok {
			logObj = map[string]any{
				"version": "1.2",
				"creator": map[string]any{"name": "browser-trace", "version": "1.0"},
				"entries": entries,
			}
			doc = map[string]any{"log": logObj}
		} else {
			// Wrap as empty-ish document with whatever was sent under log.
			logObj = map[string]any{
				"version": "1.2",
				"creator": map[string]any{"name": "browser-trace", "version": "1.0"},
				"entries": []any{},
			}
			doc = map[string]any{"log": logObj}
		}
	}
	entries, _ := logObj["entries"].([]any)
	// Sort by startedDateTime ascending.
	sort.SliceStable(entries, func(i, j int) bool {
		si := entryStarted(entries[i])
		sj := entryStarted(entries[j])
		return si < sj
	})
	logObj["entries"] = entries
	doc["log"] = logObj
	return doc, len(entries), nil
}

func entryStarted(e any) string {
	m, ok := e.(map[string]any)
	if !ok {
		return ""
	}
	s, _ := m["startedDateTime"].(string)
	return s
}

// --- chrome launcher ---

// openChromeNewWindow launches default-profile Chrome in a new window with
// optional --load-extension. Never passes --user-data-dir.
func openChromeNewWindow(sessionURL, extensionPath string) error {
	args := BuildChromeLaunchArgs(sessionURL, extensionPath)
	switch runtime.GOOS {
	case "darwin":
		// open -na "Google Chrome" --args <chrome-args...>
		cmdArgs := append([]string{"-na", "Google Chrome", "--args"}, args...)
		cmd := exec.Command("open", cmdArgs...)
		return cmd.Start()
	case "linux":
		for _, bin := range []string{"google-chrome", "google-chrome-stable", "chromium", "chromium-browser"} {
			if path, err := exec.LookPath(bin); err == nil {
				cmd := exec.Command(path, args...)
				return cmd.Start()
			}
		}
		return fmt.Errorf("chrome/chromium not found on PATH")
	case "windows":
		cmdArgs := append([]string{"/c", "start", "chrome"}, args...)
		cmd := exec.Command("cmd", cmdArgs...)
		return cmd.Start()
	default:
		return fmt.Errorf("unsupported OS for chrome launch: %s", runtime.GOOS)
	}
}
