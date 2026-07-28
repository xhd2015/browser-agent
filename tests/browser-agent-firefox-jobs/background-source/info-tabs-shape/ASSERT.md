## Expected

- `background.js` found.
- Info path present: function name **`handleInfoJob`** **or** job type token **`info`**.
- Tab listing: **`tabs.query`** (or listCapturable / equivalent listing helper language).
- Result shape markers: response constructs a **`tabs`** collection and tab fields
  **`id`**, **`url`**, **`title`** appear in the info/list path (source-level).

## Side Effects

- None.

## Errors

- Phase-1 stub without tabs listing fails info contract.

## Exit Code

- Not asserted.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	src := assertBackgroundPresent(t, req, resp)
	low := strings.ToLower(src)

	hasHandlerName := strings.Contains(src, "handleInfoJob") || strings.Contains(low, "handleinfojob")
	hasInfoToken := jobTypeTokenPresent(src, "info")
	if !hasHandlerName && !hasInfoToken {
		t.Fatalf("background must have handleInfoJob and/or info job branch; path=%v snippet=%s",
			resp.FoundPaths, truncate(src, 600))
	}
	if !hasTabsQueryOrList(src) {
		t.Fatalf("info path must list tabs (tabs.query / listCapturable / …); path=%v snippet=%s",
			resp.FoundPaths, truncate(src, 600))
	}
	// Shape: tabs array + common tab fields in source.
	if !strings.Contains(src, "tabs") && !strings.Contains(src, `"tabs"`) {
		t.Fatalf("info result shape should mention tabs; path=%v snippet=%s",
			resp.FoundPaths, truncate(src, 600))
	}
	for _, field := range []string{"id", "url", "title"} {
		// Accept quoted keys preferred for result objects.
		if !strings.Contains(src, field) {
			t.Fatalf("info tabs shape missing field marker %q; path=%v snippet=%s",
				field, resp.FoundPaths, truncate(src, 600))
		}
	}
}
```
