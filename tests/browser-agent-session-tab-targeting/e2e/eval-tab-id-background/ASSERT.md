---
label: slow, ui-automation
explanation: Real Chromium + MV3 extension; three-tab tab_id pinning regression
---

## Expected

- `playwright-debug` exits **0**.
- JSON assert `extension_connected` with `ok: true`.
- JSON assert `tab_id_background_eval` with `ok: true`.
- `tab_id_background_eval` extra `eval_url` contains `BG_MARKER=tab-id-background`.
- `active_tab_is_user` is true; `eval_on_background_tab` is true.
- `ExtensionDir` is non-empty absolute path.

## Side Effects

- Session registered; extension connects; eval runs on pinned background tab.

## Errors

- Eval on active user tab instead of background tab fails.

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
	assertLineOK(t, resp.AssertLines, "tab_id_background_eval")

	for _, l := range resp.AssertLines {
		if l.Assert == "tab_id_background_eval" {
			evalURL, _ := l.Extra["eval_url"].(string)
			if !strings.Contains(evalURL, "BG_MARKER=tab-id-background") {
				t.Fatalf("eval_url must contain BG_MARKER; got %q extra=%v", evalURL, l.Extra)
			}
			if onBg, ok := l.Extra["eval_on_background_tab"].(bool); ok && !onBg {
				t.Fatalf("eval must run on background tab; extra=%v", l.Extra)
			}
			if userActive, ok := l.Extra["active_tab_is_user"].(bool); ok && !userActive {
				t.Fatalf("user tab should remain active; extra=%v", l.Extra)
			}
		}
	}
}
```