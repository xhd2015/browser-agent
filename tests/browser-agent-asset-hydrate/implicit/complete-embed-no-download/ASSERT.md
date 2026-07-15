## Expected

- `ResolveSessionPage` err nil.
- `Source == "embed"`.
- HTML non-empty with session root marker.
- `GETCount == 0` (no download when embed complete).

## Side Effects

- No cache write required; none under XDG expected for this leaf.

## Errors

- Non-nil resolve err, wrong source, missing HTML marker, or GETCount≠0 fails.

## Exit Code

- 0.

```go
import (
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if resp.ResolveErr != nil {
		t.Fatalf("ResolveSessionPage err=%v", resp.ResolveErr)
	}
	if resp.Source != ResolveSourceEmbed {
		t.Fatalf("source=%q want %q", resp.Source, ResolveSourceEmbed)
	}
	assertHTMLHasSessionRoot(t, resp.HTML)
	if resp.GETCount != 0 {
		t.Fatalf("GETCount=%d want 0 when embed complete", resp.GETCount)
	}
	assertExitZero(t, resp)
}
```
