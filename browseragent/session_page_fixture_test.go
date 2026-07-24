package browseragent

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSessionPageDirIsMiniFixture_poller(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "assets"), 0o755); err != nil {
		t.Fatal(err)
	}
	index := `<!DOCTYPE html><html><body>
<div id="root"></div>
<details data-browser-agent-install><summary>Install</summary></details>
<script src="/assets/session-page.js"></script>
</body></html>`
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte(index), 0o644); err != nil {
		t.Fatal(err)
	}
	js := "// fixture session-page asset\n(function(){fetch('/v1/session');})();\n"
	if err := os.WriteFile(filepath.Join(dir, "assets", "session-page.js"), []byte(js), 0o644); err != nil {
		t.Fatal(err)
	}
	if !SessionPageDirIsMiniFixture(dir) {
		t.Fatal("poller fixture must be detected as mini fixture")
	}
}

func TestSessionPageDirIsMiniFixture_moduleSPA(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "assets"), 0o755); err != nil {
		t.Fatal(err)
	}
	// Large module entry like vite output.
	big := make([]byte, 10*1024)
	for i := range big {
		big[i] = 'a'
	}
	if err := os.WriteFile(filepath.Join(dir, "assets", "session-page-abc123.js"), big, 0o644); err != nil {
		t.Fatal(err)
	}
	index := `<!DOCTYPE html><html><head>
<script type="module" crossorigin src="/assets/session-page-abc123.js"></script>
</head><body><div id="root"></div></body></html>`
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte(index), 0o644); err != nil {
		t.Fatal(err)
	}
	if SessionPageDirIsMiniFixture(dir) {
		t.Fatal("vite-like SPA must not be treated as mini fixture")
	}
}
