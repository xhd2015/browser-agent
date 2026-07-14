## Expected

Requirement **P5**:

- `manifest.json` found under `Chrome-Ext-Browser-Agent`.
- `content_scripts` entry includes **`contentScript.js`** (or `contentScript` basename).
- At least one `matches` pattern targets **loopback** host (`127.0.0.1` and/or `localhost`).
- At least one `matches` pattern includes **`/go`** path segment (session page).

## Side Effects

- None (read-only FS).

## Errors

- Missing go-page match prevents content script register on session page.

## Exit Code

- Not asserted.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if !resp.FileExists || strings.TrimSpace(resp.ManifestText) == "" {
		t.Fatalf("manifest missing under ModuleRoot=%s; err=%q found=%v",
			req.ModuleRoot, resp.ErrText, resp.FoundPaths)
	}
	low := strings.ToLower(resp.ManifestText)
	if !strings.Contains(low, "content_scripts") {
		t.Fatalf("manifest must declare content_scripts; manifest=%s", truncate(resp.ManifestText, 400))
	}
	if !strings.Contains(low, "contentscript.js") {
		t.Fatalf("manifest content_scripts must include contentScript.js; manifest=%s",
			truncate(resp.ManifestText, 400))
	}
	matches := resp.ContentScriptMatch
	if len(matches) == 0 {
		t.Fatalf("manifest content_scripts matches missing or unparseable; manifest=%s",
			truncate(resp.ManifestText, 400))
	}
	hasLoopback := false
	hasGoPath := false
	for _, m := range matches {
		ml := strings.ToLower(m)
		if strings.Contains(ml, "127.0.0.1") || strings.Contains(ml, "localhost") {
			hasLoopback = true
		}
		if strings.Contains(ml, "/go") {
			hasGoPath = true
		}
	}
	if !hasLoopback {
		t.Fatalf("content_scripts matches must include loopback host; matches=%v", matches)
	}
	if !hasGoPath {
		t.Fatalf("content_scripts matches must include /go path for session page; matches=%v", matches)
	}
}
```