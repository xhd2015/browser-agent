## Expected

- `SessionNew` returns **nil** error.
- `resp.SessionID` (from stdout parse) matches `^sess-[a-z0-9]{6}$`.
- Generated id appears in `GET /v1/sessions`.
- `OpenChromeCallCount == 1` (session new still opens Chrome).

## Side Effects

- New session dir under `{BaseDir}/sessions/{generated-id}/`.

## Errors

- Empty id, invalid format, or missing from server list fails.

## Exit Code

- Not asserted.

```go
import (
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
	sid := resp.SessionID
	if sid == "" {
		sid = extractSessionIDFromStdout(resp.Stdout)
	}
	if sid == "" {
		t.Fatalf("could not determine generated session id from stdout:\n%s", truncate(resp.Stdout, 600))
	}
	if !sessGenerateRe.MatchString(sid) {
		t.Fatalf("generated session id %q does not match ^sess-[a-z0-9]{6}$", sid)
	}
	if !resp.SessionOnServer {
		t.Fatalf("generated session %q not in GET /v1/sessions: %v", sid, resp.ServerSessionIDs)
	}
	if resp.OpenChromeCallCount != 1 {
		t.Fatalf("OpenChromeFn call count=%d, want 1", resp.OpenChromeCallCount)
	}
}```
