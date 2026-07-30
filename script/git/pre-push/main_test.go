package main

import (
	"strings"
	"testing"
)

func TestExtractBackgroundFallbackVersion(t *testing.T) {
	src := []byte(`
const EXT_VERSION =
  typeof BROWSER_AGENT_BUNDLE_VERSION === "string" && BROWSER_AGENT_BUNDLE_VERSION
    ? BROWSER_AGENT_BUNDLE_VERSION
    : "1.0.2";
`)
	got, err := extractBackgroundFallbackVersion(src)
	if err != nil {
		t.Fatal(err)
	}
	if got != "1.0.2" {
		t.Fatalf("got %q, want 1.0.2", got)
	}
}

func TestExtractBackgroundFallbackVersion_letFirefox(t *testing.T) {
	src := []byte(`
let EXT_VERSION =
  typeof BROWSER_AGENT_BUNDLE_VERSION === "string" && BROWSER_AGENT_BUNDLE_VERSION
    ? BROWSER_AGENT_BUNDLE_VERSION
    : "0.3.1";
`)
	got, err := extractBackgroundFallbackVersion(src)
	if err != nil {
		t.Fatal(err)
	}
	if got != "0.3.1" {
		t.Fatalf("got %q", got)
	}
}

func TestExtractBackgroundFallbackVersion_missing(t *testing.T) {
	_, err := extractBackgroundFallbackVersion([]byte(`const x = 1;`))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestHelpText(t *testing.T) {
	h := helpText()
	if !strings.Contains(h, "VERSION.txt") {
		t.Fatal("help should mention VERSION.txt")
	}
	if !strings.Contains(h, "cat-file") {
		t.Fatal("help should mention cat-file / HEAD")
	}
}
