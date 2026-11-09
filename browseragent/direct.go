package browseragent

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"
)

// DirectWindowFeature is advertised by extensions that can address windows and
// tabs directly (no session, no /go control page), and echoed by the daemon in
// the hello ack so both sides agree before a consumer relies on it.
const DirectWindowFeature = "direct-window-v1"

// directDisconnectGrace mirrors wsDisconnectGracePeriod for instance links.
var directDisconnectGrace = wsDisconnectGracePeriod

// directConn is one instance-level extension connection. It is strictly
// in-memory: no session directory, no HAR spool, nothing on disk. A consumer
// addresses jobs by window_id / tab_id; the extension resolves the target.
type directConn struct {
	mu                   sync.Mutex
	id                   string
	extVersion           string
	extFeatures          []string
	connected            bool
	connectedAt          time.Time
	lastSeenAt           time.Time
	queue                *JobQueue
	ws                   *wsConn
	disconnectGeneration uint64

	extensionResultMu   sync.Mutex
	extensionResultSeen map[string]struct{}
}

func newDirectConn(id string) *directConn {
	return &directConn{
		id:                  id,
		queue:               NewJobQueue(),
		extensionResultSeen: make(map[string]struct{}),
	}
}

func (d *directConn) getWS() *wsConn {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.ws
}

func (d *directConn) setWS(c *wsConn) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.ws = c
	if c != nil {
		d.disconnectGeneration++
	}
}

func (d *directConn) isExtensionConnected() bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.connected
}

func (d *directConn) snapshot() directConnView {
	d.mu.Lock()
	defer d.mu.Unlock()
	return directConnView{
		InstanceID:  d.id,
		Version:     d.extVersion,
		Features:    append([]string(nil), d.extFeatures...),
		Connected:   d.connected,
		ConnectedAt: d.connectedAt,
		LastSeenAt:  d.lastSeenAt,
	}
}

// directConnView is the JSON shape exposed by GET /v1/connections.
type directConnView struct {
	InstanceID  string    `json:"instance_id"`
	Version     string    `json:"version,omitempty"`
	Features    []string  `json:"features,omitempty"`
	Connected   bool      `json:"connected"`
	ConnectedAt time.Time `json:"connected_at,omitempty"`
	LastSeenAt  time.Time `json:"last_seen_at,omitempty"`
}

func (d *directConn) markHello(version string, features []string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.connected = true
	d.connectedAt = time.Now()
	d.lastSeenAt = d.connectedAt
	d.extVersion = version
	d.extFeatures = append([]string(nil), features...)
}

func (d *directConn) markDisconnectedAfterGrace(c *wsConn) {
	d.mu.Lock()
	if d.ws != c {
		d.mu.Unlock()
		return
	}
	d.ws = nil
	d.disconnectGeneration++
	generation := d.disconnectGeneration
	d.mu.Unlock()

	time.AfterFunc(directDisconnectGrace, func() {
		d.mu.Lock()
		if d.disconnectGeneration != generation || d.ws != nil {
			d.mu.Unlock()
			return
		}
		d.connected = false
		q := d.queue
		d.mu.Unlock()
		if q != nil {
			q.FailAllInflight("extension disconnected: reconnect grace period expired")
		}
	})
}

func (d *directConn) supportsDirectWindow() bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	for _, f := range d.extFeatures {
		if f == DirectWindowFeature {
			return true
		}
	}
	return false
}

// directRegistry owns every instance connection.
type directRegistry struct {
	mu   sync.Mutex
	conns map[string]*directConn
}

func newDirectRegistry() *directRegistry {
	return &directRegistry{conns: make(map[string]*directConn)}
}

func (r *directRegistry) getOrCreate(id string) *directConn {
	r.mu.Lock()
	defer r.mu.Unlock()
	conn, ok := r.conns[id]
	if !ok {
		conn = newDirectConn(id)
		r.conns[id] = conn
	}
	return conn
}

func (r *directRegistry) get(id string) (*directConn, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	conn, ok := r.conns[id]
	return conn, ok
}

// connected returns every connection whose extension is currently linked.
func (r *directRegistry) connected() []*directConn {
	r.mu.Lock()
	defer r.mu.Unlock()
	var out []*directConn
	for _, conn := range r.conns {
		if conn.isExtensionConnected() {
			out = append(out, conn)
		}
	}
	return out
}

// sole returns the only connected connection, for consumers that do not name
// an instance explicitly.
func (r *directRegistry) sole() (*directConn, bool) {
	conns := r.connected()
	if len(conns) != 1 {
		return nil, false
	}
	return conns[0], true
}

func (r *directRegistry) list() []directConnView {
	r.mu.Lock()
	defer r.mu.Unlock()
	ids := make([]string, 0, len(r.conns))
	for id := range r.conns {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	out := make([]directConnView, 0, len(ids))
	for _, id := range ids {
		out = append(out, r.conns[id].snapshot())
	}
	return out
}

// handleWSInstance upgrades an ?instance= connection. It never touches the
// session registry: an old or new consumer cannot hijack a session link by
// query-parameter choice.
func (s *controlServer) handleWSInstance(w http.ResponseWriter, r *http.Request, instanceID string) {
	if strings.TrimSpace(instanceID) == "" {
		s.writeMissingSession(w)
		return
	}
	conn := s.direct.getOrCreate(instanceID)
	wsConn2, err := wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("browseragent: direct ws upgrade: %v", err)
		return
	}
	wc := &wsConn{conn: wsConn2}
	prev := conn.getWS()
	conn.setWS(wc)
	if prev != nil && prev != wc {
		prev.close()
	}

	defer func() {
		if conn.getWS() == wc {
			conn.markDisconnectedAfterGrace(wc)
		}
		wc.close()
	}()

	for {
		_, data, err := wc.conn.ReadMessage()
		if err != nil {
			return
		}
		var env wsEnvelope
		if err := json.Unmarshal(data, &env); err != nil {
			continue
		}
		s.handleDirectEnvelope(conn, wc, env)
	}
}

func (s *controlServer) handleDirectEnvelope(conn *directConn, wc *wsConn, env wsEnvelope) {
	switch env.Type {
	case "hello":
		version, _ := env.Payload["version"].(string)
		var features []string
		if raw, ok := env.Payload["features"].([]any); ok {
			for _, f := range raw {
				if str, ok := f.(string); ok {
					features = append(features, str)
				}
			}
		}
		conn.markHello(version, features)
		_ = wc.writeJSON(wsEnvelope{
			V:    1,
			Type: "status",
			ID:   env.ID,
			Payload: map[string]any{
				"ok":              true,
				"phase":           PhaseExtensionConnected,
				"instance":        conn.id,
				"server_features": []string{DirectWindowFeature},
			},
		})
		for _, j := range conn.queue.SnapshotQueued() {
			_ = s.pushDirectJob(conn, j)
		}
	case "result":
		s.handleDirectResult(conn, wc, env)
	case "ping":
		_ = wc.writeJSON(wsEnvelope{V: 1, Type: "pong", ID: env.ID})
	}
}

func (s *controlServer) handleDirectResult(conn *directConn, wc *wsConn, env wsEnvelope) {
	payload := env.Payload
	if payload == nil {
		payload = map[string]any{}
	}
	jobID := env.ID
	if id, ok := payload["job_id"].(string); ok && id != "" {
		jobID = id
	} else if id, ok := payload["id"].(string); ok && id != "" {
		jobID = id
	}
	ok, _ := payload["ok"].(bool)
	errMsg, _ := payload["error"].(string)
	var data map[string]any
	if d, ok := payload["data"].(map[string]any); ok {
		data = d
	}
	if debugEnabled() {
		if ok {
			debugJobLog("direct result instance=%s job_id=%s ok=true %s", conn.id, jobID, resultDataSummary(data))
		} else {
			debugJobLog("direct result instance=%s job_id=%s ok=false err=%q %s", conn.id, jobID, errMsg, resultDataSummary(data))
		}
	}
	ack := completeDirectResult(conn, JobResult{JobID: jobID, OK: ok, Error: errMsg, Data: data})
	if wc != nil && ack {
		_ = wc.writeJSON(wsEnvelope{
			V:       1,
			Type:    "result_ack",
			ID:      jobID,
			Payload: map[string]any{"job_id": jobID},
		})
	}
}

func completeDirectResult(conn *directConn, result JobResult) bool {
	conn.extensionResultMu.Lock()
	defer conn.extensionResultMu.Unlock()
	if _, seen := conn.extensionResultSeen[result.JobID]; seen {
		return true
	}
	if _, ok := conn.queue.Get(result.JobID); !ok {
		return false
	}
	if _, err := conn.queue.CompleteWithPrevious(result.JobID, result); err != nil {
		return false
	}
	conn.extensionResultSeen[result.JobID] = struct{}{}
	return true
}

// pushDirectJob sends a type=job envelope over the instance link. Direct jobs
// carry window_id/tab_id and no session identity.
func (s *controlServer) pushDirectJob(conn *directConn, j Job) bool {
	wc := conn.getWS()
	if wc == nil {
		if shouldAlwaysLogJob(j.Type, j.Params) {
			debugJobLog("pushDirectJob skip id=%s instance=%s type=%s reason=no_ws", j.ID, conn.id, j.Type)
		}
		return false
	}
	payload := map[string]any{
		"id":         j.ID,
		"job_id":     j.ID,
		"type":       j.Type,
		"params":     j.Params,
		"timeout_ms": j.TimeoutMS,
		"direct":     true,
	}
	if j.WindowID > 0 {
		payload["window_id"] = j.WindowID
	}
	if j.TabID > 0 {
		payload["tab_id"] = j.TabID
	}
	env := wsEnvelope{V: 1, Type: "job", ID: j.ID, Payload: payload}
	if err := wc.writeJSON(env); err != nil {
		if shouldAlwaysLogJob(j.Type, j.Params) {
			debugJobLog("pushDirectJob write_error id=%s instance=%s type=%s err=%v", j.ID, conn.id, j.Type, err)
		}
		return false
	}
	conn.queue.MarkRunning(j.ID)
	return true
}

// runDirectJob enqueues and awaits one job on the instance link. Shared by
// POST /v1/jobs (no session_id) and GET /v1/windows.
func (s *controlServer) runDirectJob(r *http.Request, conn *directConn, jobType string, windowID, tabID int64, params map[string]any, timeoutMS int64) JobResult {
	if timeoutMS <= 0 {
		timeoutMS = 30000
	}
	enqueued, err := conn.queue.Enqueue(Job{
		Type:      jobType,
		WindowID:  windowID,
		TabID:     tabID,
		Params:    params,
		TimeoutMS: timeoutMS,
	})
	if err != nil {
		return JobResult{OK: false, Error: err.Error()}
	}
	if shouldAlwaysLogJob(jobType, params) {
		debugJobLog("direct job enqueue id=%s instance=%s type=%s window_id=%d tab_id=%d timeout_ms=%d %s",
			enqueued.ID, conn.id, jobType, windowID, tabID, timeoutMS, jobParamsSummary(jobType, params))
	}
	pushed := s.pushDirectJob(conn, enqueued)
	if pushed {
		if shouldAlwaysLogJob(jobType, params) {
			debugJobLog("direct job push_ok id=%s instance=%s type=%s", enqueued.ID, conn.id, jobType)
		}
	} else if shouldAlwaysLogJob(jobType, params) {
		debugJobLog("direct job push_fail id=%s instance=%s type=%s reason=no_ws_or_write_error", enqueued.ID, conn.id, jobType)
	}
	ctx, cancel := context.WithTimeout(r.Context(), time.Duration(timeoutMS)*time.Millisecond)
	defer cancel()
	res, waitErr := conn.queue.Wait(ctx, enqueued.ID)
	if waitErr != nil && res.Error == "" {
		if ctx.Err() != nil {
			res = JobResult{JobID: enqueued.ID, OK: false, Error: "timeout waiting for job result"}
		} else {
			res = JobResult{JobID: enqueued.ID, OK: false, Error: waitErr.Error()}
		}
	}
	if res.JobID == "" {
		res.JobID = enqueued.ID
	}
	return res
}

// handleWindows lists Chrome windows through the instance link.
func (s *controlServer) handleWindows(w http.ResponseWriter, r *http.Request) {
	conn, errResp := s.resolveDirectConn(r)
	if errResp != "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": errResp})
		return
	}
	res := s.runDirectJob(r, conn, "windows", 0, 0, map[string]any{}, 15000)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"ok":      res.OK,
		"error":   res.Error,
		"windows": res.Data["windows"],
	})
}

// handleConnections reports instance links for consumers and diagnostics.
func (s *controlServer) handleConnections(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{"connections": s.direct.list()})
}

// resolveDirectConn picks the connection a direct job should ride: the named
// instance when given, otherwise the only connected instance.
func (s *controlServer) resolveDirectConn(r *http.Request) (*directConn, string) {
	if id := strings.TrimSpace(r.URL.Query().Get("instance")); id != "" {
		conn, ok := s.direct.get(id)
		if !ok {
			return nil, "unknown instance: " + id
		}
		if !conn.isExtensionConnected() {
			return nil, "extension not connected"
		}
		return conn, ""
	}
	conn, ok := s.direct.sole()
	if !ok {
		if len(s.direct.connected()) > 1 {
			return nil, "multiple extension instances connected; pass instance_id"
		}
		return nil, "extension not connected"
	}
	return conn, ""
}
