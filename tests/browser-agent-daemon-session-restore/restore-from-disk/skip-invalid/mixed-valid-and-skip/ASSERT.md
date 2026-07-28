## Expected

- Restore err nil.
- List contains only `sess-good` (waiting_extension).
- `sess-badmeta` not registered.

## Side Effects

- Both dirs remain on disk.

## Errors

- Missing valid registration or registering corrupt id fails this leaf.

## Exit Code

- 0.

```go
import (
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	assertRestoreOK(t, resp)
	assertListContainsID(t, resp, "sess-good")
	assertListNotContainsID(t, resp, "sess-badmeta")
	assertGetOK(t, resp, "sess-good", true)
	assertGetOK(t, resp, "sess-badmeta", false)
	if len(resp.ListIDs) != 1 {
		t.Fatalf("ListIDs=%v want exactly [sess-good]", resp.ListIDs)
	}
	assertWaitingExtension(t, resp, "sess-good")
	assertExitZero(t, resp)
}
```
