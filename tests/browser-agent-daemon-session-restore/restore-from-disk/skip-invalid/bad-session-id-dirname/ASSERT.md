## Expected

- Restore err nil (best-effort).
- List empty (or does not contain `-badid`).
- Get(`-badid`) false.

## Side Effects

- Invalid dir may remain on disk; not registered.

## Errors

- Hard restore failure or unexpected registration fails this leaf.

## Exit Code

- 0.

```go
import (
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	assertRestoreOK(t, resp)
	assertListNotContainsID(t, resp, "-badid")
	assertGetOK(t, resp, "-badid", false)
	if len(resp.ListIDs) != 0 {
		t.Fatalf("ListIDs=%v want empty", resp.ListIDs)
	}
	assertExitZero(t, resp)
}
```
