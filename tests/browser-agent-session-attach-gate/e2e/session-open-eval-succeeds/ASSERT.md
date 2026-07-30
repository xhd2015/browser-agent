---
label: e2e, slow, ui-automation
explanation: Real Chromium + MV3 extension; session-open attach-allowed baseline
---

## Expected

- `playwright-debug` exits **0**.
- Stdout contains JSON assert line `extension_connected` with `ok: true`.
- Stdout contains JSON assert line `session_open_eval` with `ok: true`.
- `session_open_eval` extra `eval_url` contains `LOOP_MARKER=attach-gate-open`.
- `ExtensionDir` is non-empty absolute path from `ExtractEmbeddedExtension`.

## Side Effects

- Session registered; extension connects; eval runs on user tab while control tab open.

## Errors

- Missing assert line, `ok: false`, or non-zero playwright exit fails.

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
	assertLineOK(t, resp.AssertLines, "session_open_eval")

	for _, l := range resp.AssertLines {
		if l.Assert == "extension_connected" && l.SessionID != req.SessionID {
			t.Fatalf("extension_connected session_id=%q want %q", l.SessionID, req.SessionID)
		}
		if l.Assert == "session_open_eval" {
			if l.SessionID != req.SessionID {
				t.Fatalf("session_open_eval session_id=%q want %q", l.SessionID, req.SessionID)
			}
			evalURL, _ := l.Extra["eval_url"].(string)
			if !strings.Contains(evalURL, "LOOP_MARKER=attach-gate-open") {
				t.Fatalf("eval_url must contain LOOP_MARKER; got %q extra=%v", evalURL, l.Extra)
			}
		}
	}
}
```
