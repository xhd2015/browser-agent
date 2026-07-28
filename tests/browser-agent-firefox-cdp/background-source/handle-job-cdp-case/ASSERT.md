## Expected

- `background.js` found under `Firefox-Ext-Browser-Agent`.
- Source has a **dedicated** job-type branch for **`cdp`**:
  - quoted / `case "cdp"` / `jobType === "cdp"` form via `jobTypeTokenPresent`, **or**
  - named handler `handleCdpJob` / `handleCDPJob`.

## Side Effects

- None (read-only FS).

## Errors

- Missing dedicated cdp case (only Phase-2 default unknown) fails.

## Exit Code

- Not asserted.

```go
import (
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	src := assertBackgroundPresent(t, req, resp)
	if !hasDedicatedCdpCase(src) {
		t.Fatalf("background missing dedicated cdp job branch (case \"cdp\" / handleCdpJob); path=%v snippet=%s",
			resp.FoundPaths, truncate(src, 700))
	}
}
```
