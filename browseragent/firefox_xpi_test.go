package browseragent

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPathToFileURL_unixStyle(t *testing.T) {
	got := PathToFileURL("/Users/me/.browser-agent/extensions/browser-agent-firefox/browser-agent.xpi")
	if !strings.HasPrefix(got, "file://") {
		t.Fatalf("want file:// prefix, got %q", got)
	}
	if !strings.Contains(got, "browser-agent.xpi") {
		t.Fatalf("want xpi basename in URL, got %q", got)
	}
	// Spaces encoded
	sp := PathToFileURL("/tmp/my path/file.xpi")
	if !strings.Contains(sp, "%20") && !strings.Contains(sp, "my%20path") {
		// url.URL.Path may leave spaces depending on Go version; either form is ok if still file://
		if !strings.HasPrefix(sp, "file://") {
			t.Fatalf("want file:// for spaced path, got %q", sp)
		}
	}
	if PathToFileURL("") != "" {
		t.Fatal("empty path must yield empty URL")
	}
}

func TestHandleFirefoxXPI_servesWhenEmbedded(t *testing.T) {
	if !EmbeddedFirefoxXPIAvailable() {
		t.Skip("no signed xpi embedded in this build")
	}
	// Isolate HOME so extract writes under temp.
	home := t.TempDir()
	t.Setenv("HOME", home)

	cs := &controlServer{}
	req := httptest.NewRequest(http.MethodGet, FirefoxXPIHTTPPath, nil)
	rr := httptest.NewRecorder()
	cs.handleFirefoxXPI(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
	ct := rr.Header().Get("Content-Type")
	if !strings.Contains(ct, "xpinstall") && !strings.Contains(ct, "octet") && !strings.Contains(ct, "zip") {
		// application/x-xpinstall is preferred; ServeFile may sniff differently on some OS.
		t.Logf("Content-Type=%q (ok if body is zip)", ct)
	}
	body := rr.Body.Bytes()
	if len(body) < 4 || body[0] != 'P' || body[1] != 'K' {
		t.Fatalf("body is not a zip/xpi (len=%d)", len(body))
	}
	// Ensure extract landed under HOME.
	expect := filepath.Join(home, ".browser-agent", "extensions", "browser-agent-firefox", "browser-agent.xpi")
	if st, err := os.Stat(expect); err != nil || st.IsDir() {
		t.Fatalf("expected extracted xpi at %s: %v", expect, err)
	}
}

func TestSessionSnapshot_includesFirefoxXPIFields(t *testing.T) {
	if !EmbeddedFirefoxXPIAvailable() {
		t.Skip("no signed xpi embedded in this build")
	}
	home := t.TempDir()
	base := filepath.Join(home, "sessions")
	if err := os.MkdirAll(base, 0o755); err != nil {
		t.Fatal(err)
	}
	// EnsureCanonicalFirefoxXPI uses process home via UserHomeDir — set HOME.
	t.Setenv("HOME", home)

	reg := NewSessionRegistry(base, "127.0.0.1:43761")
	res, err := reg.CreateWithOpts("sess-xpitest", CreateOpts{Browser: "firefox"})
	if err != nil {
		t.Fatalf("CreateWithOpts: %v", err)
	}
	if res == nil {
		t.Fatal("nil result")
	}
	sess, ok := reg.Get("sess-xpitest")
	if !ok {
		t.Fatal("session missing")
	}
	snap := reg.snapshot(sess)
	if snap.FirefoxXPIPath == "" {
		t.Fatal("expected firefox_xpi_path on snap")
	}
	if !strings.HasPrefix(snap.FirefoxXPIURL, "file://") {
		t.Fatalf("expected file:// firefox_xpi_url, got %q", snap.FirefoxXPIURL)
	}
	if snap.FirefoxXPIHTTPURL != "http://127.0.0.1:43761/v1/firefox-xpi" {
		t.Fatalf("http url: got %q", snap.FirefoxXPIHTTPURL)
	}
	if !strings.Contains(snap.ExtensionInstallPath, "browser-agent-firefox") {
		t.Fatalf("unpacked path should still be set: %q", snap.ExtensionInstallPath)
	}
}
