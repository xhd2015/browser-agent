---
label: slow, ui-automation
explanation: Real Chromium + MV3 extension; eval+screenshot same tab_id attach reuse
---

## Expected

- `playwright-debug` exits **0**.
- JSON assert `extension_connected` with `ok: true`.
- JSON assert `eval_then_screenshot_same_tab` with `ok: true`.
- Active tab is **not** the pin target (`active_tab_is_pin_target` refers to active marker tab).
- Same `tab_id` used for both eval and screenshot jobs.
- `eval_on_pin_tab` is true (eval ran on pin tab, not active tab).
- Screenshot result includes non-empty base64 or format png.
- `ExtensionDir` is non-empty absolute path.

## Side Effects

- Two jobs on same pinned tab without attach failure.

## Errors

- Screenshot failure after eval on same tab_id fails (attach lifecycle bug).

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
	assertLineOK(t, resp.AssertLines, "eval_then_screenshot_same_tab")

	for _, l := range resp.AssertLines {
		if l.Assert == "eval_then_screenshot_same_tab" {
			evalOK, _ := l.Extra["eval_ok"].(bool)
			shotOK, _ := l.Extra["screenshot_ok"].(bool)
			if !evalOK || !shotOK {
				t.Fatalf("both eval and screenshot must succeed; extra=%v", l.Extra)
			}
			sameTab, _ := l.Extra["same_tab_id"].(bool)
			if !sameTab {
				t.Fatalf("eval and screenshot must use same tab_id; extra=%v", l.Extra)
			}
			if onPin, ok := l.Extra["eval_on_pin_tab"].(bool); ok && !onPin {
				t.Fatalf("eval must run on pin tab via tab_id, not active tab; extra=%v", l.Extra)
			}
			if activeIsPin, ok := l.Extra["active_tab_is_pin_target"].(bool); ok && !activeIsPin {
				t.Fatalf("active tab must be the non-pin active marker tab; extra=%v", l.Extra)
			}
		}
	}
}
```