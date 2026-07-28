## Expected

Explicit anti-generic contract for Phase 3 CDP:

1. `background.js` found.
2. **`isCdpOnlyGenericUnknown(src)` must be false**:
   - dedicated `cdp` case / `handleCdpJob` present, **and**
   - method matrix marker (`Page.navigate` or `Runtime.evaluate`) present.
3. Must **not** rely solely on Phase-2 default unknown:
   `"not implemented: firefox job type=" + jobType` for cdp.

## Side Effects

- None.

## Errors

- Phase-2 default unknown handling of cdp **fails** this leaf.

## Exit Code

- Not asserted.

```go
import (
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	src := assertBackgroundPresent(t, req, resp)

	if isCdpOnlyGenericUnknown(src) {
		t.Fatalf("cdp still only generic unknown job type (need dedicated case \"cdp\" + method matrix); path=%v snippet=%s",
			resp.FoundPaths, truncate(src, 800))
	}
	if !hasDedicatedCdpCase(src) {
		t.Fatalf("dedicated cdp branch missing; path=%v", resp.FoundPaths)
	}
	if !hasPageNavigate(src) && !hasRuntimeEvaluate(src) {
		t.Fatalf("cdp method matrix missing Page.navigate and Runtime.evaluate; path=%v", resp.FoundPaths)
	}
}
```
