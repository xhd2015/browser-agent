---
label: e2e, slow, ui-automation
explanation: Real Chromium + MV3; two control tabs, close one, stay armed
---

## Expected

- `playwright-debug` exits **0**.
- Stdout contains `extension_connected` with `ok: true`.
- Stdout contains `eval_before_partial_close` with `ok: true`.
- Stdout contains `stays_armed_after_one_of_two_closed` with `ok: true`
  (eval still succeeds after closing one of two control tabs).
- `eval_url` on the stays-armed assert contains `LOOP_MARKER=attach-gate-two-pages`.
- `ExtensionDir` is non-empty absolute path.

## Side Effects

- One of two session control tabs closed; remaining control tab keeps session armed.

## Errors

- Unregister on single `entry.tabId` close (current master) makes post-close eval fail → RED.

## Exit Code

- Playwright subprocess exit **0**.

```go
import (
	"path/filepath"
	"strings"
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
	assertLineOK(t, resp.AssertLines, "eval_before_partial_close")
	assertLineOK(t, resp.AssertLines, "stays_armed_after_one_of_two_closed")

	for _, l := range resp.AssertLines {
		if l.Assert == "stays_armed_after_one_of_two_closed" {
			if l.SessionID != req.SessionID {
				t.Fatalf("session_id=%q want %q", l.SessionID, req.SessionID)
			}
			evalURL, _ := l.Extra["eval_url"].(string)
			if !strings.Contains(evalURL, "LOOP_MARKER=attach-gate-two-pages") {
				t.Fatalf("eval_url must contain LOOP_MARKER; got %q extra=%v", evalURL, l.Extra)
			}
		}
	}
}
```
