package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/browser-agent/script/internal/clicolor"
)

func TestGeneratedVersionSinks_allowlist(t *testing.T) {
	if len(GeneratedVersionSinks) == 0 {
		t.Fatal("GeneratedVersionSinks must not be empty")
	}
	// Must not include SSoT or fat embed paths.
	forbidden := []string{
		"VERSION.txt",
		"browseragent/embedded/",
		"/build/",
		"dist/",
	}
	for _, rel := range GeneratedVersionSinks {
		if rel == "VERSION.txt" || strings.HasPrefix(rel, "/") {
			t.Fatalf("sink %q must not be root VERSION.txt or absolute", rel)
		}
		for _, bad := range forbidden {
			if bad == "VERSION.txt" {
				continue
			}
			if strings.Contains(rel, bad) {
				t.Fatalf("sink %q must not contain %q", rel, bad)
			}
		}
		// All sinks should be under known trees.
		ok := strings.HasPrefix(rel, "browseragent/") ||
			strings.HasPrefix(rel, "Chrome-Ext-") ||
			strings.HasPrefix(rel, "Firefox-Ext-")
		if !ok {
			t.Fatalf("unexpected sink prefix: %q", rel)
		}
	}
}

func TestHandle_rejectsCheckWithGitAdd(t *testing.T) {
	err := handle([]string{"--check", "--git-add-generated"}, clicolor.Style{})
	if err == nil {
		t.Fatal("expected error combining --check and --git-add-generated")
	}
	if !strings.Contains(err.Error(), "cannot combine") {
		t.Fatalf("error = %v", err)
	}
}

func TestExistingGeneratedSinks_onlyAllowlist(t *testing.T) {
	root := t.TempDir()
	// Create one allowlisted path and one non-allowlisted file.
	sink := GeneratedVersionSinks[0]
	if err := os.MkdirAll(filepath.Join(root, filepath.Dir(sink)), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, sink), []byte("1.0.0\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	noise := filepath.Join(root, "browseragent", "embedded", "extension", "background.js")
	if err := os.MkdirAll(filepath.Dir(noise), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(noise, []byte("// not a sink"), 0o644); err != nil {
		t.Fatal(err)
	}

	got := existingGeneratedSinks(root)
	if len(got) != 1 || got[0] != sink {
		t.Fatalf("existingGeneratedSinks = %v, want [%q]", got, sink)
	}
	for _, p := range got {
		if strings.Contains(p, "embedded/extension/background") {
			t.Fatal("must not include non-allowlisted embed paths")
		}
	}
}

func TestHelpText_documentsGitAdd(t *testing.T) {
	h := helpText()
	if !strings.Contains(h, "--git-add-generated") {
		t.Fatal("help must document --git-add-generated")
	}
	if !strings.Contains(h, "--check") {
		t.Fatal("help must document --check")
	}
	if !strings.Contains(h, "--color") || !strings.Contains(h, "--no-color") {
		t.Fatal("help must document color flags")
	}
	for _, sink := range GeneratedVersionSinks {
		if !strings.Contains(h, sink) {
			t.Fatalf("help should list sink %q", sink)
		}
	}
}
