## Expected

- Restore err nil.
- `sess-nomete` not in List; Get false.

## Side Effects

- Dir remains without registration.

## Errors

- Registration of missing-meta dir fails this leaf.

## Exit Code

- 0.

```go
import (
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	assertRestoreOK(t, resp)
	assertListNotContainsID(t, resp, "sess-nomete")
	assertGetOK(t, resp, "sess-nomete", false)
	assertExitZero(t, resp)
}
```
