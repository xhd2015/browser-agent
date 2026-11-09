package browseragent

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

// directTestHarness spins a control server with one connected fake extension
// instance. Jobs posted without a session_id are answered by wsEchoJob, which
// mimics the extension: it acknowledges the job envelope and reports success
// with the payload it saw.
type directTestHarness struct {
	t        *testing.T
	server   *httptest.Server
	conn     *websocket.Conn
	incoming chan wsEnvelope
}

func newDirectTestHarness(t *testing.T, features []string) *directTestHarness {
	t.Helper()
	reg := NewSessionRegistry(t.TempDir(), "127.0.0.1:0")
	handler := NewRegistryControlHandler(reg)
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/v1/ws?instance=inst-test-1"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial instance ws: %v", err)
	}
	h := &directTestHarness{t: t, server: server, conn: conn, incoming: make(chan wsEnvelope, 16)}
	// Continuous reader: every envelope the daemon sends lands on incoming.
	go func() {
		for {
			var env wsEnvelope
			if err := conn.ReadJSON(&env); err != nil {
				close(h.incoming)
				return
			}
			h.incoming <- env
		}
	}()
	h.hello(features)
	t.Cleanup(func() { _ = conn.Close() })
	return h
}

func (h *directTestHarness) hello(features []string) {
	h.t.Helper()
	if err := h.conn.WriteJSON(wsEnvelope{V: 1, Type: "hello", ID: "hello-1", Payload: map[string]any{
		"version":  "1.0.19",
		"features": toAny(features),
	}}); err != nil {
		h.t.Fatalf("hello: %v", err)
	}
}

func toAny(values []string) []any {
	out := make([]any, 0, len(values))
	for _, v := range values {
		out = append(out, v)
	}
	return out
}

// replyResult writes a job result envelope for the given job id.
func (h *directTestHarness) replyResult(jobID string, ok bool, data map[string]any) {
	h.t.Helper()
	_ = h.conn.WriteJSON(wsEnvelope{V: 1, Type: "result", ID: jobID, Payload: map[string]any{
		"job_id": jobID,
		"ok":     ok,
		"data":   data,
	}})
}

func (h *directTestHarness) postJob(body string) (int, JobResult) {
	h.t.Helper()
	resp, err := http.Post(h.server.URL+"/v1/jobs", "application/json", bytes.NewReader([]byte(body)))
	if err != nil {
		h.t.Fatalf("post jobs: %v", err)
	}
	defer resp.Body.Close()
	var out struct {
		JobResult
		Error   string `json:"error"`
		Message string `json:"message"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&out)
	return resp.StatusCode, JobResult{JobID: out.JobID, OK: out.OK, Error: out.Error, Data: out.Data}
}

// drainAck consumes the hello status ack emitted during setup.
func (h *directTestHarness) drainAck() {
	h.t.Helper()
	select {
	case env := <-h.incoming:
		if env.Type != "status" {
			h.t.Fatalf("expected hello ack, got %q", env.Type)
		}
	case <-time.After(5 * time.Second):
		h.t.Fatal("no hello ack to drain")
	}
}

func TestDirectJobRoundTrip(t *testing.T) {
	h := newDirectTestHarness(t, []string{"browser-agent", DirectWindowFeature})
	h.drainAck()

	// The response blocks until the fake extension answers; serve the job.
	// Start the reader before posting: read the envelope from a goroutine and
	// answer from the test body.
	resultCh := make(chan JobResult, 1)
	statusCh := make(chan int, 1)
	go func() {
		status, res := h.postJob(`{"type":"screenshot","window_id":42,"tab_id":7,"timeout_ms":5000,"params":{"format":"jpeg"}}`)
		statusCh <- status
		resultCh <- res
	}()

	var job wsEnvelope
	select {
	case job = <-h.incoming:
	case <-time.After(5 * time.Second):
		t.Fatal("no job envelope arrived")
	}
	if job.Type != "job" {
		t.Fatalf("expected job envelope, got %q", job.Type)
	}
	if job.Payload["direct"] != true {
		t.Fatalf("expected direct=true, got %v", job.Payload["direct"])
	}
	if job.Payload["session_id"] != nil && job.Payload["session_id"] != "" {
		t.Fatalf("direct job must not carry session_id, got %v", job.Payload["session_id"])
	}
	if id, _ := job.Payload["window_id"].(float64); int64(id) != 42 {
		t.Fatalf("expected window_id 42, got %v", job.Payload["window_id"])
	}
	if id, _ := job.Payload["tab_id"].(float64); int64(id) != 7 {
		t.Fatalf("expected tab_id 7, got %v", job.Payload["tab_id"])
	}
	jobID, _ := job.Payload["job_id"].(string)
	h.replyResult(jobID, true, map[string]any{"type": "screenshot"})

	select {
	case status := <-statusCh:
		if status != http.StatusOK {
			t.Fatalf("status = %d", status)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("no response")
	}
	res := <-resultCh
	if !res.OK {
		t.Fatalf("expected ok result, got %+v", res)
	}
}

func TestDirectJobHelloAckAdvertisesServerFeature(t *testing.T) {
	h := newDirectTestHarness(t, []string{"browser-agent", DirectWindowFeature})
	// hello-1 was answered during harness setup; read the status ack.
	select {
	case env := <-h.incoming:
		if env.Type != "status" {
			t.Fatalf("expected status ack, got %q", env.Type)
		}
		features, ok := env.Payload["server_features"].([]any)
		if !ok || len(features) == 0 || features[0] != DirectWindowFeature {
			t.Fatalf("expected server_features [%s], got %v", DirectWindowFeature, env.Payload["server_features"])
		}
	case <-time.After(5 * time.Second):
		t.Fatal("no hello ack")
	}
}

func TestDirectJobWithoutInstanceFailsWithSessionWording(t *testing.T) {
	reg := NewSessionRegistry(t.TempDir(), "127.0.0.1:0")
	server := httptest.NewServer(NewRegistryControlHandler(reg))
	defer server.Close()
	resp, err := http.Post(server.URL+"/v1/jobs", "application/json", strings.NewReader(`{"type":"eval"}`))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", resp.StatusCode)
	}
	var out map[string]string
	_ = json.NewDecoder(resp.Body).Decode(&out)
	if out["error"] != "missing session_id" {
		t.Fatalf("error = %q", out["error"])
	}
}

func TestDirectJobRejectsHARAndUnknownInstance(t *testing.T) {
	h := newDirectTestHarness(t, []string{"browser-agent", DirectWindowFeature})
	h.drainAck()
	status, res := h.postJob(`{"type":"har_start"}`)
	if status != http.StatusOK || res.OK {
		t.Fatalf("har_start should fail gracefully, got status=%d res=%+v", status, res)
	}
	if !strings.Contains(res.Error, "session") {
		t.Fatalf("har error should mention session, got %q", res.Error)
	}

	// Named unknown instance → 503 with clear error.
	resp, err := http.Post(h.server.URL+"/v1/jobs?instance=nope", "application/json", strings.NewReader(`{"type":"eval"}`))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("unknown instance status = %d, want 503", resp.StatusCode)
	}
}

func TestDirectWindowsEndpoint(t *testing.T) {
	h := newDirectTestHarness(t, []string{"browser-agent", DirectWindowFeature})
	h.drainAck()
	resultCh := make(chan string, 1)
	go func() {
		resp, err := http.Get(h.server.URL + "/v1/windows")
		if err != nil {
			resultCh <- "ERR:" + err.Error()
			return
		}
		defer resp.Body.Close()
		body := new(bytes.Buffer)
		_, _ = body.ReadFrom(resp.Body)
		resultCh <- body.String()
	}()
	var job wsEnvelope
	select {
	case job = <-h.incoming:
	case <-time.After(5 * time.Second):
		t.Fatal("windows job not pushed")
	}
	jobID, _ := job.Payload["job_id"].(string)
	if job.Type != "job" {
		t.Fatalf("expected job, got %q", job.Type)
	}
	h.replyResult(jobID, true, map[string]any{
		"windows": []any{map[string]any{"window_id": 11.0, "title": "w"}},
	})
	select {
	case body := <-resultCh:
		if !strings.Contains(body, `"window_id":11`) {
			t.Fatalf("windows body missing window: %s", body)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("no windows response")
	}
}

func TestSessionJobsStillRouteWhenDirectPresent(t *testing.T) {
	h := newDirectTestHarness(t, []string{"browser-agent"})
	// A session-addressed job with no live session must still fail rather
	// than silently riding the direct link.
	resp, err := http.Post(h.server.URL+"/v1/jobs", "application/json", strings.NewReader(`{"type":"eval","session_id":"sess-none"}`))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404 (session rail untouched by direct mode)", resp.StatusCode)
	}
}

func TestDirectHealthReportsSupport(t *testing.T) {
	reg := NewSessionRegistry(t.TempDir(), "127.0.0.1:0")
	server := httptest.NewServer(NewRegistryControlHandler(reg))
	defer server.Close()
	resp, err := http.Get(server.URL + "/v1/health")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var out map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&out)
	if v, ok := out["direct_supported"].(bool); !ok || !v {
		t.Fatalf("direct_supported missing: %v", out)
	}
}

func TestDirectQueueFailsInflightAfterGrace(t *testing.T) {
	previous := directDisconnectGrace
	directDisconnectGrace = 30 * time.Millisecond
	t.Cleanup(func() { directDisconnectGrace = previous })

	h := newDirectTestHarness(t, []string{DirectWindowFeature})
	resultCh := make(chan JobResult, 1)
	go func() {
		status, res := h.postJob(`{"type":"screenshot","window_id":1,"timeout_ms":8000}`)
		if status != http.StatusOK {
			res = JobResult{OK: false, Error: "status"}
		}
		resultCh <- res
	}()
	<-h.incoming // job envelope
	// Drop the socket without answering; the grace period should fail the job.
	_ = h.conn.Close()
	select {
	case res := <-resultCh:
		if res.OK {
			t.Fatal("expected failure after disconnect")
		}
		if !strings.Contains(res.Error, "disconnected") {
			t.Fatalf("unexpected error: %q", res.Error)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("job did not fail after grace")
	}
}

func TestDirectContextWaitRespectsTimeout(t *testing.T) {
	h := newDirectTestHarness(t, []string{DirectWindowFeature})
	statusCh := make(chan int, 1)
	go func() {
		status, _ := h.postJob(`{"type":"screenshot","window_id":1,"timeout_ms":300}`)
		statusCh <- status
	}()
	<-h.incoming
	select {
	case <-statusCh:
	case <-time.After(5 * time.Second):
		t.Fatal("timeout job never returned")
	}
}
