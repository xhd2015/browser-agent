package browseragent

import (
	"encoding/json"
	"log"
	"net/http"
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

func (s *controlServer) handleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("browseragent: ws upgrade: %v", err)
		return
	}
	wc := &wsConn{conn: conn}
	// Install the new connection *before* closing the previous one so the old
	// handler's defer (getWS()==prev → markDisconnected) cannot wipe state after
	// we have already switched to the new socket.
	prev := s.sess.getWS()
	s.sess.setWS(wc)
	if prev != nil && prev != wc {
		prev.close()
	}

	defer func() {
		// Only clear if we are still the active conn.
		if s.sess.getWS() == wc {
			s.sess.markDisconnected()
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
		s.handleWSEnvelope(wc, env)
	}
}

func (s *controlServer) handleWSEnvelope(wc *wsConn, env wsEnvelope) {
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
		s.sess.markHello(version, features, bundleMD5)
		s.sess.setWS(wc)
		// Optional ack
		_ = wc.writeJSON(wsEnvelope{
			V:    1,
			Type: "status",
			ID:   env.ID,
			Payload: map[string]any{
				"ok":      true,
				"phase":   PhaseExtensionConnected,
				"session": s.sess.id,
			},
		})
		// Jobs enqueued while the extension was disconnected stay Queued and were
		// never pushed. Deliver them now so waiters do not only time out.
		s.repushQueuedJobs()

	case "result":
		s.handleWSResult(env)

	case "ping":
		_ = wc.writeJSON(wsEnvelope{V: 1, Type: "pong", ID: env.ID})

	case "status":
		// Extension status updates — ignore for MVP.
	}
}

// repushQueuedJobs delivers any jobs still in Queued status (push failed or
// never attempted while the extension was offline).
func (s *controlServer) repushQueuedJobs() {
	for _, j := range s.sess.queue.SnapshotQueued() {
		_ = s.pushJob(j)
	}
}

func (s *controlServer) handleWSResult(env wsEnvelope) {
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
	_ = s.sess.queue.Complete(jobID, JobResult{
		JobID: jobID,
		OK:    ok,
		Error: errMsg,
		Data:  data,
	})
}

// pushJob sends a type=job envelope to the connected extension, if any.
// Returns true if a write was attempted successfully.
func (s *controlServer) pushJob(j Job) bool {
	wc := s.sess.getWS()
	if wc == nil {
		return false
	}
	s.sess.queue.MarkRunning(j.ID)
	env := wsEnvelope{
		V:    1,
		Type: "job",
		ID:   j.ID,
		Payload: map[string]any{
			"id":         j.ID,
			"job_id":     j.ID,
			"type":       j.Type,
			"params":     j.Params,
			"timeout_ms": j.TimeoutMS,
			"session_id": s.sess.id,
		},
	}
	if err := wc.writeJSON(env); err != nil {
		return false
	}
	return true
}
