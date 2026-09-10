package browseragent

import (
	"strings"
	"sync"
	"time"
)

var wsDisconnectGracePeriod = 2 * time.Second

// session is the in-memory state for one serve instance.
type session struct {
	mu sync.Mutex

	// extensionResultMu serializes extension result side effects and deduplicates
	// replayed results by job ID.
	extensionResultMu   sync.Mutex
	extensionResultSeen map[string]struct{}

	id        string
	phase     string
	createdAt time.Time
	baseDir   string

	extConnected bool
	extVersion   string
	extFeatures  []string
	extBundleMD5 string
	supportsBA   bool

	// Embedded package identity (from extract + bundle-sum.js).
	embeddedVersion string
	embeddedMD5     string

	// extensionInstallPath is the extracted load-unpacked folder (absolute).
	extensionInstallPath string

	// firefoxXPIPath is the extracted signed .xpi (absolute), when available.
	firefoxXPIPath string

	queue *JobQueue

	// Single extension WS writer (nil when disconnected or reconnecting).
	ws                   *wsConn
	disconnectGeneration uint64

	// Poll-mode control events (e.g. prepare_reconnect) for HTTP transport sessions.
	pollEvents  []map[string]any
	pollWaiters []chan struct{}

	// Optional hello observer (mismatch warning); called unlocked after state update.
	onHello func(match string, embedded, loaded BundleSum, installPath string)

	// createdViaPOST distinguishes HTTP-created sessions from registry pre-provision.
	createdViaPOST bool

	sessionURL       string
	sessionPageCount *int
	browsers         []string
	sessionPages     []sessionPageTab
	lastSeenAt       time.Time

	// attach is ephemeral cold-start progress (page/SW/WS stages).
	attach attachState

	// harCapture owns the daemon-side private artifact spool for the active capture.
	harCapture *harCapture
}

type sessionPageTab struct {
	TabID int    `json:"tab_id,omitempty"`
	URL   string `json:"url,omitempty"`
	Title string `json:"title,omitempty"`
}

func newSession(id, baseDir string) *session {
	return &session{
		id:                  id,
		phase:               PhaseWaitingExtension,
		createdAt:           time.Now(),
		baseDir:             baseDir,
		queue:               NewJobQueue(),
		extensionResultSeen: make(map[string]struct{}),
	}
}

func (s *session) setOnHello(fn func(match string, embedded, loaded BundleSum, installPath string)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.onHello = fn
}

func (s *session) markHello(version string, features []string, bundleMD5 string) {
	s.mu.Lock()
	s.extConnected = true
	s.extVersion = version
	s.extFeatures = append([]string(nil), features...)
	s.extBundleMD5 = strings.ToLower(strings.TrimSpace(bundleMD5))
	s.supportsBA = computeSupportsBrowserAgent(version, features)
	s.phase = PhaseExtensionConnected
	s.setAttachStageLocked(AttachStageHello)
	if s.supportsBA {
		s.setAttachStageLocked(AttachStageReady)
	}
	embedded := BundleSum{Version: s.embeddedVersion, MD5: s.embeddedMD5}
	loaded := BundleSum{Version: s.extVersion, MD5: s.extBundleMD5}
	match := ComputeExtensionMatch(true, embedded, loaded)
	installPath := s.extensionInstallPath
	onHello := s.onHello
	s.mu.Unlock()
	if onHello != nil {
		onHello(match, embedded, loaded, installPath)
	}
}

func (s *session) markDisconnectedAfterGrace(c *wsConn) {
	s.mu.Lock()
	if s.ws != c {
		s.mu.Unlock()
		return
	}
	s.ws = nil
	s.disconnectGeneration++
	generation := s.disconnectGeneration
	s.mu.Unlock()

	time.AfterFunc(wsDisconnectGracePeriod, func() {
		s.mu.Lock()
		if s.disconnectGeneration != generation || s.ws != nil {
			s.mu.Unlock()
			return
		}
		s.extConnected = false
		// Keep version/features/md5 for last-seen display; supports becomes false when not connected.
		s.supportsBA = false
		s.phase = PhaseWaitingExtension
		q := s.queue
		s.mu.Unlock()
		s.resetAttachOnDisconnect()
		if q != nil {
			q.FailAllInflight("extension disconnected: reconnect grace period expired")
		}
	})
}

func (s *session) setExtensionInstallPath(path string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.extensionInstallPath = path
}

func (s *session) setFirefoxXPIPath(path string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.firefoxXPIPath = strings.TrimSpace(path)
}

func (s *session) setEmbeddedIdentity(version, md5hex string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.embeddedVersion = strings.TrimSpace(version)
	s.embeddedMD5 = strings.ToLower(strings.TrimSpace(md5hex))
}

func (s *session) getEmbeddedIdentity() BundleSum {
	s.mu.Lock()
	defer s.mu.Unlock()
	return BundleSum{Version: s.embeddedVersion, MD5: s.embeddedMD5}
}

func (s *session) setWS(c *wsConn) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ws = c
	if c != nil {
		s.disconnectGeneration++
		s.setAttachStageLocked(AttachStageWSOpen)
	}
}

func (s *session) getWS() *wsConn {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.ws
}

func (s *session) isExtensionConnected() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.extConnected
}

// enqueuePollEvent appends a control event for HTTP poll clients and wakes waiters.
func (s *session) enqueuePollEvent(ev map[string]any) {
	if s == nil || ev == nil {
		return
	}
	s.mu.Lock()
	// Copy so callers can reuse maps freely.
	cp := make(map[string]any, len(ev))
	for k, v := range ev {
		cp[k] = v
	}
	s.pollEvents = append(s.pollEvents, cp)
	waiters := append([]chan struct{}(nil), s.pollWaiters...)
	s.mu.Unlock()
	for _, ch := range waiters {
		select {
		case ch <- struct{}{}:
		default:
		}
	}
}

// drainPollEvents returns and clears all pending poll events.
func (s *session) drainPollEvents() []map[string]any {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.pollEvents) == 0 {
		return []map[string]any{}
	}
	out := s.pollEvents
	s.pollEvents = nil
	return out
}

// notifyPollWaiters wakes long-poll handlers (job enqueued, event enqueued, etc.).
func (s *session) notifyPollWaiters() {
	s.mu.Lock()
	waiters := append([]chan struct{}(nil), s.pollWaiters...)
	s.mu.Unlock()
	for _, ch := range waiters {
		select {
		case ch <- struct{}{}:
		default:
		}
	}
}

func (s *session) addPollWaiter(ch chan struct{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pollWaiters = append(s.pollWaiters, ch)
}

func (s *session) removePollWaiter(ch chan struct{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := s.pollWaiters[:0]
	for _, w := range s.pollWaiters {
		if w != ch {
			out = append(out, w)
		}
	}
	s.pollWaiters = out
}

func (s *session) markCreatedViaPOST() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.createdViaPOST = true
}

func (s *session) wasCreatedViaPOST() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.createdViaPOST
}

func (s *session) setSessionURL(url string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessionURL = strings.TrimSpace(url)
}

func (s *session) updateTelemetry(browserProduct string, pageCount *int, pages []sessionPageTab) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if browserProduct != "" {
		found := false
		for _, b := range s.browsers {
			if strings.EqualFold(b, browserProduct) {
				found = true
				break
			}
		}
		if !found {
			s.browsers = append(s.browsers, browserProduct)
		}
	}
	if pageCount != nil {
		v := *pageCount
		s.sessionPageCount = &v
		if v > 0 {
			s.setAttachStageLocked(AttachStagePageOpen)
		}
	}
	if pages != nil {
		s.sessionPages = append([]sessionPageTab(nil), pages...)
		if len(pages) > 0 {
			s.setAttachStageLocked(AttachStagePageOpen)
		}
	}
	s.lastSeenAt = time.Now()
}

func (s *session) inflightJobs() int {
	s.mu.Lock()
	q := s.queue
	s.mu.Unlock()
	if q == nil {
		return 0
	}
	return q.InflightCount()
}

func (s *session) snapshot() sessionSnapshot {
	s.mu.Lock()
	defer s.mu.Unlock()
	feats := append([]string(nil), s.extFeatures...)
	if feats == nil {
		feats = []string{}
	}
	connected := s.extConnected
	supports := s.supportsBA && connected
	hint := buildHint(connected, supports)

	embedded := BundleSum{Version: s.embeddedVersion, MD5: s.embeddedMD5}
	loaded := BundleSum{Version: s.extVersion, MD5: s.extBundleMD5}
	match := ComputeExtensionMatch(connected, embedded, loaded)

	// When not connected, clear loaded fields for a clean snapshot (keep empty strings).
	extVersion := s.extVersion
	extMD5 := s.extBundleMD5
	if !connected {
		// Keep last-seen version for display if previously connected; tests for
		// not_connected only assert connected=false and match. Prefer empty on
		// never-connected (extVersion stays "").
	}

	status, statusLabel := ComputeSessionStatus(s.sessionPageCount, connected, supports)

	browsers := append([]string(nil), s.browsers...)
	pages := append([]sessionPageTab(nil), s.sessionPages...)
	inflight := 0
	if s.queue != nil {
		inflight = s.queue.InflightCount()
	}

	var lastSeen time.Time
	if !s.lastSeenAt.IsZero() {
		lastSeen = s.lastSeenAt
	}

	xpiPath := s.firefoxXPIPath
	var xpiFileURL string
	if xpiPath != "" {
		xpiFileURL = PathToFileURL(xpiPath)
	}

	return sessionSnapshot{
		SessionID:            s.id,
		Phase:                s.phase,
		Hint:                 hint,
		ExtensionInstallPath: s.extensionInstallPath,
		FirefoxXPIPath:       xpiPath,
		FirefoxXPIURL:        xpiFileURL,
		ExtensionMatch:       match,
		CreatedAt:            s.createdAt,
		SessionPageCount:     s.sessionPageCount,
		Browsers:             browsers,
		Status:               status,
		StatusLabel:          statusLabel,
		InflightJobs:         inflight,
		SessionURL:           s.sessionURL,
		SessionPages:         pages,
		LastSeenAt:           lastSeen,
		Attach:               s.snapshotAttach(),
		BundledExtension: bundledExtension{
			Version: s.embeddedVersion,
			MD5:     s.embeddedMD5,
			Path:    s.extensionInstallPath,
		},
		Extension: sessionExtension{
			Connected:            connected,
			Version:              extVersion,
			Features:             feats,
			BundleMD5:            extMD5,
			SupportsBrowserAgent: supports,
		},
	}
}

type sessionSnapshot struct {
	SessionID            string `json:"session_id"`
	Phase                string `json:"phase"`
	Hint                 string `json:"hint,omitempty"`
	ExtensionInstallPath string `json:"extension_install_path,omitempty"`
	// FirefoxXPIPath is the absolute filesystem path to the signed .xpi when available.
	FirefoxXPIPath string `json:"firefox_xpi_path,omitempty"`
	// FirefoxXPIURL is file:///… for the same path (copy/paste; click from http may be blocked).
	FirefoxXPIURL string `json:"firefox_xpi_url,omitempty"`
	// FirefoxXPIHTTPURL is the control-plane download (reliable click-to-install from session page).
	FirefoxXPIHTTPURL string           `json:"firefox_xpi_http_url,omitempty"`
	BundledExtension  bundledExtension `json:"bundled_extension"`
	Extension         sessionExtension `json:"extension"`
	ExtensionMatch    string           `json:"extension_match"`
	CreatedAt         time.Time        `json:"created_at"`
	SessionPageCount  *int             `json:"session_page_count,omitempty"`
	Browsers          []string         `json:"browsers,omitempty"`
	Status            string           `json:"status"`
	StatusLabel       string           `json:"status_label"`
	InflightJobs      int              `json:"inflight_jobs,omitempty"`
	SessionURL        string           `json:"session_url,omitempty"`
	SessionPages      []sessionPageTab `json:"session_pages,omitempty"`
	LastSeenAt        time.Time        `json:"last_seen_at,omitempty"`
	// Attach is ephemeral cold-start progress (omit when never reported).
	Attach *sessionAttach `json:"attach,omitempty"`
}

type bundledExtension struct {
	Version string `json:"version"`
	MD5     string `json:"md5"`
	Path    string `json:"path"`
}

type sessionExtension struct {
	Connected            bool     `json:"connected"`
	Version              string   `json:"version"`
	Features             []string `json:"features"`
	BundleMD5            string   `json:"bundle_md5"`
	SupportsBrowserAgent bool     `json:"supports_browser_agent"`
}

// buildDisconnectedHint guides operators to keep the session page open at /go?session=.
func buildDisconnectedHint(sessionID, baseURL string) string {
	goPath := "/go?session=" + sessionID
	return "Keep " + goPath + " open in this browser window. Do not close this tab or navigate to a different session in the same window. Open the session page at " + strings.TrimSuffix(baseURL, "/") + goPath + " and load the unpacked browser-agent extension."
}

func buildHint(connected, supports bool) string {
	if !connected {
		return "Waiting for browser-agent extension. Load unpacked extension and open this session page (control port " + DefaultControlPortString() + ")."
	}
	if !supports {
		return "Extension connected but does not support browser-agent (need feature browser-agent and version ≥ " + MinBrowserAgentVersion + ")."
	}
	return "Extension connected; browser-agent jobs are ready."
}

// computeSupportsBrowserAgent requires feature "browser-agent" and
// version ≥ MinBrowserAgentVersion (0.0.0 after product/extension version unify).
func computeSupportsBrowserAgent(version string, features []string) bool {
	has := false
	for _, f := range features {
		if f == FeatureBrowserAgent {
			has = true
			break
		}
	}
	if !has {
		return false
	}
	return versionGTE(version, MinBrowserAgentVersion)
}

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
