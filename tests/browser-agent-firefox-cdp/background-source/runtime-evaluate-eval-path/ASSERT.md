## Expected

- `background.js` found.
- Method token **`Runtime.evaluate`** present.
- Eval-path reuse: **`handleEvalJob`** and/or **`executeScript`**
  (`scripting.executeScript` / `tabs.executeScript`) via `hasEvalPathReuse`.

## Side Effects

- None.

## Errors

- `Runtime.evaluate` string alone without eval/executeScript reuse fails.
- Debugger-only Runtime.evaluate without executeScript/handleEvalJob fails.

## Exit Code

- Not asserted.

```go
import (
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	src := assertBackgroundPresent(t, req, resp)
	if !hasRuntimeEvaluate(src) {
		t.Fatalf("background missing Runtime.evaluate CDP method token; path=%v snippet=%s",
			resp.FoundPaths, truncate(src, 700))
	}
	if !hasEvalPathReuse(src) {
		t.Fatalf("Runtime.evaluate must reuse eval path (handleEvalJob and/or executeScript); path=%v snippet=%s",
			resp.FoundPaths, truncate(src, 700))
	}
}
```
