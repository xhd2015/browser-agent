package browseragent

import "testing"

func TestChromeUnpackedExtensionID_knownPath(t *testing.T) {
	// Must match the user's Load-unpacked id for this install path.
	path := "/Users/xhd2015/.browser-agent/managed-chrome/extensions/browser-agent/1.0.7"
	got := ChromeUnpackedExtensionID(path)
	want := "abaolfdhbinmeeemkpckblhjcipmgceo"
	if got != want {
		t.Fatalf("ChromeUnpackedExtensionID(%q)=%q want %q", path, got, want)
	}
	// Trailing slash must not change id (Clean trims).
	got2 := ChromeUnpackedExtensionID(path + "/")
	if got2 != want {
		t.Fatalf("trailing slash: got %q want %q", got2, want)
	}
}

func TestChromeUnpackedExtensionID_empty(t *testing.T) {
	if ChromeUnpackedExtensionID("") != "" {
		t.Fatal("empty path should yield empty id")
	}
}
