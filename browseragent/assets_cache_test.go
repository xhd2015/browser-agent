package browseragent

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
)

func TestAssetCacheRoot_XDG(t *testing.T) {
	xdg := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", xdg)
	root := AssetCacheRoot()
	want := filepath.Join(xdg, "browser-agent", "asset-cache")
	if filepath.Clean(root) != filepath.Clean(want) {
		t.Fatalf("AssetCacheRoot=%q want %q", root, want)
	}
}

func TestAssetCacheWriteOpenComplete(t *testing.T) {
	xdg := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", xdg)

	src := fstest.MapFS{
		"index.html":             &fstest.MapFile{Data: []byte(`<!doctype html><div id="root" data-browser-agent-root></div>`)},
		"assets/session-page.js": &fstest.MapFile{Data: []byte(`// js`)},
	}
	dir, err := WriteAssetCache("browser-agent", "v0.2.0", "session-page", src)
	if err != nil {
		t.Fatalf("WriteAssetCache: %v", err)
	}
	if !strings.Contains(filepath.ToSlash(dir), "browser-agent") ||
		!strings.Contains(filepath.ToSlash(dir), "v0.2.0") ||
		!strings.Contains(filepath.ToSlash(dir), "session-page") {
		t.Fatalf("WriteDir path missing segments: %s", dir)
	}
	if _, err := os.Stat(filepath.Join(dir, "index.html")); err != nil {
		t.Fatalf("index.html missing: %v", err)
	}
	if !CacheComplete("browser-agent", "v0.2.0", "session-page") {
		t.Fatal("CacheComplete false after write")
	}
	fsys, odir, ok, oerr := OpenAssetCache("browser-agent", "v0.2.0", "session-page")
	if oerr != nil || !ok || fsys == nil {
		t.Fatalf("OpenAssetCache ok=%v err=%v fsys=%v dir=%s", ok, oerr, fsys, odir)
	}
	if filepath.Clean(odir) != filepath.Clean(dir) {
		t.Fatalf("OpenDir=%q WriteDir=%q", odir, dir)
	}
	// product isolation
	if CacheComplete("browser-trace", "v0.2.0", "session-page") {
		t.Fatal("browser-trace should miss")
	}
	_, _, okOther, _ := OpenAssetCache("browser-trace", "v0.2.0", "session-page")
	if okOther {
		t.Fatal("browser-trace open should miss")
	}
}

func TestAssetCacheMiss(t *testing.T) {
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	if CacheComplete("browser-agent", "v0.2.0", "session-page") {
		t.Fatal("expected miss")
	}
	_, _, ok, err := OpenAssetCache("browser-agent", "v0.2.0", "session-page")
	if err != nil {
		t.Fatalf("open err: %v", err)
	}
	if ok {
		t.Fatal("expected ok=false")
	}
}
