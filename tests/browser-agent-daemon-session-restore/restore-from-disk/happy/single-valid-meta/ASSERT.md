## Expected

- `RestoreSessionsFromDisk` returns nil.
- `Get("sess-alpha")` is true.
- List contains exactly `sess-alpha`.
- Phase is `waiting_extension`; `extension.connected` is false.

## Side Effects

- Disk `sessions/sess-alpha/` remains (restore does not rewrite via Create).

## Errors

- Non-nil restore error, missing registration, or wrong phase fails this leaf.

## Exit Code

- 0.

```go
import (
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	assertRestoreOK(t, resp)
	assertGetOK(t, resp, "sess-alpha", true)
	assertListContainsID(t, resp, "sess-alpha")
	if len(resp.ListIDs) != 1 {
		t.Fatalf("ListIDs len=%d want 1; got %v", len(resp.ListIDs), resp.ListIDs)
	}
	assertWaitingExtension(t, resp, "sess-alpha")
	assertExitZero(t, resp)
}
```
