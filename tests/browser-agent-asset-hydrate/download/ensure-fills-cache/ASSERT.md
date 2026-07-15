## Expected

- `EnsureAsset` returns nil error and non-empty `EnsureDir`.
- `EnsureDir` is under the leaf XDG temp cache home.
- `CacheCompleteAfter` is true for browser-agent / v0.2.0 / session-page.
- At least one GET was made (`GETCount >= 1`).

## Side Effects

- Cache populated only under isolated XDG temp.

## Errors

- Ensure error, empty dir, or incomplete cache fails this leaf.

## Exit Code

- 0.

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
	if resp.EnsureErr != nil {
		t.Fatalf("EnsureAsset err=%v", resp.EnsureErr)
	}
	if strings.TrimSpace(resp.EnsureDir) == "" {
		t.Fatal("EnsureDir empty")
	}
	assertPathUnder(t, resp.EnsureDir, req.XDGCacheHome)
	if !resp.CacheCompleteAfter {
		t.Fatal("CacheComplete=false after successful EnsureAsset")
	}
	if resp.GETCount < 1 {
		t.Fatalf("GETCount=%d want >= 1", resp.GETCount)
	}
	assertExitZero(t, resp)
}
```
