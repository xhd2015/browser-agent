## Expected

- `SessionNew` returns **nil** error.
- `OpenChromeCallCount == 1`.
- `OpenChromeSessionURL` contains `/go` and `sess-new-8`.
- `OpenChromeExtPath` is non-empty (absolute preferred).
- `AgentRunProbeCallCount == 0`.
- Session id appears in `GET /v1/sessions`.

## Side Effects

- Session registered on daemon; extension install path recorded by hook only.

## Errors

- Missing OpenChrome call, agent-run probe invoked, or session not on server fails.

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
	if resp.SessionNewErr != "" {
		t.Fatalf("SessionNew error: %s", resp.SessionNewErr)
	}
	if resp.OpenChromeCallCount != 1 {
		t.Fatalf("OpenChromeFn call count=%d, want 1", resp.OpenChromeCallCount)
	}
	if resp.AgentRunProbeCallCount != 0 {
		t.Fatalf("AgentRunProbeFn called %d times; want 0 (no agent-run)", resp.AgentRunProbeCallCount)
	}
	sid := req.SessionID
	u := resp.OpenChromeSessionURL
	if u == "" {
		t.Fatal("OpenChrome sessionURL empty")
	}
	if !strings.Contains(u, "/go") {
		t.Fatalf("sessionURL should contain /go; got %q", u)
	}
	if !strings.Contains(u, sid) {
		t.Fatalf("sessionURL should contain session id %q; got %q", sid, u)
	}
	if strings.TrimSpace(resp.OpenChromeExtPath) == "" {
		t.Fatal("OpenChrome extensionInstallPath empty")
	}
	if !filepath.IsAbs(resp.OpenChromeExtPath) {
		t.Fatalf("extensionInstallPath should be absolute; got %q", resp.OpenChromeExtPath)
	}
	if !resp.SessionOnServer {
		t.Fatalf("session %q not in GET /v1/sessions: %v", sid, resp.ServerSessionIDs)
	}
}```
