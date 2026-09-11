package browseragent

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestWaitForExtensionConnection_AdaptiveStallNoProgress(t *testing.T) {
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		_ = json.NewEncoder(w).Encode(sessionSnapshot{
			SessionID: "sess-stall",
			Extension: sessionExtension{Connected: false},
		})
	}))
	defer srv.Close()

	var stderr bytes.Buffer
	start := time.Now()
	err := waitForExtensionConnection(t.TempDir(), srv.URL, "sess-stall", "chrome", "/tmp/ext", 2*time.Second, &stderr, false, func(string) error { return nil })
	elapsed := time.Since(start)
	if err != nil {
		t.Fatal(err)
	}
	if elapsed < 1500*time.Millisecond || elapsed > 4*time.Second {
		t.Fatalf("elapsed %v; want ~2s stall", elapsed)
	}
	out := stderr.String()
	if !strings.Contains(out, "adaptive") {
		t.Fatalf("missing adaptive banner:\n%s", out)
	}
	if !strings.Contains(out, "extension did not connect within") {
		t.Fatalf("missing timeout warning:\n%s", out)
	}
	if !strings.Contains(out, "install-chrome-extension") {
		t.Fatalf("missing install help:\n%s", out)
	}
	if hits.Load() < 2 {
		t.Fatalf("expected multiple polls, got %d", hits.Load())
	}
}

func TestWaitForExtensionConnection_StageProgressExtendsThenConnects(t *testing.T) {
	var n atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		i := int(n.Add(1))
		snap := sessionSnapshot{
			SessionID: "sess-prog",
			Extension: sessionExtension{Connected: false},
		}
		switch {
		case i == 1:
			snap.Attach = &sessionAttach{Stage: AttachStagePageOpen}
		case i == 2:
			snap.Attach = &sessionAttach{Stage: AttachStageRegisterSent, RegisterAttempts: 2, LastError: "Receiving end does not exist"}
		case i == 3:
			snap.Attach = &sessionAttach{Stage: AttachStageSWAck}
		default:
			snap.Extension = sessionExtension{
				Connected:            true,
				SupportsBrowserAgent: true,
				Version:              "1.0.16",
				Features:             []string{FeatureBrowserAgent},
			}
			snap.Attach = &sessionAttach{Stage: AttachStageReady}
		}
		_ = json.NewEncoder(w).Encode(snap)
	}))
	defer srv.Close()

	var stderr bytes.Buffer
	err := waitForExtensionConnection(t.TempDir(), srv.URL, "sess-prog", "chrome", "/tmp/ext", 5*time.Second, &stderr, false, func(string) error { return nil })
	if err != nil {
		t.Fatal(err)
	}
	out := stderr.String()
	if !strings.Contains(out, "Extension connected") {
		t.Fatalf("want connected:\n%s", out)
	}
	if !strings.Contains(out, "stage=page_open") || !strings.Contains(out, "stage=register_sent") {
		t.Fatalf("want stage progress lines:\n%s", out)
	}
}

func TestWaitForExtensionConnection_StalledWithStageDiagnosis(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(sessionSnapshot{
			SessionID: "sess-diag",
			Extension: sessionExtension{Connected: false},
			Attach: &sessionAttach{
				Stage:            AttachStageRegisterSent,
				LastError:        "Could not establish connection. Receiving end does not exist.",
				RegisterAttempts: 12,
			},
		})
	}))
	defer srv.Close()

	var stderr bytes.Buffer
	err := waitForExtensionConnection(t.TempDir(), srv.URL, "sess-diag", "chrome", "/tmp/ext", 1500*time.Millisecond, &stderr, false, func(string) error { return nil })
	if err != nil {
		t.Fatal(err)
	}
	out := stderr.String()
	if !strings.Contains(out, "stalled at stage=register_sent") {
		t.Fatalf("want stall diagnosis:\n%s", out)
	}
	if !strings.Contains(out, "Receiving end does not exist") {
		t.Fatalf("want last_error:\n%s", out)
	}
	if !strings.Contains(out, "register_attempts: 12") {
		t.Fatalf("want attempts:\n%s", out)
	}
	if !strings.Contains(out, "service worker looks asleep") {
		t.Fatalf("want Receiving-end SW wake hint:\n%s", out)
	}
	if !strings.Contains(out, "toolbar popup") {
		t.Fatalf("want toolbar popup hint:\n%s", out)
	}
}
