package browseragent

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	maxHARArtifactBytes int64 = 256 << 20
	maxHARArtifacts           = 256
	maxHARTotalBytes    int64 = 1 << 30
)

var harLateResultGracePeriod = 30 * time.Second

type harArtifact struct {
	TabID  int    `json:"tab_id"`
	Size   int64  `json:"size"`
	Path   string `json:"-"`
	SHA256 string `json:"-"`
}

type harCapture struct {
	mu sync.Mutex

	ID             string
	Token          string
	StartedAt      time.Time
	EndedAt        time.Time
	Dir            string
	Artifacts      map[int]harArtifact
	TotalBytes     int64
	StartConfirmed bool
	EndResult      *JobResult
	Ending         bool
	Ended          bool
	Closed         bool
}

func randomHex(bytes int) (string, error) {
	buf := make([]byte, bytes)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func (s *session) beginHARCapture() (*harCapture, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.harCapture != nil {
		return nil, errors.New("HAR capture is already active for this session")
	}
	id, err := randomHex(12)
	if err != nil {
		return nil, fmt.Errorf("generate HAR capture id: %w", err)
	}
	token, err := randomHex(24)
	if err != nil {
		return nil, fmt.Errorf("generate HAR artifact token: %w", err)
	}
	dir := filepath.Join(SessionDirPath(s.baseDir, s.id), "har-spool", id)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, fmt.Errorf("create HAR artifact spool: %w", err)
	}
	if err := os.Chmod(filepath.Dir(dir), 0o700); err != nil {
		_ = os.RemoveAll(dir)
		return nil, fmt.Errorf("secure HAR artifact spool: %w", err)
	}
	if err := os.Chmod(dir, 0o700); err != nil {
		_ = os.RemoveAll(dir)
		return nil, fmt.Errorf("secure HAR capture spool: %w", err)
	}
	capture := &harCapture{
		ID:        id,
		Token:     token,
		StartedAt: time.Now().UTC(),
		Dir:       dir,
		Artifacts: make(map[int]harArtifact),
	}
	s.harCapture = capture
	return capture, nil
}

func (s *session) activeHARCapture() (*harCapture, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.harCapture == nil {
		return nil, errors.New("no active HAR capture for this session")
	}
	return s.harCapture, nil
}

func (s *session) prepareHAREnd() (*harCapture, error) {
	capture, err := s.activeHARCapture()
	if err != nil {
		return nil, err
	}
	capture.mu.Lock()
	defer capture.mu.Unlock()
	if capture.Ended {
		return nil, errors.New("HAR capture has already ended")
	}
	if capture.Ending {
		return nil, errors.New("HAR capture is already ending")
	}
	capture.Ending = true
	return capture, nil
}

func cloneJobResult(result JobResult) JobResult {
	cloned := result
	if result.Data != nil {
		cloned.Data = make(map[string]any, len(result.Data))
		for key, value := range result.Data {
			cloned.Data[key] = value
		}
	}
	return cloned
}

func (s *session) completedHAREnd() (JobResult, bool) {
	capture, err := s.activeHARCapture()
	if err != nil {
		return JobResult{}, false
	}
	capture.mu.Lock()
	defer capture.mu.Unlock()
	if !capture.Ended || capture.EndResult == nil || capture.Closed {
		return JobResult{}, false
	}
	return cloneJobResult(*capture.EndResult), true
}

func (s *session) finishHAREnd(capture *harCapture, result *JobResult) {
	if capture == nil {
		return
	}
	capture.mu.Lock()
	capture.Ending = false
	if result != nil && result.OK {
		stored := cloneJobResult(*result)
		capture.EndResult = &stored
		capture.Ended = true
		capture.EndedAt = time.Now().UTC()
	}
	capture.mu.Unlock()
}

func (capture *harCapture) startIsConfirmed() bool {
	if capture == nil {
		return false
	}
	capture.mu.Lock()
	defer capture.mu.Unlock()
	return capture.StartConfirmed
}

func (s *session) confirmHARStart(capture *harCapture) {
	if capture == nil {
		return
	}
	capture.mu.Lock()
	capture.StartConfirmed = true
	capture.mu.Unlock()
}

// reconcileHARHello aligns daemon and extension capture ownership after reconnect.
// It returns an extension capture id that must be rolled back as stale.
func (s *session) reconcileHARHello(extensionCaptureID string, pendingEndCaptureIDs []string) string {
	extensionCaptureID = strings.TrimSpace(extensionCaptureID)
	s.mu.Lock()
	capture := s.harCapture
	s.mu.Unlock()
	if capture == nil {
		return extensionCaptureID
	}
	capture.mu.Lock()
	captureID := capture.ID
	ended := capture.Ended
	startConfirmed := capture.StartConfirmed
	capture.mu.Unlock()
	if ended {
		return extensionCaptureID
	}
	if extensionCaptureID == captureID {
		s.confirmHARStart(capture)
		return ""
	}
	if extensionCaptureID == "" && !startConfirmed {
		// Start ownership begins before enqueue/push. Preserve that uncertain
		// state until a start result or an explicit end failure reconciles it.
		return ""
	}
	for _, pendingID := range pendingEndCaptureIDs {
		if strings.TrimSpace(pendingID) == captureID {
			// A successful end result was produced while disconnected. Keep its
			// spool until the extension flushes that queued result after hello.
			return extensionCaptureID
		}
	}
	// The extension cannot recreate a missing/mismatched in-memory recorder.
	// Drop only the daemon's incomplete spool; the status acknowledgement tells
	// the extension to roll back its mismatched capture, if one exists.
	s.abortHARCapture(capture)
	return extensionCaptureID
}

func (s *session) cleanupHARCapture(capture *harCapture) error {
	if capture == nil {
		return nil
	}
	capture.mu.Lock()
	if capture.Closed {
		capture.mu.Unlock()
		return nil
	}
	capture.Closed = true
	dir := capture.Dir
	capture.mu.Unlock()

	if err := os.RemoveAll(dir); err != nil {
		capture.mu.Lock()
		capture.Closed = false
		capture.mu.Unlock()
		return fmt.Errorf("remove HAR capture spool: %w", err)
	}
	s.mu.Lock()
	if s.harCapture == capture {
		s.harCapture = nil
	}
	s.mu.Unlock()
	return nil
}

func (s *session) abortHARCapture(capture *harCapture) {
	_ = s.cleanupHARCapture(capture)
}

func (s *session) resolveHARCapture(id, token string) (*harCapture, error) {
	s.mu.Lock()
	capture := s.harCapture
	s.mu.Unlock()
	if capture == nil || capture.ID != id {
		return nil, errors.New("HAR capture not found")
	}
	if token == "" || capture.Token != token {
		return nil, errors.New("invalid HAR artifact token")
	}
	capture.mu.Lock()
	closed := capture.Closed
	capture.mu.Unlock()
	if closed {
		return nil, errors.New("HAR capture is closed")
	}
	return capture, nil
}

func (s *session) harCaptureForJob(job Job) *harCapture {
	captureID, _ := job.Params["capture_id"].(string)
	if captureID == "" {
		return nil
	}
	s.mu.Lock()
	capture := s.harCapture
	s.mu.Unlock()
	if capture == nil || capture.ID != captureID {
		return nil
	}
	return capture
}

func (s *session) scheduleHARReconciliation(job Job, capture *harCapture) {
	if capture == nil {
		return
	}
	time.AfterFunc(harLateResultGracePeriod, func() {
		current := s.harCaptureForJob(job)
		if current != capture {
			return
		}
		capture.mu.Lock()
		ended := capture.Ended
		capture.mu.Unlock()
		switch job.Type {
		case JobTypeHARStart:
			// Preserve uncertain starts so a later har_end can reconcile the extension.
			return
		case JobTypeHAREnd:
			if !ended {
				s.finishHAREnd(capture, nil)
			}
		}
	})
}

func validateHARFile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	type harCreator struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	}
	type harLog struct {
		Version string          `json:"version"`
		Creator *harCreator     `json:"creator"`
		Entries json.RawMessage `json:"entries"`
	}
	var envelope struct {
		Log *harLog `json:"log"`
	}
	dec := json.NewDecoder(f)
	if err := dec.Decode(&envelope); err != nil {
		return fmt.Errorf("invalid HAR JSON: %w", err)
	}
	var trailing any
	if err := dec.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			return errors.New("invalid HAR JSON: trailing JSON value")
		}
		return fmt.Errorf("invalid HAR JSON after envelope: %w", err)
	}
	if envelope.Log == nil {
		return errors.New("invalid HAR: missing log object")
	}
	if strings.TrimSpace(envelope.Log.Version) == "" {
		return errors.New("invalid HAR: missing log.version")
	}
	if envelope.Log.Creator == nil || strings.TrimSpace(envelope.Log.Creator.Name) == "" || strings.TrimSpace(envelope.Log.Creator.Version) == "" {
		return errors.New("invalid HAR: missing log.creator name or version")
	}
	var entries []json.RawMessage
	entriesJSON := strings.TrimSpace(string(envelope.Log.Entries))
	if entriesJSON == "" || entriesJSON == "null" || json.Unmarshal(envelope.Log.Entries, &entries) != nil {
		return errors.New("invalid HAR: log.entries must be an array")
	}
	return nil
}

func (capture *harCapture) storeArtifact(tabID int, body io.Reader) (harArtifact, error) {
	if tabID <= 0 {
		return harArtifact{}, errors.New("tab_id must be a positive integer")
	}
	capture.mu.Lock()
	defer capture.mu.Unlock()
	if capture.Closed {
		return harArtifact{}, errors.New("HAR capture is closed")
	}
	if existing, exists := capture.Artifacts[tabID]; exists {
		hash := sha256.New()
		n, err := io.Copy(hash, io.LimitReader(body, maxHARArtifactBytes+1))
		if err != nil {
			return harArtifact{}, err
		}
		if n > maxHARArtifactBytes {
			return harArtifact{}, fmt.Errorf("HAR artifact exceeds %d bytes", maxHARArtifactBytes)
		}
		digest := hex.EncodeToString(hash.Sum(nil))
		if n != existing.Size || digest != existing.SHA256 {
			return harArtifact{}, fmt.Errorf("HAR artifact for tab %d conflicts with existing upload", tabID)
		}
		return existing, nil
	}
	if len(capture.Artifacts) >= maxHARArtifacts {
		return harArtifact{}, fmt.Errorf("HAR capture exceeds %d artifacts", maxHARArtifacts)
	}
	remaining := maxHARTotalBytes - capture.TotalBytes
	if remaining <= 0 {
		return harArtifact{}, fmt.Errorf("HAR capture exceeds %d total bytes", maxHARTotalBytes)
	}
	finalPath := filepath.Join(capture.Dir, fmt.Sprintf("tab-%d.har", tabID))
	tmp, err := os.CreateTemp(capture.Dir, fmt.Sprintf(".tab-%d-*.tmp", tabID))
	if err != nil {
		return harArtifact{}, err
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)
	if err := tmp.Chmod(0o600); err != nil {
		_ = tmp.Close()
		return harArtifact{}, err
	}
	limit := maxHARArtifactBytes
	if remaining < limit {
		limit = remaining
	}
	hash := sha256.New()
	n, copyErr := io.Copy(io.MultiWriter(tmp, hash), io.LimitReader(body, limit+1))
	closeErr := tmp.Close()
	if copyErr != nil {
		return harArtifact{}, copyErr
	}
	if closeErr != nil {
		return harArtifact{}, closeErr
	}
	if n > maxHARArtifactBytes {
		return harArtifact{}, fmt.Errorf("HAR artifact exceeds %d bytes", maxHARArtifactBytes)
	}
	if n > remaining {
		return harArtifact{}, fmt.Errorf("HAR capture exceeds %d total bytes", maxHARTotalBytes)
	}
	if err := validateHARFile(tmpPath); err != nil {
		return harArtifact{}, err
	}
	if err := os.Rename(tmpPath, finalPath); err != nil {
		return harArtifact{}, err
	}
	artifact := harArtifact{TabID: tabID, Size: n, Path: finalPath, SHA256: hex.EncodeToString(hash.Sum(nil))}
	capture.Artifacts[tabID] = artifact
	capture.TotalBytes += n
	return artifact, nil
}

func (capture *harCapture) artifactList() []harArtifact {
	capture.mu.Lock()
	defer capture.mu.Unlock()
	out := make([]harArtifact, 0, len(capture.Artifacts))
	for _, artifact := range capture.Artifacts {
		out = append(out, artifact)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].TabID < out[j].TabID })
	return out
}

func (c *controlServer) handleHARArtifact(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Pragma", "no-cache")
	sessionID := strings.TrimSpace(r.URL.Query().Get("session"))
	captureID := strings.TrimSpace(r.URL.Query().Get("capture"))
	token := strings.TrimSpace(r.URL.Query().Get("token"))
	tabID, err := strconv.Atoi(strings.TrimSpace(r.URL.Query().Get("tab_id")))
	if sessionID == "" || captureID == "" || token == "" || err != nil || tabID <= 0 {
		http.Error(w, "session, capture, token, and positive tab_id are required", http.StatusBadRequest)
		return
	}
	sess, ok := c.registry.Get(sessionID)
	if !ok {
		c.writeSessionNotFound(w)
		return
	}
	capture, err := sess.resolveHARCapture(captureID, token)
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
	switch r.Method {
	case http.MethodPut:
		if r.ContentLength > maxHARArtifactBytes {
			http.Error(w, fmt.Sprintf("HAR artifact exceeds %d bytes", maxHARArtifactBytes), http.StatusRequestEntityTooLarge)
			return
		}
		artifact, err := capture.storeArtifact(tabID, r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(artifact)
	case http.MethodGet:
		capture.mu.Lock()
		artifact, exists := capture.Artifacts[tabID]
		capture.mu.Unlock()
		if !exists {
			http.Error(w, "HAR artifact not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="tab-%d.har"`, tabID))
		http.ServeFile(w, r, artifact.Path)
	default:
		w.Header().Set("Allow", "GET, PUT")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (c *controlServer) handleHARCapture(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Pragma", "no-cache")
	if r.Method != http.MethodDelete {
		w.Header().Set("Allow", "DELETE")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	sessionID := strings.TrimSpace(r.URL.Query().Get("session"))
	captureID := strings.TrimSpace(r.URL.Query().Get("capture"))
	token := strings.TrimSpace(r.URL.Query().Get("token"))
	sess, ok := c.registry.Get(sessionID)
	if !ok {
		c.writeSessionNotFound(w)
		return
	}
	capture, err := sess.resolveHARCapture(captureID, token)
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
	capture.mu.Lock()
	ended := capture.Ended
	ending := capture.Ending
	capture.mu.Unlock()
	if !ended {
		if ending {
			http.Error(w, "HAR capture is still ending", http.StatusConflict)
		} else {
			http.Error(w, "HAR capture has not ended", http.StatusConflict)
		}
		return
	}
	if err := sess.cleanupHARCapture(capture); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "capture_id": captureID})
}

func harResultTabCount(data map[string]any) int {
	if data == nil {
		return 0
	}
	if tabs, ok := data["tabs"].([]any); ok {
		return len(tabs)
	}
	return 0
}

func decorateHAREndResult(result *JobResult, capture *harCapture) error {
	if result.Data == nil {
		result.Data = make(map[string]any)
	}
	artifacts := capture.artifactList()
	expected := harResultTabCount(result.Data)
	if expected != len(artifacts) {
		return fmt.Errorf("HAR artifact count mismatch: extension reported %d tabs, daemon received %d files", expected, len(artifacts))
	}
	publicArtifacts := make([]map[string]any, 0, len(artifacts))
	for _, artifact := range artifacts {
		publicArtifacts = append(publicArtifacts, map[string]any{
			"tab_id": artifact.TabID,
			"size":   artifact.Size,
		})
	}
	capture.mu.Lock()
	startedAt := capture.StartedAt
	capture.mu.Unlock()
	result.Data["capture_id"] = capture.ID
	result.Data["artifact_token"] = capture.Token
	result.Data["artifacts"] = publicArtifacts
	result.Data["started_at"] = startedAt.Format(time.RFC3339Nano)
	result.Data["ended_at"] = time.Now().UTC().Format(time.RFC3339Nano)
	return nil
}
