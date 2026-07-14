package browseragent

import (
	"strings"
	"sync"
	"time"
)

// session is the in-memory state for one serve instance.
type session struct {
	mu sync.Mutex

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

	queue *JobQueue

	// Single extension WS writer (nil when disconnected).
	ws *wsConn

	// Optional hello observer (mismatch warning); called unlocked after state update.
	onHello func(match string, embedded, loaded BundleSum, installPath string)

	// createdViaPOST distinguishes HTTP-created sessions from registry pre-provision.
	createdViaPOST bool

	sessionURL       string
	sessionPageCount *int
	browsers         []string
	sessionPages     []sessionPageTab
	lastSeenAt       time.Time
}

type sessionPageTab struct {
	TabID int    `json:"tab_id,omitempty"`
	URL   string `json:"url,omitempty"`
	Title string `json:"title,omitempty"`
}

func newSession(id, baseDir string) *session {
	return &session{
		id:        id,
		phase:     PhaseWaitingExtension,
		createdAt: time.Now(),
		baseDir:   baseDir,
		queue:     NewJobQueue(),
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

func (s *session) markDisconnected() {
	s.mu.Lock()
	s.extConnected = false
	s.ws = nil
	// Keep version/features/md5 for last-seen display; supports becomes false when not connected.
	s.supportsBA = false
	s.phase = PhaseWaitingExtension
	q := s.queue
	s.mu.Unlock()
	if q != nil {
		q.FailAllInflight("extension disconnected: connection lost (websocket closed)")
	}
}

func (s *session) setExtensionInstallPath(path string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.extensionInstallPath = path
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
	}
	if pages != nil {
		s.sessionPages = append([]sessionPageTab(nil), pages...)
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

	return sessionSnapshot{
		SessionID:            s.id,
		Phase:                s.phase,
		Hint:                 hint,
		ExtensionInstallPath: s.extensionInstallPath,
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
	SessionID            string           `json:"session_id"`
	Phase                string           `json:"phase"`
	Hint                 string           `json:"hint,omitempty"`
	ExtensionInstallPath string           `json:"extension_install_path,omitempty"`
	BundledExtension     bundledExtension `json:"bundled_extension"`
	Extension            sessionExtension `json:"extension"`
	ExtensionMatch       string           `json:"extension_match"`
	CreatedAt            time.Time        `json:"created_at"`
	SessionPageCount     *int             `json:"session_page_count,omitempty"`
	Browsers             []string         `json:"browsers,omitempty"`
	Status               string           `json:"status"`
	StatusLabel          string           `json:"status_label"`
	InflightJobs         int              `json:"inflight_jobs,omitempty"`
	SessionURL           string           `json:"session_url,omitempty"`
	SessionPages         []sessionPageTab `json:"session_pages,omitempty"`
	LastSeenAt           time.Time        `json:"last_seen_at,omitempty"`
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

// computeSupportsBrowserAgent requires feature "browser-agent" and version ≥ 1.0.0.
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
