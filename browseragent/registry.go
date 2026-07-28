package browseragent

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

// ErrSessionExists indicates a session id is already registered or its directory exists on disk.
var ErrSessionExists = errors.New("session already exists")

// ErrSessionNotFound indicates the session id is absent from the registry and on-disk layout.
var ErrSessionNotFound = errors.New("session not found")

// ErrSessionExtensionConnected indicates delete was rejected because the extension is connected.
var ErrSessionExtensionConnected = errors.New("cannot delete session: extension connected")

// SessionRegistry holds live sessions keyed by session id.
type SessionRegistry struct {
	mu      sync.RWMutex
	baseDir string
	addr    string
	sessions map[string]*session
}

// CreateSessionResult is returned when a session is created successfully.
type CreateSessionResult struct {
	SessionID  string
	SessionDir string
	MetaPath   string
	SystemPath string
	SessionURL string
}

// NewSessionRegistry constructs a registry rooted at baseDir using addr for meta URLs.
func NewSessionRegistry(baseDir, addr string) *SessionRegistry {
	return &SessionRegistry{
		baseDir:  baseDir,
		addr:     addr,
		sessions: make(map[string]*session),
	}
}

// CreateSessionResultFor returns create metadata for an already-registered session.
func (r *SessionRegistry) CreateSessionResultFor(id string) (*CreateSessionResult, bool) {
	if !r.Exists(id) {
		return nil, false
	}
	sessionDir := SessionDirPath(r.baseDir, id)
	metaPath := filepath.Join(sessionDir, "meta.json")
	sysPath := filepath.Join(sessionDir, "SYSTEM.md")
	absSessionDir, _ := filepath.Abs(sessionDir)
	if absSessionDir == "" {
		absSessionDir = sessionDir
	}
	absMetaPath, _ := filepath.Abs(metaPath)
	absSysPath, _ := filepath.Abs(sysPath)
	baseURL := r.BaseURL()
	return &CreateSessionResult{
		SessionID:  id,
		SessionDir: absSessionDir,
		MetaPath:   absMetaPath,
		SystemPath: absSysPath,
		SessionURL: baseURL + "/go?session=" + id,
	}, true
}

// CreateOpts configures optional fields when creating a session.
type CreateOpts struct {
	// Browser selects the extension install tree: "chrome" (default/empty) or "firefox".
	// When "firefox", meta.json stamps browser=firefox and extension_install_path under
	// browser-agent-firefox via EnsureCanonicalFirefoxExtension.
	Browser string
}

// Create validates id, writes session artifacts, and registers the session (Chrome default).
func (r *SessionRegistry) Create(id string) (*CreateSessionResult, error) {
	return r.CreateWithOpts(id, CreateOpts{})
}

// CreateWithOpts is like Create but stamps browser-specific extension path and meta.
func (r *SessionRegistry) CreateWithOpts(id string, opts CreateOpts) (*CreateSessionResult, error) {
	if err := ValidateSessionID(id); err != nil {
		return nil, err
	}

	browser := strings.ToLower(strings.TrimSpace(opts.Browser))
	switch browser {
	case "", "chrome":
		browser = "chrome"
	case "firefox":
		// ok
	default:
		return nil, fmt.Errorf("unknown browser %q", opts.Browser)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.sessions[id]; ok {
		return nil, fmt.Errorf("%w: %s", ErrSessionExists, id)
	}
	if SessionDirExists(r.baseDir, id) {
		return nil, fmt.Errorf("%w: %s", ErrSessionExists, id)
	}

	sessionDir := SessionDirPath(r.baseDir, id)
	if err := os.MkdirAll(sessionDir, 0o755); err != nil {
		return nil, fmt.Errorf("create session dir: %w", err)
	}

	var extPath, extVer string
	var extErr error
	if browser == "firefox" {
		extPath, extVer, extErr = EnsureCanonicalFirefoxExtension()
	} else {
		extPath, extVer, extErr = EnsureCanonicalExtension()
	}
	var embeddedSum BundleSum
	if extErr == nil && extPath != "" {
		if sum, sumErr := EnsureExtensionBundleSum(extPath, extVer); sumErr == nil {
			embeddedSum = sum
		}
		if embeddedSum.Version == "" {
			embeddedSum.Version = extVer
		}
	}

	sysPath := filepath.Join(sessionDir, "SYSTEM.md")
	if err := os.WriteFile(sysPath, []byte(FormatSystemPrompt(id)), 0o644); err != nil {
		return nil, fmt.Errorf("write SYSTEM.md: %w", err)
	}
	absSysPath, err := filepath.Abs(sysPath)
	if err != nil {
		absSysPath = sysPath
	}

	baseURL := "http://" + r.addr
	sessionURL := baseURL + "/go?session=" + id
	controlPort := controlPortFromAddr(r.addr)

	metaPath := filepath.Join(sessionDir, "meta.json")
	meta := map[string]any{
		"session_id":         id,
		"addr":               r.addr,
		"base_url":           baseURL,
		"session_url":        sessionURL,
		"system_prompt_path": absSysPath,
		"product":            ProductName,
		"control_port":       controlPort,
	}
	if browser == "firefox" {
		meta["browser"] = "firefox"
	}
	if extPath != "" {
		meta["extension_install_path"] = extPath
		meta["extension_version"] = embeddedSum.Version
		if embeddedSum.MD5 != "" {
			meta["extension_md5"] = embeddedSum.MD5
		}
	}
	metaBytes, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal meta.json: %w", err)
	}
	if err := os.WriteFile(metaPath, append(metaBytes, '\n'), 0o644); err != nil {
		return nil, fmt.Errorf("write meta.json: %w", err)
	}

	absSessionDir, _ := filepath.Abs(sessionDir)
	if absSessionDir == "" {
		absSessionDir = sessionDir
	}
	absMetaPath, err := filepath.Abs(metaPath)
	if err != nil {
		absMetaPath = metaPath
	}

	sess := newSession(id, r.baseDir)
	sess.setSessionURL(sessionURL)
	if extPath != "" {
		sess.setExtensionInstallPath(extPath)
		sess.setEmbeddedIdentity(embeddedSum.Version, embeddedSum.MD5)
	}
	if browser == "firefox" {
		// Stamp preferred browser so live snap / injectSessionBoot agree before telemetry.
		sess.updateTelemetry("firefox", nil, nil)
	}
	r.sessions[id] = sess

	return &CreateSessionResult{
		SessionID:  id,
		SessionDir: absSessionDir,
		MetaPath:   absMetaPath,
		SystemPath: absSysPath,
		SessionURL: sessionURL,
	}, nil
}

// Get returns a live session from the registry.
func (r *SessionRegistry) Get(id string) (*session, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	sess, ok := r.sessions[id]
	return sess, ok
}

// List returns snapshots of all registered sessions sorted by session id.
func (r *SessionRegistry) List() []sessionSnapshot {
	r.mu.RLock()
	defer r.mu.RUnlock()

	ids := make([]string, 0, len(r.sessions))
	for id := range r.sessions {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	out := make([]sessionSnapshot, 0, len(ids))
	for _, id := range ids {
		out = append(out, r.snapshot(r.sessions[id]))
	}
	return out
}

// Addr returns the registry listen address metadata (host:port).
func (r *SessionRegistry) Addr() string {
	return r.addr
}

// CountExtensionConnected returns the number of sessions with extension.connected == true.
func (r *SessionRegistry) CountExtensionConnected() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	n := 0
	for _, sess := range r.sessions {
		if sess.isExtensionConnected() {
			n++
		}
	}
	return n
}

// BaseURL returns http://{addr} for hint and session URL construction.
func (r *SessionRegistry) BaseURL() string {
	return "http://" + r.addr
}

// onlySessionID returns the sole session id when exactly one session is registered.
func (r *SessionRegistry) onlySessionID() (string, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if len(r.sessions) != 1 {
		return "", false
	}
	for id := range r.sessions {
		return id, true
	}
	return "", false
}

// snapshot returns a session snapshot with disconnected hint post-processing.
func (r *SessionRegistry) snapshot(sess *session) sessionSnapshot {
	snap := sess.snapshot()
	if !snap.Extension.Connected {
		snap.Hint = buildDisconnectedHint(sess.id, r.BaseURL())
	}
	return snap
}

// Exists reports whether id is in the registry or has an on-disk session directory.
func (r *SessionRegistry) Exists(id string) bool {
	r.mu.RLock()
	_, inRegistry := r.sessions[id]
	r.mu.RUnlock()
	if inRegistry {
		return true
	}
	return SessionDirExists(r.baseDir, id)
}

// RestoreSessionsFromDisk scans {baseDir}/sessions/*/ for valid session
// directories with readable meta.json and registers them in-memory as
// waiting_extension (extension.connected=false). It does not call Create and
// does not rewrite disk files. Best-effort: invalid dirname, missing meta, or
// corrupt meta are skipped. A missing sessions/ directory is not an error.
func RestoreSessionsFromDisk(r *SessionRegistry) error {
	if r == nil {
		return errors.New("nil session registry")
	}
	sessionsRoot := filepath.Join(r.baseDir, "sessions")
	entries, err := os.ReadDir(sessionsRoot)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read sessions dir: %w", err)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		id := e.Name()
		if ValidateSessionID(id) != nil {
			continue
		}
		if _, ok := r.sessions[id]; ok {
			continue
		}
		metaPath := filepath.Join(sessionsRoot, id, "meta.json")
		data, err := os.ReadFile(metaPath)
		if err != nil {
			continue
		}
		var meta map[string]any
		if err := json.Unmarshal(data, &meta); err != nil {
			continue
		}

		sess := newSession(id, r.baseDir)
		if url, ok := meta["session_url"].(string); ok && strings.TrimSpace(url) != "" {
			sess.setSessionURL(url)
		} else {
			// Fall back to current registry addr when meta lacks session_url.
			sess.setSessionURL(r.BaseURL() + "/go?session=" + id)
		}
		if p, ok := meta["extension_install_path"].(string); ok && strings.TrimSpace(p) != "" {
			sess.setExtensionInstallPath(p)
		}
		ver, _ := meta["extension_version"].(string)
		md5, _ := meta["extension_md5"].(string)
		if strings.TrimSpace(ver) != "" || strings.TrimSpace(md5) != "" {
			sess.setEmbeddedIdentity(ver, md5)
		}
		// Restore stamped preferred browser so inject/session-info agree after restart.
		if b, ok := meta["browser"].(string); ok {
			b = strings.ToLower(strings.TrimSpace(b))
			if b == "firefox" || b == "chrome" {
				sess.updateTelemetry(b, nil, nil)
			}
		}
		r.sessions[id] = sess
	}
	return nil
}

// Delete removes a session from the registry and deletes its on-disk directory.
// Returns ErrSessionNotFound when the id is absent from the registry and disk.
// Returns ErrSessionExtensionConnected when the extension is connected.
// Disk-only directories (on disk but not in the registry) are removed without error.
func (r *SessionRegistry) Delete(id string) error {
	r.mu.RLock()
	sess, inRegistry := r.sessions[id]
	r.mu.RUnlock()

	if !inRegistry {
		if !SessionDirExists(r.baseDir, id) {
			return fmt.Errorf("%w: %s", ErrSessionNotFound, id)
		}
		return os.RemoveAll(SessionDirPath(r.baseDir, id))
	}

	if sess.isExtensionConnected() {
		return ErrSessionExtensionConnected
	}

	if sess.queue != nil {
		sess.queue.FailAllInflight("session deleted")
	}

	r.mu.Lock()
	delete(r.sessions, id)
	r.mu.Unlock()

	if SessionDirExists(r.baseDir, id) {
		if err := os.RemoveAll(SessionDirPath(r.baseDir, id)); err != nil {
			return fmt.Errorf("remove session dir: %w", err)
		}
	}
	return nil
}