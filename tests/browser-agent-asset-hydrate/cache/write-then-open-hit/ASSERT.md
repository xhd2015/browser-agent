## Expected

- `WriteAssetCache` succeeds (`WriteErr == nil`); `WriteDir` non-empty and under
  XDG temp.
- `OpenAssetCache` → `OpenOK == true`, `OpenDir` non-empty (same key path as
  write preferred).
- `CacheComplete` → true.
- Second open → `OpenOK2 == true`.

## Side Effects

- Cache files under isolated XDG temp only.

## Errors

- Write/open failure or complete=false fails this leaf.

## Exit Code

- 0.

```go
import (
	"path/filepath"
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if resp.WriteErr != nil {
		t.Fatalf("WriteAssetCache err=%v", resp.WriteErr)
	}
	if strings.TrimSpace(resp.WriteDir) == "" {
		t.Fatal("WriteDir empty")
	}
	assertPathUnder(t, resp.WriteDir, req.XDGCacheHome)

	if !resp.OpenOK {
		t.Fatalf("OpenAssetCache ok=false; dir=%q err=%v", resp.OpenDir, resp.OpenErr)
	}
	if strings.TrimSpace(resp.OpenDir) == "" {
		t.Fatal("OpenDir empty on hit")
	}
	if !resp.CacheComplete {
		t.Fatal("CacheComplete=false after write of complete fixture")
	}
	// Prefer open path equals write path (or under write path).
	if filepath.Clean(resp.OpenDir) != filepath.Clean(resp.WriteDir) {
		// allow either equal or open under write
		rel, rerr := filepath.Rel(filepath.Clean(resp.WriteDir), filepath.Clean(resp.OpenDir))
		if rerr != nil || strings.HasPrefix(rel, "..") {
			// also allow write under open
			rel2, rerr2 := filepath.Rel(filepath.Clean(resp.OpenDir), filepath.Clean(resp.WriteDir))
			if rerr2 != nil || strings.HasPrefix(rel2, "..") {
				t.Fatalf("OpenDir=%q WriteDir=%q not the same cache key path",
					resp.OpenDir, resp.WriteDir)
			}
		}
	}
	if !resp.OpenOK2 {
		t.Fatal("second OpenAssetCache ok=false")
	}
	assertExitZero(t, resp)
}
```
