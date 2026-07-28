## Expected

- `background.js` found.
- Job type token **`screenshot`** present.
- Viewport API: **`captureVisibleTab`** (`browser.tabs.captureVisibleTab` / `tabs.captureVisibleTab`).

## Side Effects

- None.

## Errors

- CDP-only `Page.captureScreenshot` without `captureVisibleTab` fails Phase 2 contract.

## Exit Code

- Not asserted.

```go
import (
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	src := assertBackgroundPresent(t, req, resp)
	if !jobTypeTokenPresent(src, "screenshot") {
		t.Fatalf("background missing screenshot job branch token; path=%v snippet=%s",
			resp.FoundPaths, truncate(src, 600))
	}
	if !hasCaptureVisibleTab(src) {
		t.Fatalf("screenshot must use tabs.captureVisibleTab; path=%v snippet=%s",
			resp.FoundPaths, truncate(src, 600))
	}
}
```
