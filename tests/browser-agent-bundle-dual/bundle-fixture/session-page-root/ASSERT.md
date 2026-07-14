## Expected

Requirement **A2**:

- Bundle succeeds.
- `SessionPageDir` non-empty absolute under BundleRoot.
- Index file present (`index.html` or `session-page.html`).
- Index HTML has root mount: `id="root"` and/or `data-browser-agent-root` /
  `browser-agent-root`.

## Side Effects

- Session-page staged only under BundleRoot.

## Errors

- Empty SessionPageDir or missing root mount fails SPA embed contract.

## Exit Code

- 0.

```go
import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	assertExitZero(t, resp)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if strings.TrimSpace(resp.SessionPageDir) == "" {
		t.Fatal("SessionPageDir is empty")
	}
	if !filepath.IsAbs(resp.SessionPageDir) {
		t.Fatalf("SessionPageDir must be absolute; got %q", resp.SessionPageDir)
	}
	if req.BundleRoot != "" {
		rel, rerr := filepath.Rel(req.BundleRoot, resp.SessionPageDir)
		if rerr != nil || strings.HasPrefix(rel, "..") {
			t.Fatalf("SessionPageDir %q not under BundleRoot %q",
				resp.SessionPageDir, req.BundleRoot)
		}
	}
	html := resp.SessionIndexText
	if html == "" {
		// Try read from disk
		for _, name := range []string{"index.html", "session-page.html"} {
			p := filepath.Join(resp.SessionPageDir, name)
			if b, e := os.ReadFile(p); e == nil {
				html = string(b)
				resp.SessionIndexPath = p
				break
			}
		}
	}
	if strings.TrimSpace(html) == "" {
		t.Fatalf("session-page index missing/empty under %s", resp.SessionPageDir)
	}
	if !hasRootMount(html) {
		t.Fatalf("session-page index missing root mount; path=%s body=%s",
			resp.SessionIndexPath, truncate(html, 400))
	}
}
```
