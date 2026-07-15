## Expected

- `EnsureAsset` returns **non-nil** error.
- Error message looks like download/HTTP failure (substring case-insensitive:
  `404`, `not found`, `download`, `http`, `status`, `connect`, `refused`, or
  `ensure`).
- `CacheCompleteAfter` is **false** (must not mark incomplete as complete).

## Side Effects

- No successful complete cache promotion under XDG temp.

## Errors

- nil Ensure error or CacheComplete true fails this leaf.

## Exit Code

- 0 (harness); Ensure error is asserted on Response.

```go
import (
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if resp.EnsureErr == nil {
		t.Fatalf("EnsureAsset err=nil want download/404 error; dir=%q", resp.EnsureDir)
	}
	if !downloadErrorMessage(resp.EnsureErr) {
		t.Fatalf("EnsureAsset error %q does not look like download/HTTP failure",
			resp.EnsureErr.Error())
	}
	if resp.CacheCompleteAfter {
		t.Fatal("CacheComplete=true after failed EnsureAsset; must not promote incomplete cache")
	}
	assertExitZero(t, resp)
}
```
