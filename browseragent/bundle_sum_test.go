package browseragent

import (
	"os"
	"path/filepath"
	"testing"
)

func TestComputeExtensionContentMD5IgnoresRootPlaceholder(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "manifest.json"), []byte(`{"version":"1.0.21"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "background.js"), []byte("console.log('ok')\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	without, err := ComputeExtensionContentMD5(dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "placeholder.txt"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	with, err := ComputeExtensionContentMD5(dir)
	if err != nil {
		t.Fatal(err)
	}
	if without != with {
		t.Fatalf("root placeholder.txt changed digest: without=%s with=%s", without, with)
	}
}

func TestComputeExtensionContentMD5IncludesNestedPlaceholder(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "manifest.json"), []byte(`{"version":"1.0.21"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	without, err := ComputeExtensionContentMD5(dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "nested"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "nested", "placeholder.txt"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	with, err := ComputeExtensionContentMD5(dir)
	if err != nil {
		t.Fatal(err)
	}
	if without == with {
		t.Fatalf("nested placeholder.txt should be part of digest: both=%s", without)
	}
}
