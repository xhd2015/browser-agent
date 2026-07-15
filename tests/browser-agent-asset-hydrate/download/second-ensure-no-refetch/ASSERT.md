## Expected

- First `EnsureAsset` succeeds; second `EnsureAsset` succeeds.
- `GETCount == 1` (second ensure must not issue another GET).
- `CacheCompleteAfter` is true.

## Side Effects

- Only one archive download under isolated XDG.

## Errors

- Either Ensure error, or GETCount != 1, fails this leaf.

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
		t.Fatalf("first EnsureAsset err=%v", resp.EnsureErr)
	}
	if resp.EnsureErr2 != nil {
		t.Fatalf("second EnsureAsset err=%v", resp.EnsureErr2)
	}
	if strings.TrimSpace(resp.EnsureDir) == "" {
		t.Fatal("EnsureDir empty after first ensure")
	}
	if strings.TrimSpace(resp.EnsureDir2) == "" {
		t.Fatal("EnsureDir2 empty after second ensure")
	}
	if !resp.CacheCompleteAfter {
		t.Fatal("CacheComplete=false after ensures")
	}
	if resp.GETCount != 1 {
		t.Fatalf("GETCount=%d want 1 (second ensure must not refetch)", resp.GETCount)
	}
	assertExitZero(t, resp)
}
```
