package browseragent

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestEnsureAsset_downloadAndNoRefetch(t *testing.T) {
	xdg := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", xdg)

	tarBytes := mustTarGZ(t, map[string]string{
		"index.html":             `<!doctype html><div id="root" data-browser-agent-root></div>`,
		"assets/session-page.js": `// js`,
	})

	var gets atomic.Int64
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			gets.Add(1)
		}
		if !strings.Contains(r.URL.Path, "v0.2.0") {
			t.Errorf("path missing version pin: %s", r.URL.Path)
		}
		if strings.Contains(r.URL.Path, "latest") {
			t.Errorf("path must not use latest: %s", r.URL.Path)
		}
		if !strings.HasSuffix(r.URL.Path, ".tar.gz") {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/gzip")
		_, _ = w.Write(tarBytes)
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	cfg := AssetDownloadConfig{
		BaseURL:    strings.TrimRight(srv.URL, "/") + "/releases/download",
		HTTPClient: srv.Client(),
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	dir1, err := EnsureAsset(ctx, "browser-agent", "v0.2.0", "session-page", cfg)
	if err != nil {
		t.Fatalf("EnsureAsset: %v", err)
	}
	if dir1 == "" {
		t.Fatal("empty dir")
	}
	if !CacheComplete("browser-agent", "v0.2.0", "session-page") {
		t.Fatal("cache incomplete after ensure")
	}
	if gets.Load() != 1 {
		t.Fatalf("gets=%d want 1", gets.Load())
	}

	dir2, err := EnsureAsset(ctx, "browser-agent", "v0.2.0", "session-page", cfg)
	if err != nil {
		t.Fatalf("second EnsureAsset: %v", err)
	}
	if filepath.Clean(dir2) != filepath.Clean(dir1) && !strings.HasPrefix(filepath.Clean(dir2), filepath.Clean(dir1)) {
		// allow equal paths only
		if filepath.Clean(dir2) != filepath.Clean(dir1) {
			t.Fatalf("dir2=%q dir1=%q", dir2, dir1)
		}
	}
	if gets.Load() != 1 {
		t.Fatalf("after second ensure gets=%d want 1", gets.Load())
	}
}

func TestEnsureAsset_404(t *testing.T) {
	xdg := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", xdg)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	t.Cleanup(srv.Close)

	cfg := AssetDownloadConfig{
		BaseURL:    strings.TrimRight(srv.URL, "/") + "/releases/download",
		HTTPClient: srv.Client(),
	}
	_, err := EnsureAsset(context.Background(), "browser-agent", "v0.2.0", "session-page", cfg)
	if err == nil {
		t.Fatal("want error on 404")
	}
	low := strings.ToLower(err.Error())
	if !strings.Contains(low, "404") && !strings.Contains(low, "download") && !strings.Contains(low, "not found") {
		t.Fatalf("error %q not download-like", err)
	}
	if CacheComplete("browser-agent", "v0.2.0", "session-page") {
		t.Fatal("must not complete cache on 404")
	}
}

func mustTarGZ(t *testing.T, files map[string]string) []byte {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	for name, body := range files {
		hdr := &tar.Header{
			Name: name,
			Mode: 0o644,
			Size: int64(len(body)),
		}
		if err := tw.WriteHeader(hdr); err != nil {
			t.Fatal(err)
		}
		if _, err := tw.Write([]byte(body)); err != nil {
			t.Fatal(err)
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}
