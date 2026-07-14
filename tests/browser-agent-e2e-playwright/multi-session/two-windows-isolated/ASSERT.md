---
label: slow, ui-automation
explanation: Two Chromium tabs; dual session WS connect poll
---

## Expected

- `playwright-debug` exits **0**.
- Assert `session_a_connected` with `ok: true`, `session_id` = `sess-e2e-a`.
- Assert `session_b_connected` with `ok: true`, `session_id` = `sess-e2e-b`.

## Side Effects

- Two sessions on daemon; both tabs connect extension independently.

## Errors

- Either session fails to connect or playwright non-zero exit fails.

## Exit Code

- Playwright subprocess exit **0**.

```go
import (
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if resp.Skipped {
		t.Skip(resp.SkipMsg)
	}
	if resp.PlaywrightExitCode != 0 {
		t.Fatalf("playwright-debug exit=%d stderr=%q stdout=%q",
			resp.PlaywrightExitCode, resp.PlaywrightStderr, resp.PlaywrightStdout)
	}
	assertLineOK(t, resp.AssertLines, "session_a_connected")
	assertLineOK(t, resp.AssertLines, "session_b_connected")
	for _, l := range resp.AssertLines {
		if l.Assert == "session_a_connected" && l.SessionID != req.SessionID {
			t.Fatalf("session_a_connected session_id=%q want %q", l.SessionID, req.SessionID)
		}
		if l.Assert == "session_b_connected" && l.SessionID != req.SessionIDB {
			t.Fatalf("session_b_connected session_id=%q want %q", l.SessionID, req.SessionIDB)
		}
	}
}
```