package browseragent

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"
)

func writeTestXPI(t *testing.T, dir, manifestVersion, fileName string) string {
	t.Helper()
	path := filepath.Join(dir, fileName)
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(f)
	w, err := zw.Create("manifest.json")
	if err != nil {
		t.Fatal(err)
	}
	_, err = w.Write([]byte(`{"manifest_version":3,"name":"t","version":"` + manifestVersion + `"}`))
	if err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestPeekSignedXPIVersion_fromManifest(t *testing.T) {
	dir := t.TempDir()
	path := writeTestXPI(t, dir, "1.0.2", "browser-agent-1.0.0-signed.xpi")
	got, err := PeekSignedXPIVersion(path)
	if err != nil {
		t.Fatal(err)
	}
	if got != "1.0.2" {
		t.Fatalf("PeekSignedXPIVersion = %q, want 1.0.2 (manifest beats filename)", got)
	}
	if !SignedXPIMatchesProduct(path, "1.0.2") {
		t.Fatal("expected match for 1.0.2")
	}
	if SignedXPIMatchesProduct(path, "1.0.0") {
		t.Fatal("must not match product 1.0.0 when manifest is 1.0.2")
	}
}

func TestPeekSignedXPIVersion_filenameFallback(t *testing.T) {
	dir := t.TempDir()
	// ZIP without manifest.json
	path := filepath.Join(dir, "browser-agent-9.8.7-signed.xpi")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(f)
	w, err := zw.Create("readme.txt")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = w.Write([]byte("no manifest"))
	_ = zw.Close()
	_ = f.Close()

	got, err := PeekSignedXPIVersion(path)
	if err != nil {
		t.Fatal(err)
	}
	if got != "9.8.7" {
		t.Fatalf("filename fallback got %q", got)
	}
}

func TestEnsureEmbedPlaceholders_restoresMissing(t *testing.T) {
	root := t.TempDir()
	// Minimal module-like layout
	for _, rel := range EmbedPlaceholderRels {
		dir := filepath.Join(root, filepath.Dir(rel))
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	// Only create one placeholder; leave others missing
	if err := os.WriteFile(filepath.Join(root, EmbedPlaceholderRels[0]), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := EnsureEmbedPlaceholders(root); err != nil {
		t.Fatal(err)
	}
	for _, rel := range EmbedPlaceholderRels {
		if _, err := os.Stat(filepath.Join(root, rel)); err != nil {
			t.Fatalf("missing %s after ensure: %v", rel, err)
		}
	}
}

func TestRemoveEmbeddedFirefoxXPIIfMismatch(t *testing.T) {
	root := t.TempDir()
	xpiDir := filepath.Join(root, "browseragent", "embedded", "firefox-xpi")
	if err := os.MkdirAll(xpiDir, 0o755); err != nil {
		t.Fatal(err)
	}
	src := writeTestXPI(t, t.TempDir(), "1.0.0", "old.xpi")
	data, err := os.ReadFile(src)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(xpiDir, firefoxXPIFileName), data, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(xpiDir, firefoxXPIVersionFile), []byte("1.0.2\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := RemoveEmbeddedFirefoxXPIIfMismatch(root, "1.0.2"); err != nil {
		t.Fatal(err)
	}
	if IsRealFirefoxXPIFile(filepath.Join(xpiDir, firefoxXPIFileName)) {
		t.Fatal("mismatched xpi should be removed")
	}
	if _, err := os.Stat(filepath.Join(xpiDir, "placeholder.txt")); err != nil {
		t.Fatalf("placeholder should exist: %v", err)
	}
}

func TestStageFirefoxXPIEmbed_keepsPlaceholder(t *testing.T) {
	root := t.TempDir()
	// seed placeholder like git
	phDir := filepath.Join(root, "browseragent", "embedded", "firefox-xpi")
	if err := os.MkdirAll(phDir, 0o755); err != nil {
		t.Fatal(err)
	}
	ph := filepath.Join(phDir, "placeholder.txt")
	if err := os.WriteFile(ph, []byte("keep-me\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	src := writeTestXPI(t, t.TempDir(), "1.0.2", "browser-agent-1.0.2-signed.xpi")
	if _, err := StageFirefoxXPIEmbed(root, src, "1.0.2"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(ph); err != nil {
		t.Fatalf("placeholder must remain after stage: %v", err)
	}
	if !IsRealFirefoxXPIFile(filepath.Join(phDir, firefoxXPIFileName)) {
		t.Fatal("xpi should be staged")
	}
}
