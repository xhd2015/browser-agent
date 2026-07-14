# Scenario

**Feature**: embed Chrome-Ext-Capture-API (or mini fixture), extract, install UX

```
# Bundle stages MV3 into go:embed (production or mini fixture — no webpack in CI)
Bundle Tool -> browsertrace/embedded/extension/** -> //go:embed

# Extractor writes stable path under BaseDir
browser-trace -> Extractor -> {BaseDir}/extension/{version}/

# Surfaces under test
User -> --install-chrome-extension -> stdout path + Load unpacked + chrome://extensions + \n
Chrome Launcher arg builder -> --load-extension=<path> (no --user-data-dir)
Test Client -> GET /v1/session | GET /go -> install path + install panel
```

## Preconditions

- Module path `github.com/xhd2015/browser-agent` is the workspace root.
- Package `browsertrace` exports:
  - `ExtractEmbeddedExtension(baseDir string) (installPath, version string, err error)`
  - `BuildChromeLaunchArgs(sessionURL, extensionPath string) []string`
  - `InstallChromeExtension(w io.Writer, baseDir string) error`
  - existing `Config` + `Run` for HTTP modes (extract runs on session start)
- Embedded payload is a valid mini MV3 (or production stage) with `manifest.json`
  containing `"version"`. See `testdata/mini-extension/` for the CI shape.
- Each leaf uses an isolated temp `BaseDir`.
- No real Chrome process; no webpack/npm in this tree.
- HTTP modes use free loopback ports and short ready/complete timeouts.

## Steps

1. Allocate a unique temp `BaseDir` for the leaf.
2. Leave `Mode` unset; grouping Setup sets the surface.
3. Default `ExtractPasses = 1` when extract mode is selected later.
4. For HTTP descendants: `NoOpenChrome = true`, short timeouts, random SessionSuffix.

## Context

- Parallel-safe: temp dirs + free ports per leaf.
- `DOCTEST_SESSION_ID` is available if implementers later cache a built CLI binary;
  this tree prefers package APIs over `go build` of `cmd/browser-trace`.
- Shared helpers below are available to all descendant Assert/Setup packages.
- Fixture path relative to this tree root: `testdata/mini-extension/`
  (reference shape; product embed lives under package `browsertrace`).

```go
import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	dir := t.TempDir()
	req.BaseDir = filepath.Join(dir, "browser-trace-base")
	if err := os.MkdirAll(req.BaseDir, 0o755); err != nil {
		return err
	}
	req.NoOpenChrome = true
	if req.ExtractPasses == 0 {
		req.ExtractPasses = 1
	}
	if req.ReadyTimeout == 0 {
		req.ReadyTimeout = 5 * time.Second
	}
	if req.CompleteTimeout == 0 {
		req.CompleteTimeout = 2 * time.Second
	}
	if req.SessionSuffix == "" {
		req.SessionSuffix = fmt.Sprintf("emb-%d", time.Now().UnixNano()%1e12)
	}
	return nil
}

func assertExitZero(t *testing.T, resp *Response) {
	t.Helper()
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if resp.ExitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr=%q err=%q stdout=%q",
			resp.ExitCode, resp.Stderr, resp.ErrText, resp.Stdout)
	}
}

func assertStdoutTrailingNewline(t *testing.T, stdout string) {
	t.Helper()
	if stdout == "" {
		t.Fatal("stdout is empty")
	}
	if !strings.HasSuffix(stdout, "\n") {
		t.Fatalf("stdout must end with \\n (POSIX); last bytes=%q", tail(stdout, 40))
	}
}

func assertContainsFold(t *testing.T, haystack string, needles ...string) {
	t.Helper()
	low := strings.ToLower(haystack)
	for _, n := range needles {
		if !strings.Contains(low, strings.ToLower(n)) {
			t.Fatalf("expected text to contain %q; got:\n%s", n, truncate(haystack, 800))
		}
	}
}

func assertHTTPStatus(t *testing.T, resp *Response, want int) {
	t.Helper()
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if resp.StatusCode != want {
		t.Fatalf("HTTP status = %d, want %d; content-type=%q body=%s",
			resp.StatusCode, want, resp.ContentType, truncate(resp.BodyString, 400))
	}
}

func assertJSONContentType(t *testing.T, resp *Response) {
	t.Helper()
	ct := strings.ToLower(resp.ContentType)
	if !strings.Contains(ct, "json") {
		t.Fatalf("Content-Type %q does not look like JSON; body=%s",
			resp.ContentType, truncate(resp.BodyString, 200))
	}
}

func assertHTMLContentType(t *testing.T, resp *Response) {
	t.Helper()
	ct := strings.ToLower(resp.ContentType)
	if !strings.Contains(ct, "html") {
		t.Fatalf("Content-Type %q does not look like HTML; body=%s",
			resp.ContentType, truncate(resp.BodyString, 200))
	}
}

func assertExtractLayout(t *testing.T, req *Request, resp *Response) {
	t.Helper()
	if resp.InstallPath == "" {
		t.Fatal("InstallPath is empty")
	}
	if !filepath.IsAbs(resp.InstallPath) {
		t.Fatalf("InstallPath %q is not absolute", resp.InstallPath)
	}
	// Canonical layout: {BaseDir}/extension/{version}
	rel, err := filepath.Rel(req.BaseDir, resp.InstallPath)
	if err != nil {
		t.Fatal(err)
	}
	parts := strings.Split(rel, string(filepath.Separator))
	if len(parts) < 2 || parts[0] != "extension" {
		t.Fatalf("InstallPath %q not under {BaseDir}/extension/… (rel=%q)", resp.InstallPath, rel)
	}
	if resp.Version == "" {
		t.Fatal("Version is empty")
	}
	// Last path segment should be the version directory.
	if filepath.Base(resp.InstallPath) != resp.Version {
		t.Fatalf("InstallPath base %q should equal version %q", filepath.Base(resp.InstallPath), resp.Version)
	}

	mani := resp.ManifestPath
	if mani == "" {
		mani = filepath.Join(resp.InstallPath, "manifest.json")
	}
	data, err := os.ReadFile(mani)
	if err != nil {
		t.Fatalf("read manifest.json: %v", err)
	}
	var doc struct {
		Version string `json:"version"`
	}
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatalf("parse manifest.json: %v\n%s", err, data)
	}
	if doc.Version == "" {
		t.Fatal("manifest.json version field is empty")
	}
	if doc.Version != resp.Version {
		t.Fatalf("returned version %q != manifest version %q", resp.Version, doc.Version)
	}
}

func assertChromeArgsContract(t *testing.T, args []string, extPath, sessionURL string) {
	t.Helper()
	if len(args) == 0 {
		t.Fatal("ChromeArgs is empty")
	}
	joined := strings.Join(args, "\x00")
	// --load-extension=<path> or --load-extension <path>
	hasLoad := false
	for i, a := range args {
		if a == "--load-extension" && i+1 < len(args) && args[i+1] == extPath {
			hasLoad = true
			break
		}
		if strings.HasPrefix(a, "--load-extension=") {
			val := strings.TrimPrefix(a, "--load-extension=")
			if val == extPath {
				hasLoad = true
				break
			}
		}
	}
	if !hasLoad {
		// Also accept path as substring of a load-extension value (normalized abs).
		for _, a := range args {
			if strings.HasPrefix(a, "--load-extension=") && strings.Contains(a, extPath) {
				hasLoad = true
				break
			}
		}
	}
	if !hasLoad {
		t.Fatalf("ChromeArgs missing --load-extension=%s; args=%v", extPath, args)
	}

	for _, a := range args {
		if a == "--user-data-dir" || strings.HasPrefix(a, "--user-data-dir=") {
			t.Fatalf("ChromeArgs must not include --user-data-dir; args=%v", args)
		}
	}
	// Session URL should appear as an arg.
	if sessionURL != "" {
		foundURL := false
		for _, a := range args {
			if a == sessionURL || strings.Contains(a, sessionURL) {
				foundURL = true
				break
			}
		}
		if !foundURL {
			t.Fatalf("ChromeArgs should include session URL %q; args=%v", sessionURL, args)
		}
	}
	_ = joined
}

func assertInstallPathInJSON(t *testing.T, resp *Response) {
	t.Helper()
	if strings.TrimSpace(resp.ExtensionInstallPath) == "" {
		t.Fatalf("extension_install_path missing/empty; body=%s", truncate(resp.BodyString, 500))
	}
	if !filepath.IsAbs(resp.ExtensionInstallPath) {
		t.Fatalf("extension_install_path %q is not absolute", resp.ExtensionInstallPath)
	}
	// Path should exist on disk after Run extract.
	if st, err := os.Stat(resp.ExtensionInstallPath); err != nil || !st.IsDir() {
		t.Fatalf("extension_install_path %q is not a directory: %v", resp.ExtensionInstallPath, err)
	}
	if strings.TrimSpace(resp.EmbeddedVersion) == "" {
		t.Fatalf("embedded_version missing/empty; body=%s", truncate(resp.BodyString, 400))
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

func tail(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[len(s)-n:]
}

// silence unused in some leaves
var (
	_ = http.StatusOK
	_ = json.Marshal
)
```
