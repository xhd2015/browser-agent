## Expected

- `background.js` found.
- Method token **`Page.navigate`** present.
- Navigation API **`tabs.update`** present (`browser.tabs.update` /
  `chrome.tabs.update` / `.tabs.update`).

## Side Effects

- None.

## Errors

- `Page.navigate` without `tabs.update` is incomplete Firefox polyfill.
- Debugger-only navigate without tabs.update fails.

## Exit Code

- Not asserted.

```go
import (
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	src := assertBackgroundPresent(t, req, resp)
	if !hasPageNavigate(src) {
		t.Fatalf("background missing Page.navigate CDP method token; path=%v snippet=%s",
			resp.FoundPaths, truncate(src, 700))
	}
	if !hasTabsUpdate(src) {
		t.Fatalf("Page.navigate must map to tabs.update on Firefox; path=%v snippet=%s",
			resp.FoundPaths, truncate(src, 700))
	}
}
```
