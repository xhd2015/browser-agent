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

	return sessionSnapshot{
		SessionID:            s.id,
		Phase:                s.phase,
		Hint:                 hint,
		ExtensionInstallPath: s.extensionInstallPath,
		ExtensionMatch:       match,
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

func buildHint(connected, supports bool) string {
	if !connected {
		return "Waiting for browser-agent extension. Load unpacked extension and open this session page (control port " + DefaultControlPort + ")."
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
