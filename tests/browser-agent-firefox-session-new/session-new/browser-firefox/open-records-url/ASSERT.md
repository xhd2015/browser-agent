## Expected

- `SessionNew` returns nil error.
- `OpenFirefoxCallCount == 1`.
- `OpenFirefoxURL` contains `/go` and the session id.
- `OpenChromeCallCount == 0` (chrome open must not run).
- Stdout includes firefox extension path marker (`browser-agent-firefox`) and
  about:debugging / Load Temporary guidance.
- Stdout does not use `chrome://extensions` as primary install path.
- Stdout ends with trailing `\n`.
- Stdout still includes session id / Session URL style markers.

## Side Effects

- Session created on daemon; OpenFirefoxFn spy only (no real browser).
- Firefox extension extracted under TestHome.

## Errors

- Missing open call, chrome open invoked, or missing stdout markers fails.

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
	if resp.OpenFirefoxCallCount != 1 {
		t.Fatalf("OpenFirefoxCallCount = %d, want 1", resp.OpenFirefoxCallCount)
	}
	if resp.OpenChromeCallCount != 0 {
		t.Fatalf("OpenChromeCallCount = %d, want 0 (Browser=firefox must not open Chrome)", resp.OpenChromeCallCount)
	}
	u := resp.OpenFirefoxURL
	if u == "" {
		t.Fatal("OpenFirefoxURL is empty")
	}
	if !strings.Contains(u, "/go") {
		t.Fatalf("OpenFirefoxURL should contain /go; got %q", u)
	}
	sid := req.SessionID
	if sid == "" {
		sid = resp.SessionID
	}
	if sid != "" && !strings.Contains(u, sid) {
		t.Fatalf("OpenFirefoxURL should contain session id %q; got %q", sid, u)
	}
	assertFirefoxSessionNewStdout(t, resp.Stdout)
	assertContainsFold(t, resp.Stdout, sid)
}
```
