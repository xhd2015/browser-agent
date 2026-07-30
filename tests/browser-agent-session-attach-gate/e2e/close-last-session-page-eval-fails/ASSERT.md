---
label: e2e, slow, ui-automation
explanation: Real Chromium + MV3; last control tab closed then eval must fail
---

## Expected

- `playwright-debug` exits **0**.
- Stdout contains `extension_connected` with `ok: true`.
- Stdout contains `eval_before_close` with `ok: true` (armed path works).
- Stdout contains `eval_fails_after_last_session_page_closed` with `ok: true`
  (subsequent eval did **not** succeed after last control tab closed).
- `ExtensionDir` is non-empty absolute path.

## Side Effects

- Last session control tab closed; attach gate / leave detach must prevent a
  successful follow-up eval via sticky leftover attach.

## Errors

- If post-close eval still succeeds (`ok: false` on the fail-assert line), leaf fails.

## Exit Code

- Playwright subprocess exit **0**.

```go
import (
	"path/filepath"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
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
	if filepath.IsAbs(resp.ExtensionDir) == false || resp.ExtensionDir == "" {
		t.Fatalf("ExtensionDir must be non-empty absolute; got %q", resp.ExtensionDir)
	}
	assertLineOK(t, resp.AssertLines, "extension_connected")
	assertLineOK(t, resp.AssertLines, "eval_before_close")
	assertLineOK(t, resp.AssertLines, "eval_fails_after_last_session_page_closed")

	for _, l := range resp.AssertLines {
		if l.Assert == "eval_fails_after_last_session_page_closed" && l.SessionID != req.SessionID {
			t.Fatalf("session_id=%q want %q", l.SessionID, req.SessionID)
		}
	}
}
```
