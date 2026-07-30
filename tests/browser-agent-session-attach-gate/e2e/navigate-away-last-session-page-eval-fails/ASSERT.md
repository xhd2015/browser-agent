---
label: e2e, slow, ui-automation
explanation: Real Chromium + MV3; last control tab navigated away then eval must fail
---

## Expected

- `playwright-debug` exits **0**.
- Stdout contains `extension_connected` with `ok: true`.
- Stdout contains `eval_before_navigate_away` with `ok: true`.
- Stdout contains `eval_fails_after_last_session_page_navigated_away` with `ok: true`
  (subsequent eval did **not** succeed after control tab left `/go?session=`).
- `ExtensionDir` is non-empty absolute path.

## Side Effects

- Last session control tab navigated away; gate/leave detach prevents successful eval.

## Errors

- Post-navigate eval still succeeding fails the leaf.

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
	assertLineOK(t, resp.AssertLines, "eval_before_navigate_away")
	assertLineOK(t, resp.AssertLines, "eval_fails_after_last_session_page_navigated_away")

	for _, l := range resp.AssertLines {
		if l.Assert == "eval_fails_after_last_session_page_navigated_away" && l.SessionID != req.SessionID {
			t.Fatalf("session_id=%q want %q", l.SessionID, req.SessionID)
		}
	}
}
```
