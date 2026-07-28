## Expected

- Daemon health becomes OK.
- `GET /v1/sessions` status 200.
- Response JSON array includes `session_id` `sess-wire01`.

## Side Effects

- server.json written under BaseDir; process cleaned up via context cancel.

## Errors

- Missing id in list means restore not wired into RunDaemon (RED until implementer).

## Exit Code

- 0.

```go
import (
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if resp.HTTPStatusCode != 200 {
		t.Fatalf("GET /v1/sessions status=%d body=%q", resp.HTTPStatusCode, resp.HTTPBody)
	}
	found := false
	for _, id := range resp.HTTPSessionIDs {
		if id == "sess-wire01" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("GET /v1/sessions ids=%v missing sess-wire01 (RunDaemon must RestoreSessionsFromDisk)",
			resp.HTTPSessionIDs)
	}
	assertExitZero(t, resp)
}
```
