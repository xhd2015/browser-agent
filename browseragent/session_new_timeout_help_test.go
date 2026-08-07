package browseragent

import (
	"strings"
	"testing"
)

func TestFormatExtensionTimeoutHelp_Chrome(t *testing.T) {
	got := formatExtensionTimeoutHelp("chrome", "/tmp/ext/chrome", "sess-abc12")
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
}

func TestFormatExtensionTimeoutHelp_Firefox(t *testing.T) {
	got := formatExtensionTimeoutHelp("firefox", "/tmp/ext/ff", "sess-ff1")
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
	got := formatExtensionTimeoutHelp("", "", "")
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
