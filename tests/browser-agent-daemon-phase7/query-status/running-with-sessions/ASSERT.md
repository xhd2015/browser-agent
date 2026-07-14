## Expected

- `QueryDaemonStatus` returns **nil** error.
- `Running=true`.
- `PID` > 0; `Addr` and `BaseURL` match live daemon; `BaseDir` equals request base dir.
- `StartedAt` is non-zero; `Uptime` >= 0.
- `Sessions` session ids match `GET /v1/sessions` (includes created session).

## Side Effects

- `server.json` bytes **unchanged** after query.
- Daemon remains healthy after query.

## Errors

- Query error, `Running=false`, mismatched fields, or meta mutation fails.

## Exit Code

- Not asserted.

```go
import (
	"os"
	"strings"
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
	st := resp.Status
	if !st.Running {
		t.Fatal("Running=false, want true")
	}
	if st.PID <= 0 {
		t.Fatalf("PID=%d want >0", st.PID)
	}
	if st.Addr != resp.Addr {
		t.Fatalf("Addr=%q want %q", st.Addr, resp.Addr)
	}
	wantBaseURL := strings.TrimRight("http://"+resp.Addr, "/")
	if strings.TrimRight(st.BaseURL, "/") != wantBaseURL {
		t.Fatalf("BaseURL=%q want %q", st.BaseURL, wantBaseURL)
	}
	if st.BaseDir != req.BaseDir {
		t.Fatalf("BaseDir=%q want %q", st.BaseDir, req.BaseDir)
	}
	if st.StartedAt.IsZero() {
		t.Fatal("StartedAt is zero")
	}
	if st.Uptime < 0 {
		t.Fatalf("Uptime=%v want >=0", st.Uptime)
	}
	wantIDs := append([]string(nil), resp.SessionIDsFromHTTP...)
	if len(wantIDs) == 0 {
		wantIDs = []string{req.SessionID}
	}
	assertStringSlicesEqual(t, resp.StatusSessionIDs, wantIDs)
	assertMetaUnchanged(t, resp)
	if !resp.DaemonHealthyAfter {
		t.Fatal("daemon not healthy after read-only status query")
	}
	_ = os.Getpid()
}
```