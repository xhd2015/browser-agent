## Expected

- `SessionNew` returns nil error.
- `OpenFirefoxCallCount == 0`.
- `OpenChromeCallCount == 0`.
- Stdout still includes `browser-agent-firefox` path and about:debugging /
  Load Temporary guidance (for manual load).
- Stdout does not use `chrome://extensions` as primary install path.
- Stdout ends with trailing `\n`.

## Side Effects

- Session created; no browser open; extension ensure may still write under TestHome.

## Errors

- Any open call or missing path markers fails.

## Exit Code

- N/A (package API).

```go
import (
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
	if resp.OpenFirefoxCallCount != 0 {
		t.Fatalf("OpenFirefoxCallCount = %d, want 0 when NoOpenChrome", resp.OpenFirefoxCallCount)
	}
	if resp.OpenChromeCallCount != 0 {
		t.Fatalf("OpenChromeCallCount = %d, want 0 when NoOpenChrome", resp.OpenChromeCallCount)
	}
	assertFirefoxSessionNewStdout(t, resp.Stdout)
	assertContainsFold(t, resp.Stdout, req.SessionID)
}
```
