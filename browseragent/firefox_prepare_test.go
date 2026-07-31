package browseragent

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPrepareFirefoxExtensionBuild_writesBundleSum(t *testing.T) {
	root := t.TempDir()
	// Minimal Firefox public shell
	public := filepath.Join(root, "Firefox-Ext-Browser-Agent", "public")
	if err := os.MkdirAll(public, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(public, "manifest.json"), []byte(`{"manifest_version":3,"version":"9.9.9","name":"t","browser_specific_settings":{"gecko":{"id":"t@x"}}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "VERSION.txt"), []byte("1.2.3\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	buildDir, ver, err := PrepareFirefoxExtensionBuild(root)
	if err != nil {
		t.Fatal(err)
	}
	if ver != "1.2.3" {
		t.Fatalf("version=%q want 1.2.3", ver)
	}
	if _, err := os.Stat(filepath.Join(buildDir, "manifest.json")); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(buildDir, "bundle-sum.js")); err != nil {
		t.Fatalf("bundle-sum.js missing: %v", err)
	}

	pkg, err := StageFirefoxUnsignedPackage(root, buildDir, ver)
	if err != nil {
		t.Fatal(err)
	}
	if pkg.StageDir == "" {
		t.Fatal("empty StageDir")
	}
	if _, err := os.Stat(filepath.Join(pkg.StageDir, "manifest.json")); err != nil {
		t.Fatal(err)
	}
}
