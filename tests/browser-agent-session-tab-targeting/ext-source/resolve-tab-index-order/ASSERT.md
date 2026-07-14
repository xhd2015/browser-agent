## Expected

- `background.js` defines **tab_index** / index resolution for session window capturable tabs.
- Uses **1-based** indexing (not 0-based array index alone).
- Queries capturable tabs scoped to **`entry.windowId`** (left→right Chrome order).
- Includes session page in capturable tab list.

## Side Effects

- None (read-only FS).

## Errors

- Missing index resolver or 0-based-only logic breaks `--tab-index` contract.

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
		t.Fatalf("shell background missing under ModuleRoot=%s; err=%q found=%v",
			req.ModuleRoot, resp.ErrText, resp.FoundPaths)
	}
	text := resp.CombinedText
	if !hasTabIndexOrderResolution(text) {
		t.Fatalf("background must resolve 1-based tab_index over capturable tabs in session window; text=%s",
			truncate(text, 900))
	}
}
```