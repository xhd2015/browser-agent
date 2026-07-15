## Expected

- `ResolveSessionPage` err nil.
- `Source == "cache"`.
- HTML non-empty with session root marker.
- `GETCount >= 1`.
- `CacheCompleteAfter` true for session-page.

## Side Effects

- Cache under isolated XDG only.

## Errors

- Resolve error, source≠cache, empty HTML, no GET, or incomplete cache fails.

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
	if resp.Source != ResolveSourceCache {
		t.Fatalf("source=%q want %q", resp.Source, ResolveSourceCache)
	}
	assertHTMLHasSessionRoot(t, resp.HTML)
	if resp.GETCount < 1 {
		t.Fatalf("GETCount=%d want >= 1", resp.GETCount)
	}
	if !resp.CacheCompleteAfter {
		t.Fatal("CacheComplete=false after implicit ensure+serve")
	}
	assertExitZero(t, resp)
}
```
