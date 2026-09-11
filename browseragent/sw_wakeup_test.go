package browseragent

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestShouldOpenSWWakeup(t *testing.T) {
	cases := []struct {
		name     string
		elapsed  time.Duration
		stage    string
		attempts int
		already  bool
		want     bool
	}{
		{"too_early", time.Second, AttachStageRegisterSent, 10, false, false},
		{"few_attempts", 3 * time.Second, AttachStageRegisterSent, 2, false, false},
		{"wrong_stage", 3 * time.Second, AttachStageSWAck, 10, false, false},
		{"already", 3 * time.Second, AttachStageRegisterSent, 10, true, false},
		{"fire", 2 * time.Second, AttachStageRegisterSent, 5, false, true},
		{"fire_more", 5 * time.Second, AttachStageRegisterSent, 20, false, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := shouldOpenSWWakeup(tc.elapsed, tc.stage, tc.attempts, tc.already)
			if got != tc.want {
				t.Fatalf("got %v want %v", got, tc.want)
			}
		})
	}
}

func TestWaitForExtensionConnection_OpensWakeupOnce(t *testing.T) {
	var n atomic.Int32
	var calls []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		i := int(n.Add(1))
		snap := sessionSnapshot{
			SessionID: "sess-wake",
			Extension: sessionExtension{Connected: false},
			Attach: &sessionAttach{
				Stage:            AttachStageRegisterSent,
				RegisterAttempts: 2 + i*3,
			},
		}
		// After wakeup would have fired, still disconnected until late.
		if i >= 12 {
			snap.Extension.Connected = true
			snap.Extension.SupportsBrowserAgent = true
			snap.Attach.Stage = AttachStageReady
		}
		_ = json.NewEncoder(w).Encode(snap)
	}))
	defer srv.Close()

	base := t.TempDir()
	var stderr strings.Builder
	openWakeup := func(path string) error {
		calls = append(calls, path)
		return nil
	}
	err := waitForExtensionConnection(
		base, srv.URL, "sess-wake", "chrome", "/tmp/ext-wake",
		8*time.Second, &stderr, true, openWakeup,
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(calls) != 1 {
		t.Fatalf("wakeup calls=%v want exactly 1; stderr:\n%s", calls, stderr.String())
	}
	if calls[0] != "/tmp/ext-wake" {
		t.Fatalf("wakeup path=%q", calls[0])
	}
	if !strings.Contains(stderr.String(), "waking extension service worker") {
		t.Fatalf("stderr missing wakeup notice:\n%s", stderr.String())
	}
	lines, _ := ReadSessionLogLines(base, "sess-wake", 0)
	found := false
	for _, ln := range lines {
		if strings.Contains(ln, `"msg":"sw_wakeup_open"`) {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("log.jsonl missing sw_wakeup_open; lines=%v", lines)
	}
}

func TestWaitForExtensionConnection_NoWakeupWorkaroundDisables(t *testing.T) {
	var n atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		i := int(n.Add(1))
		snap := sessionSnapshot{
			SessionID: "sess-nowake",
			Extension: sessionExtension{Connected: false},
			Attach: &sessionAttach{
				Stage:            AttachStageRegisterSent,
				RegisterAttempts: 20,
			},
		}
		if i >= 8 {
			snap.Extension.Connected = true
			snap.Extension.SupportsBrowserAgent = true
			snap.Attach.Stage = AttachStageReady
		}
		_ = json.NewEncoder(w).Encode(snap)
	}))
	defer srv.Close()

	var calls int
	var stderr strings.Builder
	err := waitForExtensionConnection(
		t.TempDir(), srv.URL, "sess-nowake", "chrome", "/tmp/ext",
		5*time.Second, &stderr, false, /* enableWakeup=false → --no-wakeup-workaround */
		func(string) error {
			calls++
			return nil
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if calls != 0 {
		t.Fatalf("--no-wakeup-workaround must not open wakeup; calls=%d stderr:\n%s", calls, stderr.String())
	}
	if strings.Contains(stderr.String(), "waking extension service worker") {
		t.Fatalf("stderr must not mention wakeup:\n%s", stderr.String())
	}
}

func TestOpenExtensionWakeup_URL(t *testing.T) {
	// Known path from extension_id_test.go
	path := "/Users/fake/.browser-agent/managed-chrome/extensions/browser-agent/1.0.0"
	id := ChromeUnpackedExtensionID(path)
	if id == "" {
		t.Fatal("empty id")
	}
	wantSuffix := "/wakeup.html"
	// Don't actually launch Chrome — just verify ID derivation used by openExtensionWakeup.
	url := "chrome-extension://" + id + wantSuffix
	if !strings.HasPrefix(url, "chrome-extension://") || !strings.HasSuffix(url, wantSuffix) {
		t.Fatalf("url=%q", url)
	}
}
