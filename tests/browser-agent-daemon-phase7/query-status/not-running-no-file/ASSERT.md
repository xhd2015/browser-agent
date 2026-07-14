## Expected

- `QueryDaemonStatus` returns **nil** error.
- `Running=false`.
- `Sessions` empty (or nil).
- `server.json` still **absent**.

## Side Effects

- No new files under `BaseDir`.

## Errors

- Query error or `Running=true` fails.

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
	if resp.QueryErr != "" {
		t.Fatalf("QueryDaemonStatus error: %s", resp.QueryErr)
	}
	if resp.Status.Running {
		t.Fatal("Running=true, want false when server.json missing")
	}
	if len(resp.Status.Sessions) != 0 {
		t.Fatalf("Sessions len=%d want 0", len(resp.Status.Sessions))
	}
	if resp.DaemonMetaBeforeHit {
		t.Fatal("server.json existed before query")
	}
	if resp.DaemonMetaAfterHit {
		t.Fatal("server.json created by read-only status query")
	}
	assertMetaUnchanged(t, resp)
}
```