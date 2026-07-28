## Expected

- `SessionNew` returns nil error.
- `OpenChromeCallCount == 1`.
- `OpenFirefoxCallCount == 0`.
- `OpenChromeSessionURL` contains `/go` and session id.
- Stdout includes session id and is **not** firefox-primary:
  - does **not** require about:debugging
  - must **not** present `browser-agent-firefox` as the extension path
    (chrome path uses managed-chrome / browser-agent chrome segment)

## Side Effects

- Session created; chrome open spied only.

## Errors

- Firefox open invoked, chrome not opened, or firefox path in stdout fails.

## Exit Code

- N/A (package API).

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if resp.SessionNewErr != "" {
		t.Fatalf("SessionNewErr = %q", resp.SessionNewErr)
	}
	if resp.OpenChromeCallCount != 1 {
		t.Fatalf("OpenChromeCallCount = %d, want 1 (default Browser is chrome)", resp.OpenChromeCallCount)
	}
	if resp.OpenFirefoxCallCount != 0 {
		t.Fatalf("OpenFirefoxCallCount = %d, want 0 when Browser empty/chrome", resp.OpenFirefoxCallCount)
	}
	u := resp.OpenChromeSessionURL
	if u == "" {
		t.Fatal("OpenChromeSessionURL is empty")
	}
	if !strings.Contains(u, "/go") {
		t.Fatalf("OpenChromeSessionURL should contain /go; got %q", u)
	}
	if !strings.Contains(u, req.SessionID) {
		t.Fatalf("OpenChromeSessionURL should contain session id %q; got %q", req.SessionID, u)
	}
	assertContainsFold(t, resp.Stdout, req.SessionID)
	// Chrome default must not print the Firefox canonical segment as extension path.
	assertNotContainsFold(t, resp.Stdout, "browser-agent-firefox")
	// Chrome instructions stay chrome-oriented (Load unpacked / install-chrome-extension).
	// Accept either existing formatSessionNewOutput chrome text.
	low := strings.ToLower(resp.Stdout)
	if strings.Contains(low, "about:debugging") {
		t.Fatalf("default chrome stdout must not use about:debugging as install path; got:\n%s", truncate(resp.Stdout, 800))
	}
}
```
