## Expected

- Restore err nil.
- `sess-badmeta` not registered.

## Side Effects

- Corrupt meta.json left on disk.

## Errors

- Hard failure or accidental registration fails this leaf.

## Exit Code

- 0.

```go
import (
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	assertRestoreOK(t, resp)
	assertListNotContainsID(t, resp, "sess-badmeta")
	assertGetOK(t, resp, "sess-badmeta", false)
	assertExitZero(t, resp)
}
```
