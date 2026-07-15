## Expected

Requirement **E3** (Tier A full polyfill):

- background.js found.
- Source mentions each Tier A method token:
  - `Target.createTarget`
  - `Target.closeTarget`
  - `Target.activateTarget`
  - `Target.getTargets`
  - `Target.getTargetInfo`
  Accept short forms without `Target.` prefix if `createTarget`/`closeTarget`/…
  appear **and** at least one full `Target.` method is present.
- Source uses chrome.tabs APIs: at least **three** of
  `chrome.tabs.create`, `chrome.tabs.remove`, `chrome.tabs.update`,
  `chrome.tabs.query`, `chrome.tabs.get`.

## Side Effects

- None.

## Errors

- Missing Tier A methods means incomplete polyfill.

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

	methods := []struct {
		full  string
		short string
	}{
		{"Target.createTarget", "createTarget"},
		{"Target.closeTarget", "closeTarget"},
		{"Target.activateTarget", "activateTarget"},
		{"Target.getTargets", "getTargets"},
		{"Target.getTargetInfo", "getTargetInfo"},
	}
	fullHits := 0
	for _, m := range methods {
		if strings.Contains(src, m.full) {
			fullHits++
			continue
		}
		if strings.Contains(src, m.short) && strings.Contains(src, "Target.") {
			fullHits++
			continue
		}
		t.Fatalf("shell background missing Tier A method %q (or short %q with Target.); path=%v snippet=%s",
			m.full, m.short, resp.FoundPaths, truncate(src, 600))
	}
	_ = fullHits

	tabsAPIs := []string{
		"chrome.tabs.create",
		"chrome.tabs.remove",
		"chrome.tabs.update",
		"chrome.tabs.query",
		"chrome.tabs.get",
	}
	tabHits := 0
	for _, api := range tabsAPIs {
		if strings.Contains(src, api) {
			tabHits++
		}
	}
	if tabHits < 3 {
		t.Fatalf("shell background need ≥3 chrome.tabs.* APIs for Tier A; got %d; path=%v snippet=%s",
			tabHits, resp.FoundPaths, truncate(src, 600))
	}
}
```
