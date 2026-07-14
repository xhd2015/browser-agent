package browseragent

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var wsUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// wsEnvelope is the versioned control-plane message.
type wsEnvelope struct {
	V       int            `json:"v"`
	Type    string         `json:"type"`
	ID      string         `json:"id"`
	Payload map[string]any `json:"payload"`
}

// wsConn wraps a single extension WebSocket with a write mutex.
type wsConn struct {
	conn *websocket.Conn
	mu   sync.Mutex
}

func (c *wsConn) writeJSON(v any) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	_ = c.conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	return c.conn.WriteJSON(v)
}

func (c *wsConn) close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	_ = c.conn.Close()
}

func (s *controlServer) resolveWSSession(r *http.Request) (*session, bool) {
	id := strings.TrimSpace(r.URL.Query().Get("session"))
	if id == "" {
		if only, ok := s.registry.onlySessionID(); ok {
			id = only
		}
	}
	if id == "" {
		return nil, false
	}
	sess, ok := s.registry.Get(id)
	return sess, ok
}

func (s *controlServer) handleWS(w http.ResponseWriter, r *http.Request) {
	sess, ok := s.resolveWSSession(r)
	if !ok {
		if strings.TrimSpace(r.URL.Query().Get("session")) == "" {
			s.writeMissingSession(w)
			return
		}
		s.writeSessionNotFound(w)
		return
	}

	conn, err := wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("browseragent: ws upgrade: %v", err)
		return
	}
	wc := &wsConn{conn: conn}
	// Install the new connection *before* closing the previous one so the old
	// handler's defer (getWS()==prev → markDisconnected) cannot wipe state after
	// we have already switched to the new socket.
	prev := sess.getWS()
	sess.setWS(wc)
	if prev != nil && prev != wc {
		prev.close()
	}

	defer func() {
		// Only clear if we are still the active conn.
		if sess.getWS() == wc {
			sess.markDisconnected()
		}
		wc.close()
	}()

	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			return
		}
		var env wsEnvelope
		if err := json.Unmarshal(data, &env); err != nil {
			continue
		}
		s.handleWSEnvelope(sess, wc, env)
	}
}

func (s *controlServer) handleWSEnvelope(sess *session, wc *wsConn, env wsEnvelope) {
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
		// bundle_md5 is optional; missing → md5_unknown match status.
		bundleMD5 := ""
		if v, ok := env.Payload["bundle_md5"].(string); ok {
			bundleMD5 = v
		} else if v, ok := env.Payload["bundleMd5"].(string); ok {
			bundleMD5 = v
		} else if v, ok := env.Payload["md5"].(string); ok {
			bundleMD5 = v
		}
		// Also accept []string via re-marshal if needed — features usually []any from JSON.
		sess.markHello(version, features, bundleMD5)
		browserProduct, _ := env.Payload["browser_product"].(string)
		pageCount := parseTelemetryPageCount(env.Payload["session_page_count"])
		pages := parseTelemetrySessionPages(env.Payload["session_pages"])
		sess.updateTelemetry(browserProduct, pageCount, pages)
		sess.setWS(wc)
		// Optional ack
		_ = wc.writeJSON(wsEnvelope{
			V:    1,
			Type: "status",
			ID:   env.ID,
			Payload: map[string]any{
				"ok":      true,
				"phase":   PhaseExtensionConnected,
				"session": sess.id,
			},
		})
		// Jobs enqueued while the extension was disconnected stay Queued and were
		// never pushed. Deliver them now so waiters do not only time out.
		s.repushQueuedJobs(sess)

	case "result":
		s.handleWSResult(sess, env)

	case "ping":
		_ = wc.writeJSON(wsEnvelope{V: 1, Type: "pong", ID: env.ID})

	case "status":
		if env.Payload == nil {
			return
		}
		browserProduct, _ := env.Payload["browser_product"].(string)
		pageCount := parseTelemetryPageCount(env.Payload["session_page_count"])
		pages := parseTelemetrySessionPages(env.Payload["session_pages"])
		if browserProduct != "" || pageCount != nil || pages != nil {
			sess.updateTelemetry(browserProduct, pageCount, pages)
		}
	}
}

func parseTelemetryPageCount(v any) *int {
	switch n := v.(type) {
	case float64:
		i := int(n)
		return &i
	case int:
		return &n
	case int64:
		i := int(n)
		return &i
	case json.Number:
		if i, err := n.Int64(); err == nil {
			v := int(i)
			return &v
		}
	}
	return nil
}

func parseTelemetrySessionPages(raw any) []sessionPageTab {
	arr, ok := raw.([]any)
	if !ok {
		return nil
	}
	out := make([]sessionPageTab, 0, len(arr))
	for _, item := range arr {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		tab := sessionPageTab{}
		switch id := m["tab_id"].(type) {
		case float64:
			tab.TabID = int(id)
		case int:
			tab.TabID = id
		}
		if url, ok := m["url"].(string); ok {
			tab.URL = url
		}
		if title, ok := m["title"].(string); ok {
			tab.Title = title
		}
		out = append(out, tab)
	}
	return out
}

// repushQueuedJobs delivers any jobs still in Queued status (push failed or
// never attempted while the extension was offline).
func (s *controlServer) repushQueuedJobs(sess *session) {
	for _, j := range sess.queue.SnapshotQueued() {
		_ = s.pushJob(sess, j)
	}
}

func (s *controlServer) handleWSResult(sess *session, env wsEnvelope) {
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
	_ = sess.queue.Complete(jobID, JobResult{
		JobID: jobID,
		OK:    ok,
		Error: errMsg,
		Data:  data,
	})
}

// pushJob sends a type=job envelope to the connected extension, if any.
// Returns true if a write was attempted successfully.
func (s *controlServer) pushJob(sess *session, j Job) bool {
	wc := sess.getWS()
	if wc == nil {
		return false
	}
	sess.queue.MarkRunning(j.ID)
	payload := map[string]any{
		"id":         j.ID,
		"job_id":     j.ID,
		"type":       j.Type,
		"params":     j.Params,
		"timeout_ms": j.TimeoutMS,
		"session_id": sess.id,
	}
	if j.TabID > 0 {
		payload["tab_id"] = j.TabID
	}
	env := wsEnvelope{
		V:    1,
		Type: "job",
		ID:   j.ID,
		Payload: payload,
	}
	if err := wc.writeJSON(env); err != nil {
		return false
	}
	return true
}