## Expected

Requirement **B2**:

- Serve healthy.
- `OpenChromeCallCount == 1`.
- `OpenChromeSessionURL` contains `/go` and the live session id.
- `OpenChromeExtPath` is non-empty (absolute preferred; non-empty required).

## Side Effects

- Injector only; no real Chrome binary.

## Errors

- Zero calls means serve skipped chrome open incorrectly.
- Bad URL/path means launcher cannot load extension on session page.

## Exit Code

- Not asserted.

```go
import (
	"path/filepath"
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if resp.OpenChromeCallCount != 1 {
		t.Fatalf("OpenChromeFn call count=%d, want 1", resp.OpenChromeCallCount)
	}
	sid := req.SessionID
	if sid == "" {
		sid = resp.RealSessionID
	}
	u := resp.OpenChromeSessionURL
	if u == "" {
		t.Fatal("OpenChrome sessionURL empty")
	}
	if !strings.Contains(u, "/go") {
		t.Fatalf("sessionURL should contain /go; got %q", u)
	}
	if sid != "" && !strings.Contains(u, sid) {
		t.Fatalf("sessionURL should contain session id %q; got %q", sid, u)
	}
	if strings.TrimSpace(resp.OpenChromeExtPath) == "" {
		t.Fatal("OpenChrome extensionInstallPath empty")
	}
	// Prefer absolute path (soft: warn-level as fatal for CI clarity).
	if !filepath.IsAbs(resp.OpenChromeExtPath) {
		t.Fatalf("extensionInstallPath should be absolute; got %q", resp.OpenChromeExtPath)
	}
}
```
