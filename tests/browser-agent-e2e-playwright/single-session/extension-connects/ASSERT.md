---
label: slow, ui-automation
explanation: Real Chromium + MV3 extension; polls /v1/session until connected
---

## Expected

- `playwright-debug` exits **0**.
- Stdout contains JSON assert line `extension_connected` with `ok: true`.
- `session_id` in assert line matches `sess-e2e-connect`.
- `ExtensionDir` is non-empty absolute path from `ExtractEmbeddedExtension`.

## Side Effects

- Session registered on daemon; extension connects in real browser tab.

## Errors

- Missing assert line, `ok: false`, or non-zero playwright exit fails.

## Exit Code

- Playwright subprocess exit **0**.

```go
import (
	"path/filepath"
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
	if filepath.IsAbs(resp.ExtensionDir) == false || resp.ExtensionDir == "" {
		t.Fatalf("ExtensionDir must be non-empty absolute; got %q", resp.ExtensionDir)
	}
	assertLineOK(t, resp.AssertLines, "extension_connected")
	for _, l := range resp.AssertLines {
		if l.Assert == "extension_connected" && l.SessionID != req.SessionID {
			t.Fatalf("extension_connected session_id=%q want %q", l.SessionID, req.SessionID)
		}
	}
}
```