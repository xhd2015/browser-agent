## Expected

Requirement **E4**:

- background.js found.
- Session window scoping: mentions `windowId` and/or `entry.windowId`.
- Protects session control page: mentions `/go?session` **or** `session` page
  close/navigate rejection language near Target.closeTarget / create path.
- Public identity: mentions **`tab_id`** in create/getTargets/result path.
- Soft: if `targetId` appears, it should be for **input resolution** only
  (decimal string → tab_id), not as sole public identity — do not require absence
  of all targetId strings if used as inbound alias.

## Side Effects

- None.

## Errors

- Missing window scope or tab_id identity breaks multi-tab session contract.

## Exit Code

- Not asserted.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if !resp.FileExists || strings.TrimSpace(resp.CombinedText) == "" {
		t.Fatalf("shell background missing; err=%q found=%v", resp.ErrText, resp.FoundPaths)
	}
	src := resp.CombinedText
	low := strings.ToLower(src)

	// These rules must apply to the Target.* / create_tab polyfill path — not only
	// pre-existing tab-targeting helpers. Require Target + create path together with scope.
	hasTargetPolyfillSurface := strings.Contains(src, "Target.") ||
		strings.Contains(src, "Target.createTarget") ||
		strings.Contains(src, "create_tab")
	if !hasTargetPolyfillSurface {
		t.Fatalf("session rules leaf requires Target.*/create_tab polyfill surface present; path=%v snippet=%s",
			resp.FoundPaths, truncate(src, 700))
	}

	hasWindow := strings.Contains(src, "windowId") ||
		strings.Contains(src, "window_id") ||
		strings.Contains(src, "entry.windowId")
	if !hasWindow {
		t.Fatalf("background must scope tabs to session windowId; path=%v snippet=%s",
			resp.FoundPaths, truncate(src, 700))
	}

	hasSessionPageGuard := strings.Contains(src, "/go?session") ||
		strings.Contains(src, "go?session") ||
		strings.Contains(src, "isSessionPage") ||
		strings.Contains(src, "sessionPage") ||
		strings.Contains(low, "session page") ||
		strings.Contains(low, "session-page") ||
		strings.Contains(low, "control page")
	// Require close/remove protection language when Target.closeTarget is polyfilled.
	hasCloseGuard := strings.Contains(src, "closeTarget") || strings.Contains(src, "tabs.remove")
	if !hasSessionPageGuard || !hasCloseGuard {
		t.Fatalf("background must protect session control page on closeTarget/tabs.remove; path=%v snippet=%s",
			resp.FoundPaths, truncate(src, 700))
	}

	// Result identity for create_tab / Target polyfill responses.
	hasTabIDResult := strings.Contains(src, "tab_id") ||
		(strings.Contains(src, "tabId") && (strings.Contains(src, "create_tab") || strings.Contains(src, "createTarget")))
	if !hasTabIDResult {
		t.Fatalf("background polyfill results must use tab_id identity; path=%v snippet=%s",
			resp.FoundPaths, truncate(src, 700))
	}
}
```
