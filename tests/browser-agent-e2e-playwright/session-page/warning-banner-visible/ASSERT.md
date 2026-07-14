---
label: slow, ui-automation
explanation: Real Chromium session page; DOM warning banner + data-session-id
---

## Expected

- `playwright-debug` exits **0**.
- Assert line `warning_banner_present` with `ok: true`.
- Assert line `warning_banner_session_id` with `ok: true`.
- `session_id` fields match `sess-e2e-banner`.

## Side Effects

- Session page served at `/go`; warning banner rendered for active session.

## Errors

- Missing banner, mismatched `data-session-id`, or non-zero exit fails.

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
	assertLineOK(t, resp.AssertLines, "warning_banner_present")
	assertLineOK(t, resp.AssertLines, "warning_banner_session_id")
	for _, name := range []string{"warning_banner_present", "warning_banner_session_id"} {
		for _, l := range resp.AssertLines {
			if l.Assert == name && l.SessionID != "" && l.SessionID != req.SessionID {
				t.Fatalf("%s session_id=%q want %q", name, l.SessionID, req.SessionID)
			}
		}
	}
}
```