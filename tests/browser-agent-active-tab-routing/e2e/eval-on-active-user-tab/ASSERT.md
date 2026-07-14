---
label: slow, ui-automation
explanation: Real Chromium + MV3 extension; two-tab active-tab routing regression
---

## Expected

- `playwright-debug` exits **0**.
- Stdout contains JSON assert line `extension_connected` with `ok: true`.
- Stdout contains JSON assert line `active_tab_routing` with `ok: true`.
- `active_tab_routing` extra `eval_url` contains `LOOP_MARKER=active-tab-routing`.
- `user_tab_active` is true in `active_tab_routing` assert line.
- `eval_on_session_page` is false (eval did not run on `/go?session=` tab).
- `ExtensionDir` is non-empty absolute path from `ExtractEmbeddedExtension`.

## Side Effects

- Session registered on daemon; extension connects; eval executes on active user tab.

## Errors

- Missing assert line, `ok: false`, `eval_url` on session page, or non-zero playwright exit fails.

## Exit Code

- Playwright subprocess exit **0**.

```go
import (
	"path/filepath"
	"strings"
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
	assertLineOK(t, resp.AssertLines, "active_tab_routing")

	for _, l := range resp.AssertLines {
		if l.Assert == "extension_connected" && l.SessionID != req.SessionID {
			t.Fatalf("extension_connected session_id=%q want %q", l.SessionID, req.SessionID)
		}
		if l.Assert == "active_tab_routing" {
			if l.SessionID != req.SessionID {
				t.Fatalf("active_tab_routing session_id=%q want %q", l.SessionID, req.SessionID)
			}
			evalURL, _ := l.Extra["eval_url"].(string)
			if !strings.Contains(evalURL, "LOOP_MARKER=active-tab-routing") {
				t.Fatalf("eval_url must contain LOOP_MARKER; got %q extra=%v", evalURL, l.Extra)
			}
			if evalOnSession, ok := l.Extra["eval_on_session_page"].(bool); ok && evalOnSession {
				t.Fatalf("eval must not run on session page; extra=%v", l.Extra)
			}
			if userActive, ok := l.Extra["user_tab_active"].(bool); ok && !userActive {
				t.Fatalf("user tab must be active before eval; extra=%v", l.Extra)
			}
		}
	}
}
```