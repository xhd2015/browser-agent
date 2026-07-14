package browseragent

import (
	"strings"
	"testing"
)

func TestIsSessionSPAHTML(t *testing.T) {
	spa := `<!DOCTYPE html><html><head>
<script type="module" crossorigin src="/assets/session-page-abc.js"></script>
</head><body><div id="root" data-browser-agent-root></div></body></html>`
	if !isSessionSPAHTML(spa) {
		t.Fatal("expected SPA detection for vite session-page asset")
	}
	plain := `<!DOCTYPE html><html><body><h1>browser-agent</h1></body></html>`
	if isSessionSPAHTML(plain) {
		t.Fatal("plain shell must not be detected as SPA")
	}
}

func TestInjectSessionBoot_SPA_noDuplicateIdentityOrInstall(t *testing.T) {
	spa := `<!DOCTYPE html><html><head>
<title>Browser Agent Session</title>
<script type="module" crossorigin src="/assets/session-page-8mAClDMn.js"></script>
</head>
<body>
  <div id="root" data-browser-agent-root></div>
</body></html>`
	snap := sessionSnapshot{
		SessionID: "sess-x",
		BundledExtension: bundledExtension{
			Version: "1.0.1",
			MD5:     "deadbeef",
			Path:    "/tmp/ext",
		},
		ExtensionMatch: ExtensionMatchNotConnected,
	}
	out := injectSessionBoot(spa, "sess-x", snap)
	if !strings.Contains(out, "browser-agent-boot") {
		t.Fatal("expected boot inject")
	}
	if !strings.Contains(out, "data-session-id") {
		t.Fatal("expected data-session-id on body")
	}
	// SPA must not get SSR identity/install panels (React owns them).
	if n := strings.Count(out, "data-browser-agent-ext-identity"); n != 0 {
		t.Fatalf("SPA inject must not add identity panel, got count=%d", n)
	}
	if strings.Contains(out, `id="browser-agent-install"`) {
		t.Fatal("SPA inject must not add install details panel")
	}
}

func TestInjectSessionBoot_nonSPA_addsIdentityOnce(t *testing.T) {
	plain := `<!DOCTYPE html><html><head><title>t</title></head>
<body><p>shell</p></body></html>`
	snap := sessionSnapshot{
		SessionID: "sess-y",
		BundledExtension: bundledExtension{
			Version: "1.0.1",
			MD5:     "cafebabe",
			Path:    "/tmp/ext2",
		},
		ExtensionMatch: ExtensionMatchNotConnected,
	}
	out := injectSessionBoot(plain, "sess-y", snap)
	if n := strings.Count(out, "data-browser-agent-ext-identity"); n != 1 {
		t.Fatalf("non-SPA inject wants exactly 1 identity panel, got %d", n)
	}
	if !strings.Contains(out, "cafebabe") {
		t.Fatal("expected bundled md5 in identity panel")
	}
	if !strings.Contains(strings.ToLower(out), "chrome://extensions") {
		t.Fatal("non-SPA inject should include install guidance")
	}
	// Second inject path: already has identity marker → still one.
	out2 := injectSessionBoot(out, "sess-y", snap)
	// Note: out2 re-injects boot into already-injected HTML; count identity still 1.
	if n := strings.Count(out2, "data-browser-agent-ext-identity"); n != 1 {
		t.Fatalf("re-inject should not duplicate identity, got %d", n)
	}
}
