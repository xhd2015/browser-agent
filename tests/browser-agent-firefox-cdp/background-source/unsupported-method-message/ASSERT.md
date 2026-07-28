## Expected

- `background.js` found.
- Source contains product-owned unsupported wording with **both**:
  - **`not supported`** or **`unsupported`** (case-insensitive), and
  - **`firefox`** (case-insensitive).
- Prefer message near CDP / method handling (not only unrelated comments).

## Side Effects

- None.

## Errors

- Missing either token fails (generic `"not implemented: firefox job type="`
  without **not supported** / **unsupported** is insufficient for this leaf).

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
	if !hasUnsupportedFirefoxMessage(src) {
		t.Fatalf("background missing unsupported CDP message tokens (need \"not supported\"|\"unsupported\" AND \"firefox\"); path=%v snippet=%s",
			resp.FoundPaths, truncate(src, 700))
	}
	// Soft preference: message should relate to cdp/method, not only unrelated text.
	low := strings.ToLower(src)
	if !(strings.Contains(low, "cdp") || strings.Contains(src, "method") ||
		strings.Contains(src, "Page.") || strings.Contains(src, "Runtime.")) {
		t.Fatalf("unsupported message should appear in CDP/method handling context; path=%v snippet=%s",
			resp.FoundPaths, truncate(src, 700))
	}
}
```
