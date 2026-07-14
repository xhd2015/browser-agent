package browseragent

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBuildExtensionShell_copiesPublicToBuild(t *testing.T) {
	root := t.TempDir()
	public := filepath.Join(root, "Chrome-Ext-Browser-Agent", "public")
	if err := os.MkdirAll(public, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(public, "manifest.json"), []byte(`{"version":"1.0.0","name":"t"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(public, "background.js"), []byte("// bg\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	buildDir, err := BuildExtensionShell(root)
	if err != nil {
		t.Fatalf("BuildExtensionShell: %v", err)
	}
	if _, err := os.Stat(filepath.Join(buildDir, "manifest.json")); err != nil {
		t.Fatalf("manifest in build: %v", err)
	}
	if _, err := os.Stat(filepath.Join(buildDir, "background.js")); err != nil {
		t.Fatalf("background.js in build: %v", err)
	}
}

func TestNormalizeSessionPageDist_promotesSessionPageHTML(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "session-page.html"), []byte(`<html id="root"></html>`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := normalizeSessionPageDist(dir); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(dir, "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	if string(b) == "" {
		t.Fatal("index.html empty")
	}
}
