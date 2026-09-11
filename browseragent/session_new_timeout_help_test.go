package browseragent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestFormatExtensionTimeoutHelp_Chrome(t *testing.T) {
	got := formatExtensionTimeoutHelp("chrome", "/tmp/ext/chrome", "sess-abc12", "")
	for _, want := range []string{
		extensionTimeoutUserHandling,
		"browser-agent install-chrome-extension",
		"chrome://extensions",
		"Load unpacked",
		"/tmp/ext/chrome",
		"session info --session-id sess-abc12",
		"do not session new again",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("chrome help missing %q; got:\n%s", want, got)
		}
	}
	if strings.Contains(got, "install-firefox-extension") {
		t.Fatalf("chrome help must not mention firefox install; got:\n%s", got)
	}
	if strings.Contains(got, "about:debugging") {
		t.Fatalf("chrome help must not mention firefox debugging; got:\n%s", got)
	}
	if strings.Contains(got, "service worker looks asleep") {
		t.Fatalf("plain chrome help must not mention SW asleep without Receiving end; got:\n%s", got)
	}
}

func TestFormatExtensionTimeoutHelp_ChromeReceivingEnd(t *testing.T) {
	got := formatExtensionTimeoutHelp(
		"chrome",
		"/tmp/ext/chrome",
		"sess-abc12",
		"Could not establish connection. Receiving end does not exist.",
	)
	for _, want := range []string{
		extensionTimeoutUserHandling,
		"service worker looks asleep",
		"Receiving end does not exist",
		"toolbar popup",
		"Reload",
		"--no-wakeup-workaround",
		"do not run session new again",
		"browser-agent install-chrome-extension",
		"session info --session-id sess-abc12",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("receiving-end help missing %q; got:\n%s", want, got)
		}
	}
}

func TestFormatExtensionTimeoutHelp_Firefox(t *testing.T) {
	got := formatExtensionTimeoutHelp("firefox", "/tmp/ext/ff", "sess-ff1", "")
	for _, want := range []string{
		extensionTimeoutUserHandling,
		"browser-agent install-firefox-extension",
		"about:debugging#/runtime/this-firefox",
		"Load Temporary Add-on",
		"/tmp/ext/ff",
		"session info --session-id sess-ff1",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("firefox help missing %q; got:\n%s", want, got)
		}
	}
	if strings.Contains(got, "install-chrome-extension") {
		t.Fatalf("firefox help must not mention chrome install; got:\n%s", got)
	}
	if strings.Contains(got, "chrome://extensions") {
		t.Fatalf("firefox help must not mention chrome://extensions; got:\n%s", got)
	}
}

func TestFormatExtensionTimeoutHelp_EmptyBrowserDefaultsChrome(t *testing.T) {
	got := formatExtensionTimeoutHelp("", "", "", "")
	if !strings.Contains(got, extensionTimeoutUserHandling) {
		t.Fatalf("missing user-handling banner; got:\n%s", got)
	}
	if !strings.Contains(got, "install-chrome-extension") {
		t.Fatalf("empty browser should default to chrome install; got:\n%s", got)
	}
	if !strings.Contains(got, "session info --session-id <session-id>") {
		t.Fatalf("empty session id should use placeholder; got:\n%s", got)
	}
}

func TestWarnIfManyWaitingSessions(t *testing.T) {
	n := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n++
		list := make([]sessionSnapshot, waitingExtensionWarnThreshold)
		for i := range list {
			list[i] = sessionSnapshot{
				SessionID: fmt.Sprintf("sess-w%d", i),
				Extension: sessionExtension{Connected: false},
			}
		}
		_ = json.NewEncoder(w).Encode(list)
	}))
	defer srv.Close()

	var stderr bytes.Buffer
	warnIfManyWaitingSessions(srv.URL, &stderr)
	out := stderr.String()
	if !strings.Contains(out, "already waiting for extension connection") {
		t.Fatalf("want waiting warning; got:\n%s", out)
	}
	if !strings.Contains(out, "toolbar popup") {
		t.Fatalf("want popup hint; got:\n%s", out)
	}
	if n != 1 {
		t.Fatalf("expected one GET /v1/sessions, got %d", n)
	}

	stderr.Reset()
	srv2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode([]sessionSnapshot{
			{SessionID: "sess-a", Extension: sessionExtension{Connected: false}},
			{SessionID: "sess-b", Extension: sessionExtension{Connected: true}},
		})
	}))
	defer srv2.Close()
	warnIfManyWaitingSessions(srv2.URL, &stderr)
	if stderr.Len() != 0 {
		t.Fatalf("below threshold must be silent; got:\n%s", stderr.String())
	}
}
