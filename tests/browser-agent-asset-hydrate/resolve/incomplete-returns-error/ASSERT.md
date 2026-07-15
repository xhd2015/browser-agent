## Expected

- `ResolveSessionPageIndexFS` on empty FS returns **non-nil error**.
- Error message (case-insensitive) contains one of:
  - `incomplete`
  - `not available`
  - `embed incomplete`
  - `not complete`
- Must **not** pretend success: if err is non-nil, OK; if err is nil, fail
  even when HTML is empty.

## Side Effects

- None. No network, no cache writes.

## Errors

- nil resolve error (success path) fails this leaf.

## Exit Code

- 0 (harness); resolve API error is asserted in Response, not as Run err.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if resp.ResolveErr == nil {
		t.Fatalf("ResolveSessionPageIndexFS err=nil want incomplete/not-available error; source=%q html=%s FSRoot=%s",
			resp.Source, truncate(resp.HTML, 200), resp.FSRoot)
	}
	if !incompleteErrorMessage(resp.ResolveErr) {
		t.Fatalf("resolve error %q does not look like incomplete/not-available; want substring incomplete|not available|embed incomplete|not complete",
			resp.ErrText)
	}
	// Do not treat empty success as OK — already failed if err nil.
	// If err non-nil but source claims embed, still OK as long as error is clear.
	if resp.Source == ResolveSourceEmbed && strings.TrimSpace(resp.HTML) != "" && resp.ResolveErr == nil {
		t.Fatal("incomplete path must not return source=embed with non-empty HTML and nil error")
	}
	assertExitZero(t, resp)
}
```
