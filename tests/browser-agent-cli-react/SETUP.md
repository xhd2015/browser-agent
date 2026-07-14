# Scenario

**Feature**: browser-agent CLI + React product shell + extension embed (no Chrome)

```
# CLI dispatch / side-commands
Operator -> HandleCLI(args, env, stdout, stderr) -> brief|help|info|eval

# Live side-commands
Test Client -> browseragent.Run(NoOpenChrome, NoAgentRun) -> Control Server
Fake Extension -> WS /v1/ws hello|result
HandleCLI session info|eval --session-id --addr -> JobResult / session snapshot + \n

# Product + embed + FS layout
Test Client -> ProductBrowserAgent | ProductBrowserTrace
Test Client -> ExtractEmbeddedExtension | BuildChromeArgs | InstallChromeExtension
Test Client -> GET /go SPA HTML (product markers)
Test Client -> read ModuleRoot/react/** and Chrome-Ext-Browser-Agent/**
```

## Preconditions

- Module path `github.com/xhd2015/browser-agent` is the workspace root.
- Tree root is `tests/browser-agent-cli-react/`; **ModuleRoot** =
  `filepath.Clean(filepath.Join(DOCTEST_ROOT, "..", ".."))`.
- Package `browseragent` will export (TDD red until implemented):
  - `HandleCLI(args, env, stdout, stderr) error`
  - `ProductConfig` + `ProductBrowserAgent` + `ProductBrowserTrace`
  - `ExtractEmbeddedExtension`, `InstallChromeExtension`, `BuildChromeArgs`
  - existing `Config` + `Run` for serve (NoOpenChrome / NoAgentRun)
- Each server/extract leaf uses isolated temp `BaseDir` and free loopback `Addr`.
- No real Chrome; no real agent-run; no webpack/npm in this tree.
- Sealed `tests/browser-agent/` must not be modified.

## Steps

1. Resolve `ModuleRoot` from `DOCTEST_ROOT`.
2. Allocate a unique temp `BaseDir` for leaves that extract or serve.
3. Default `SessionID`, `NoOpenChrome`, `NoAgentRun`, short `ReadyTimeout`.
4. Leave `Mode` and surface-specific fields for grouping/leaf Setup.
5. Default CLI env to empty map semantics in Run when nil (no ambient session).

## Context

- Parallel-safe: temp dirs + free ports per leaf.
- Shared helpers below available to all descendant Assert/Setup packages.
- Fixture shape reference: `testdata/mini-extension/` (CI embed under package).
- Prefer package APIs over building `cmd/browser-agent` binary.

```go
import (
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
	req.ModuleRoot = filepath.Clean(filepath.Join(DOCTEST_ROOT, "..", ".."))
	dir := t.TempDir()
	req.BaseDir = filepath.Join(dir, "browser-agent-base")
	if err := os.MkdirAll(req.BaseDir, 0o755); err != nil {
		return err
	}
	req.NoOpenChrome = true
	req.NoAgentRun = true
	if req.SessionID == "" {
		req.SessionID = fmt.Sprintf("ba-cli-%d", time.Now().UnixNano()%1e12)
	}
	if req.ReadyTimeout == 0 {
		req.ReadyTimeout = 5 * time.Second
	}
	if req.MaxDispatchWait == 0 {
		req.MaxDispatchWait = 3 * time.Second
	}
	return nil
}

func assertNoRunErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("Run transport/setup error: %v", err)
	}
}

func assertExitZero(t *testing.T, resp *Response) {
	t.Helper()
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if resp.ExitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr=%q err=%q stdout=%q cliErr=%q",
			resp.ExitCode, resp.Stderr, resp.ErrText, resp.Stdout, resp.CLIErr)
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

func assertPrintedTrailingNewline(t *testing.T, resp *Response) {
	t.Helper()
	body := resp.Stdout
	if strings.TrimSpace(body) == "" {
		body = resp.Stderr
	}
	if strings.TrimSpace(body) == "" {
		t.Fatal("printed usage/help empty on stdout+stderr")
	}
	if !strings.HasSuffix(body, "\n") {
		t.Fatalf("printed text must end with \\n; last bytes=%q", tail(body, 40))
	}
}

func combinedCLIText(resp *Response) string {
	return resp.Stdout + resp.Stderr
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

func assertHTMLContentType(t *testing.T, resp *Response) {
	t.Helper()
	ct := strings.ToLower(resp.ContentType)
	if !strings.Contains(ct, "html") {
		if !strings.Contains(strings.ToLower(resp.BodyString), "<html") &&
			!strings.Contains(strings.ToLower(resp.BodyString), "<!doctype") {
			t.Fatalf("Content-Type %q / body not HTML; body=%s",
				resp.ContentType, truncate(resp.BodyString, 200))
		}
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
	// Minimal JSON version check without requiring full schema.
	if !strings.Contains(string(data), "version") {
		t.Fatalf("manifest.json missing version field text; data=%s", truncate(string(data), 200))
	}
}

func assertChromeArgsContract(t *testing.T, args []string, extPath, sessionURL string) {
	t.Helper()
	if len(args) == 0 {
		t.Fatal("ChromeArgs is empty")
	}
	hasLoad := false
	for i, a := range args {
		if a == "--load-extension" && i+1 < len(args) && args[i+1] == extPath {
			hasLoad = true
			break
		}
		if strings.HasPrefix(a, "--load-extension=") {
			val := strings.TrimPrefix(a, "--load-extension=")
			if val == extPath || strings.Contains(a, extPath) {
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
}

func assertSessionSourcesInText(t *testing.T, text string) {
	t.Helper()
	if !strings.Contains(text, "--session-id") {
		t.Fatalf("error/text must mention --session-id; got %q", text)
	}
	if !strings.Contains(text, "BROWSER_AGENT_SESSION_ID") {
		t.Fatalf("error/text must mention BROWSER_AGENT_SESSION_ID; got %q", text)
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
)
```
