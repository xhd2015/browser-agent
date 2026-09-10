package browseragent

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"
)

// Defaults for POST /v1/ext/poll wait_ms.
const (
	DefaultExtPollWaitMS = 25000
	MaxExtPollWaitMS     = 30000
)

// NormalizeExtPollWaitMS returns DefaultExtPollWaitMS when waitMS <= 0,
// MaxExtPollWaitMS when waitMS > MaxExtPollWaitMS, otherwise waitMS.
func NormalizeExtPollWaitMS(waitMS int) int {
	if waitMS <= 0 {
		return DefaultExtPollWaitMS
	}
	if waitMS > MaxExtPollWaitMS {
		return MaxExtPollWaitMS
	}
	return waitMS
}

// --- request bodies ---

type extHelloRequest struct {
	SessionID      string   `json:"session_id"`
	Version        string   `json:"version"`
	Features       []string `json:"features"`
	BundleMD5      string   `json:"bundle_md5"`
	BundleMd5Camel string   `json:"bundleMd5"`
	MD5            string   `json:"md5"`
	BrowserProduct string   `json:"browser_product"`
}

type extPollRequest struct {
	SessionID string `json:"session_id"`
	WaitMS    int    `json:"wait_ms"`
}

type extResultRequest struct {
	SessionID string         `json:"session_id"`
	JobID     string         `json:"job_id"`
	ID        string         `json:"id"`
	OK        bool           `json:"ok"`
	Data      map[string]any `json:"data"`
	Error     string         `json:"error"`
}

func (c *controlServer) writeMissingSessionID(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"error":   "missing session_id",
		"message": "session_id is required",
	})
}

// handleExtAttach is POST /v1/ext/attach — page/content-script cold-start progress.
func (c *controlServer) handleExtAttach(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req AttachEvent
	if r.Body != nil {
		dec := json.NewDecoder(r.Body)
		if err := dec.Decode(&req); err != nil && err != io.EOF {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid json: " + err.Error()})
			return
		}
	}
	sid := strings.TrimSpace(req.SessionID)
	if sid == "" {
		c.writeMissingSessionID(w)
		return
	}
	sess, ok := c.registry.Get(sid)
	if !ok {
		c.writeSessionNotFound(w)
		return
	}
	sess.applyAttachEvent(req)
	snap := sess.snapshot()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"ok":         true,
		"session_id": sid,
		"attach":     snap.Attach,
	})
}

// handleExtHello is POST /v1/ext/hello — mark extension connected without WebSocket.
func (c *controlServer) handleExtHello(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req extHelloRequest
	if r.Body != nil {
		dec := json.NewDecoder(r.Body)
		if err := dec.Decode(&req); err != nil && err != io.EOF {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid json: " + err.Error()})
			return
		}
	}
	sid := strings.TrimSpace(req.SessionID)
	if sid == "" {
		c.writeMissingSessionID(w)
		return
	}
	sess, ok := c.registry.Get(sid)
	if !ok {
		c.writeSessionNotFound(w)
		return
	}

	bundleMD5 := strings.TrimSpace(req.BundleMD5)
	if bundleMD5 == "" {
		bundleMD5 = strings.TrimSpace(req.BundleMd5Camel)
	}
	if bundleMD5 == "" {
		bundleMD5 = strings.TrimSpace(req.MD5)
	}
	features := req.Features
	if features == nil {
		features = []string{}
	}
	sess.markHello(req.Version, features, bundleMD5)
	if req.BrowserProduct != "" {
		sess.updateTelemetry(req.BrowserProduct, nil, nil)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"ok":         true,
		"phase":      PhaseExtensionConnected,
		"session":    sid,
		"session_id": sid,
	})
}

// handleExtPoll is POST /v1/ext/poll — long-poll lease jobs + drain session events.
func (c *controlServer) handleExtPoll(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req extPollRequest
	if r.Body != nil {
		dec := json.NewDecoder(r.Body)
		if err := dec.Decode(&req); err != nil && err != io.EOF {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid json: " + err.Error()})
			return
		}
	}
	sid := strings.TrimSpace(req.SessionID)
	if sid == "" {
		c.writeMissingSessionID(w)
		return
	}
	sess, ok := c.registry.Get(sid)
	if !ok {
		c.writeSessionNotFound(w)
		return
	}

	// Normalize: <=0 → DefaultExtPollWaitMS (25000), >max → cap 30000.
	// Short positive waits (e.g. 100) pass through for empty-timeout tests.
	effectiveWait := NormalizeExtPollWaitMS(req.WaitMS)

	jobs, events := c.pollCollect(sess, effectiveWait)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"jobs":   jobs,
		"events": events,
	})
}

// pollCollect leases queued jobs and drains poll events, optionally waiting.
func (c *controlServer) pollCollect(sess *session, waitMS int) (jobs []map[string]any, events []map[string]any) {
	jobs, events = c.pollCollectNow(sess)
	if len(jobs) > 0 || len(events) > 0 || waitMS <= 0 {
		if jobs == nil {
			jobs = []map[string]any{}
		}
		if events == nil {
			events = []map[string]any{}
		}
		return jobs, events
	}

	deadline := time.Now().Add(time.Duration(waitMS) * time.Millisecond)
	wake := make(chan struct{}, 1)
	sess.addPollWaiter(wake)
	defer sess.removePollWaiter(wake)

	for {
		remaining := time.Until(deadline)
		if remaining <= 0 {
			break
		}
		timer := time.NewTimer(remaining)
		select {
		case <-wake:
			timer.Stop()
		case <-timer.C:
		}
		jobs, events = c.pollCollectNow(sess)
		if len(jobs) > 0 || len(events) > 0 {
			break
		}
		if !time.Now().Before(deadline) {
			break
		}
	}
	if jobs == nil {
		jobs = []map[string]any{}
	}
	if events == nil {
		events = []map[string]any{}
	}
	return jobs, events
}

func (c *controlServer) pollCollectNow(sess *session) (jobs []map[string]any, events []map[string]any) {
	jobs = make([]map[string]any, 0)
	for {
		j, ok := sess.queue.TryDequeue()
		if !ok {
			break
		}
		jobs = append(jobs, jobToPollJSON(sess.id, j))
	}
	events = sess.drainPollEvents()
	if events == nil {
		events = []map[string]any{}
	}
	return jobs, events
}

func jobToPollJSON(sessionID string, j Job) map[string]any {
	m := map[string]any{
		"id":         j.ID,
		"job_id":     j.ID,
		"type":       j.Type,
		"session_id": sessionID,
	}
	if j.Params != nil {
		m["params"] = j.Params
	} else {
		m["params"] = map[string]any{}
	}
	if j.TimeoutMS > 0 {
		m["timeout_ms"] = j.TimeoutMS
	}
	if j.TabID > 0 {
		m["tab_id"] = j.TabID
	}
	return m
}

// handleExtResult is POST /v1/ext/result — complete a job like WS type=result.
func (c *controlServer) handleExtResult(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req extResultRequest
	if r.Body != nil {
		dec := json.NewDecoder(r.Body)
		if err := dec.Decode(&req); err != nil && err != io.EOF {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid json: " + err.Error()})
			return
		}
	}
	sid := strings.TrimSpace(req.SessionID)
	if sid == "" {
		c.writeMissingSessionID(w)
		return
	}
	sess, ok := c.registry.Get(sid)
	if !ok {
		c.writeSessionNotFound(w)
		return
	}
	jobID := strings.TrimSpace(req.JobID)
	if jobID == "" {
		jobID = strings.TrimSpace(req.ID)
	}
	if jobID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error":   "missing job_id",
			"message": "job_id is required",
		})
		return
	}

	c.completeExtensionResult(sess, JobResult{
		JobID: jobID,
		OK:    req.OK,
		Error: req.Error,
		Data:  req.Data,
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"ok": true,
	})
}
