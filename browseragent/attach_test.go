package browseragent

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestAttachStageRankMonotonic(t *testing.T) {
	stages := []string{
		AttachStageNoPage,
		AttachStagePageOpen,
		AttachStageRegisterSent,
		AttachStageSWAck,
		AttachStageWSConnecting,
		AttachStageWSOpen,
		AttachStageHello,
		AttachStageReady,
	}
	prev := -1
	for _, s := range stages {
		r := attachStageRank(s)
		if r <= prev {
			t.Fatalf("stage %s rank %d not > prev %d", s, r, prev)
		}
		prev = r
	}
	if attachStageRank("bogus") != -1 {
		t.Fatal("unknown stage should rank -1")
	}
}

func TestApplyAttachEventForwardOnly(t *testing.T) {
	s := newSession("sess-attach", t.TempDir())
	s.applyAttachEvent(AttachEvent{SessionID: "sess-attach", Stage: AttachStageRegisterSent, RegisterAttempts: 2, LastError: "Receiving end does not exist"})
	snap := s.snapshot()
	if snap.Attach == nil || snap.Attach.Stage != AttachStageRegisterSent {
		t.Fatalf("want register_sent, got %+v", snap.Attach)
	}
	if snap.Attach.LastError == "" || snap.Attach.RegisterAttempts != 2 {
		t.Fatalf("want error+attempts, got %+v", snap.Attach)
	}

	// Cannot move backward.
	s.applyAttachEvent(AttachEvent{SessionID: "sess-attach", Stage: AttachStagePageOpen})
	if s.snapshot().Attach.Stage != AttachStageRegisterSent {
		t.Fatalf("stage moved backward: %s", s.snapshot().Attach.Stage)
	}

	// SW ack clears last_error.
	s.applyAttachEvent(AttachEvent{SessionID: "sess-attach", Stage: AttachStageSWAck})
	got := s.snapshot().Attach
	if got.Stage != AttachStageSWAck || got.LastError != "" {
		t.Fatalf("sw_ack should clear last_error, got %+v", got)
	}
}

func TestAttachDeadlineExplicitShrinksBase(t *testing.T) {
	base, grant, cap := attachDeadline(3 * time.Second)
	if cap != 3*time.Second {
		t.Fatalf("cap=%v want 3s", cap)
	}
	if base != 3*time.Second {
		t.Fatalf("base=%v want 3s (shrunk to cap)", base)
	}
	if grant != DefaultAttachWaitGrant {
		t.Fatalf("grant=%v", grant)
	}
	base, _, cap = attachDeadline(0)
	if base != DefaultAttachWaitBase || cap != DefaultAttachWaitCap {
		t.Fatalf("defaults base=%v cap=%v", base, cap)
	}
}

func TestHandleExtAttachAndSnapshot(t *testing.T) {
	reg := NewSessionRegistry(t.TempDir(), "http://127.0.0.1:43761")
	if _, err := reg.Create("sess-att1"); err != nil {
		t.Fatal(err)
	}
	sess, ok := reg.Get("sess-att1")
	if !ok {
		t.Fatal("session missing after Create")
	}
	c := &controlServer{registry: reg}
	body := `{"session_id":"sess-att1","stage":"register_sent","last_error":"Receiving end does not exist","register_attempts":4}`
	req := httptest.NewRequest(http.MethodPost, "/v1/ext/attach", strings.NewReader(body))
	rr := httptest.NewRecorder()
	c.handleExtAttach(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status %d body %s", rr.Code, rr.Body.String())
	}
	var resp map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp["ok"] != true {
		t.Fatalf("resp=%v", resp)
	}
	snap := sess.snapshot()
	if snap.Attach == nil || snap.Attach.Stage != AttachStageRegisterSent {
		t.Fatalf("snapshot attach=%+v", snap.Attach)
	}
	if snap.Attach.RegisterAttempts != 4 {
		t.Fatalf("attempts=%d", snap.Attach.RegisterAttempts)
	}
}

func TestMarkHelloSetsAttachReady(t *testing.T) {
	s := newSession("sess-h", t.TempDir())
	s.markHello("1.0.16", []string{FeatureBrowserAgent}, "abc")
	snap := s.snapshot()
	if !snap.Extension.Connected || !snap.Extension.SupportsBrowserAgent {
		t.Fatalf("extension=%+v", snap.Extension)
	}
	if snap.Attach == nil || snap.Attach.Stage != AttachStageReady {
		t.Fatalf("attach=%+v want ready", snap.Attach)
	}
}

func TestInjectSessionBootReportsAttach(t *testing.T) {
	plain := `<!DOCTYPE html><html><head><title>t</title></head><body><p>shell</p></body></html>`
	snap := sessionSnapshot{
		SessionID: "sess-z",
		BundledExtension: bundledExtension{
			Version: "1.0.16",
			MD5:     "deadbeef",
			Path:    "/tmp/ext",
		},
		ExtensionInstallPath: "/tmp/ext",
		ExtensionMatch:       ExtensionMatchNotConnected,
	}
	out := injectSessionBoot(plain, "sess-z", snap)
	for _, want := range []string{
		"/v1/ext/attach",
		"reportAttach",
		"page_open",
		"register_sent",
		"sw_ack",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("inject missing %q", want)
		}
	}
	// Uncapped burst: no "n >= 40" early stop.
	if strings.Contains(out, "n >= 40") || strings.Contains(out, "burstAttempt >= 40") {
		t.Fatal("burst must continue until connected, not stop at 40")
	}
}
