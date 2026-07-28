package browseragent

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"sort"
	"strings"
)

// DefaultPrepareReconnectDelayMS is used when DelayMS <= 0.
const DefaultPrepareReconnectDelayMS = 1000

// PrepareReconnectOptions configures prepare_reconnect payload / broadcast.
type PrepareReconnectOptions struct {
	Reason      string // optional; e.g. "daemon-upgrade"
	DelayMS     int    // <=0 → DefaultPrepareReconnectDelayMS (1000)
	RetryBaseMS int    // optional; >0 → payload["retry_base_ms"]
	RetryMaxMS  int    // optional; >0 → payload["retry_max_ms"]
}

// BuildPrepareReconnectPayload returns the WS payload map for type=prepare_reconnect.
// Always includes delay_ms (int). Includes reason when non-empty.
// Includes retry_base_ms / retry_max_ms only when > 0.
func BuildPrepareReconnectPayload(opts PrepareReconnectOptions) map[string]any {
	delay := opts.DelayMS
	if delay <= 0 {
		delay = DefaultPrepareReconnectDelayMS
	}
	payload := map[string]any{
		"delay_ms": delay,
	}
	if strings.TrimSpace(opts.Reason) != "" {
		payload["reason"] = opts.Reason
	}
	if opts.RetryBaseMS > 0 {
		payload["retry_base_ms"] = opts.RetryBaseMS
	}
	if opts.RetryMaxMS > 0 {
		payload["retry_max_ms"] = opts.RetryMaxMS
	}
	return payload
}

// BroadcastPrepareReconnect notifies connected extensions of an impending
// control-plane reconnect (daemon upgrade, etc.):
//   - Sessions with a live WebSocket receive type=prepare_reconnect on the socket.
//   - HTTP-poll sessions (extension.connected, no WS) receive the same event on
//     their poll event queue so POST /v1/ext/poll can drain it.
// Returns sorted session ids that were notified (WS write ok and/or poll enqueued).
// delay_ms / reason / retry hints follow BuildPrepareReconnectPayload rules.
func BroadcastPrepareReconnect(r *SessionRegistry, opts PrepareReconnectOptions) []string {
	if r == nil {
		return nil
	}
	payload := BuildPrepareReconnectPayload(opts)

	type target struct {
		id       string
		wc       *wsConn
		sess     *session
		httpOnly bool
	}
	r.mu.RLock()
	targets := make([]target, 0, len(r.sessions))
	for id, sess := range r.sessions {
		if sess == nil {
			continue
		}
		wc := sess.getWS()
		if wc != nil {
			targets = append(targets, target{id: id, wc: wc, sess: sess})
			continue
		}
		// HTTP transport: connected via POST /v1/ext/hello without a socket.
		if sess.isExtensionConnected() {
			targets = append(targets, target{id: id, sess: sess, httpOnly: true})
		}
	}
	r.mu.RUnlock()

	notified := make([]string, 0, len(targets))
	for _, t := range targets {
		evID := newID("prepare")
		if t.httpOnly {
			if t.sess == nil {
				continue
			}
			// Copy payload so concurrent mutations of the map do not race.
			pl := make(map[string]any, len(payload))
			for k, v := range payload {
				pl[k] = v
			}
			t.sess.enqueuePollEvent(map[string]any{
				"type":    "prepare_reconnect",
				"id":      evID,
				"payload": pl,
			})
			notified = append(notified, t.id)
			continue
		}
		// Re-check: connection may have dropped between snapshot and write.
		if t.wc == nil {
			continue
		}
		env := wsEnvelope{
			V:       1,
			Type:    "prepare_reconnect",
			ID:      evID,
			Payload: payload,
		}
		if err := t.wc.writeJSON(env); err != nil {
			// Fall back to poll event queue if WS write fails but session is connected.
			if t.sess != nil && t.sess.isExtensionConnected() {
				pl := make(map[string]any, len(payload))
				for k, v := range payload {
					pl[k] = v
				}
				t.sess.enqueuePollEvent(map[string]any{
					"type":    "prepare_reconnect",
					"id":      evID,
					"payload": pl,
				})
				notified = append(notified, t.id)
			}
			continue
		}
		notified = append(notified, t.id)
	}
	sort.Strings(notified)
	return notified
}

type prepareUpgradeRequest struct {
	Reason      string `json:"reason"`
	DelayMS     int    `json:"delay_ms"`
	RetryBaseMS int    `json:"retry_base_ms"`
	RetryMaxMS  int    `json:"retry_max_ms"`
}

func (c *controlServer) handlePrepareUpgrade(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	opts := PrepareReconnectOptions{}
	if r.Body != nil {
		var body prepareUpgradeRequest
		dec := json.NewDecoder(r.Body)
		err := dec.Decode(&body)
		// Empty body / EOF is fine (defaults). Successful decode applies fields.
		// Invalid JSON still broadcasts with defaults (admin path is best-effort).
		if err == nil {
			opts.Reason = body.Reason
			opts.DelayMS = body.DelayMS
			opts.RetryBaseMS = body.RetryBaseMS
			opts.RetryMaxMS = body.RetryMaxMS
		} else if !errors.Is(err, io.EOF) {
			_ = err
		}
	}

	notified := BroadcastPrepareReconnect(c.registry, opts)
	if notified == nil {
		notified = []string{}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"ok":       true,
		"notified": notified,
	})
}
