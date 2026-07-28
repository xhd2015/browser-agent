## Expected

- `background.js` found.
- Job type token **`logs`** present (quoted / case form preferred).
- Soft markers: result path mentions **`entries`** and/or **`type`** near logs
  (empty buffer language OK). Does **not** require `Log.enable` / debugger.

## Side Effects

- None.

## Errors

- Missing logs branch fails; requiring debugger-only path is out of scope (not asserted as required).

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
	if !jobTypeTokenPresent(src, "logs") {
		t.Fatalf("background missing logs job branch token; path=%v snippet=%s",
			resp.FoundPaths, truncate(src, 600))
	}
	// Best-effort: entries array and/or handleLogsJob name.
	low := strings.ToLower(src)
	hasEntries := strings.Contains(src, "entries") || strings.Contains(src, `"entries"`)
	hasHandler := strings.Contains(src, "handleLogsJob") || strings.Contains(low, "handlelogsjob")
	hasTypeLogs := strings.Contains(src, `"logs"`) || strings.Contains(src, "'logs'")
	if !hasEntries && !hasHandler && !hasTypeLogs {
		t.Fatalf("logs path should mention entries and/or handleLogsJob; path=%v snippet=%s",
			resp.FoundPaths, truncate(src, 600))
	}
}
```
