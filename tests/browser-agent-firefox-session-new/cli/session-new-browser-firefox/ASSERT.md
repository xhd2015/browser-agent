## Expected

- `HandleCLI` returns nil (`CLIErr` empty).
- `OpenFirefoxCallCount == 1`.
- `OpenFirefoxURL` contains `/go` and session id.
- Stdout includes `browser-agent-firefox` and about:debugging / Load Temporary.
- Stdout does not use `chrome://extensions` as primary install path.
- Stdout ends with trailing `\n`.
- Stdout contains session id.

## Side Effects

- Session created; OpenFirefoxFn spy only; HOME-isolated ensure under TestHome.

## Errors

- CLI error, zero open calls, or missing markers fails.

## Exit Code

- **0** (`HandleCLI` returns nil).

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
	if resp.CLIErr != "" {
		t.Fatalf("HandleCLI error: %s", resp.CLIErr)
	}
	if resp.OpenFirefoxCallCount != 1 {
		t.Fatalf("OpenFirefoxCallCount = %d, want 1", resp.OpenFirefoxCallCount)
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
