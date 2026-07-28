## Expected

Soft optional polyfill:

1. `background.js` found.
2. If **`Target.createTarget`** is **absent** → **PASS** (optional not required).
3. If **`Target.createTarget`** is **present** → must also have create_tab path
   (`tabs.create` / `handleCreateTabJob` / `createTabInSession` / `"create_tab"`).

## Side Effects

- None.

## Errors

- Target.createTarget without create_tab coupling fails.
- Missing Target.createTarget alone does **not** fail.

## Exit Code

- Not asserted.

```go
import (
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	src := assertBackgroundPresent(t, req, resp)

	if !hasTargetCreateTarget(src) {
		// Optional soft polyfill not implemented — acceptable for Phase 3.
		t.Log("Target.createTarget soft polyfill not present (optional) — OK")
		return
	}
	if !hasCreateTabPath(src) {
		t.Fatalf("Target.createTarget present but create_tab/tabs.create path missing; path=%v snippet=%s",
			resp.FoundPaths, truncate(src, 700))
	}
}
```
