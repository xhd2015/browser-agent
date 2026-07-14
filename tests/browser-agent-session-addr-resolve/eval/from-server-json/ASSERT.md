## Expected

After fix (Classic TDD target — **RED** on current code):

- Combined stdout/stderr/CLIErr must **not** contain:
  - `session not found`
  - `unknown session`
  - `status 404` together with `session` (wrong-host 404 symptom)
- CLI may still error on job timeout / disconnected extension — that is **not** a failure.
- `DispatchTimedOut` false (harness wait only).

## Side Effects

- Job may be enqueued on correct daemon (optional; not asserted without fake WS).

## Errors

- Wrong-host `unknown session id` / `session not found` 404 fails.

## Exit Code

- 0 or non-zero acceptable when error is **not** unknown-session 404.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if resp.DispatchTimedOut {
		t.Fatal("session eval timed out in harness")
	}
	if !resp.DaemonMetaExists {
		t.Fatal("server.json must exist for from-server-json leaf")
	}
	if resp.SessionID == "" {
		t.Fatal("harness did not create session")
	}

	combined := strings.ToLower(combinedOutput(resp))
	if strings.Contains(combined, "session not found") {
		t.Fatalf("eval must not hit wrong host and return session not found; output=%s",
			truncate(combinedOutput(resp), 800))
	}
	if strings.Contains(combined, "unknown session") {
		t.Fatalf("eval must not return unknown session (404 on wrong daemon); output=%s",
			truncate(combinedOutput(resp), 800))
	}
	// Reject the LOOP symptom: info/eval 404 body from default :43761.
	if strings.Contains(combined, "status 404") && strings.Contains(combined, "session") {
		t.Fatalf("eval must not fail with session-related status 404 when meta addr differs; output=%s",
			truncate(combinedOutput(resp), 800))
	}
}
```