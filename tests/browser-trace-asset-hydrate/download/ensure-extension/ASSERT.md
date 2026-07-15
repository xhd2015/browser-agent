## Expected

- `EnsureAsset` returns nil error and non-empty `EnsureDir`.
- `EnsureDir` is under the leaf XDG temp cache home.
- `EnsureDir` path contains product segment `browser-trace` and kind `extension`.
- `CacheCompleteAfter` is true for browser-trace / v0.2.0 / extension.
- At least one GET was made (`GETCount >= 1`).
- Request path includes `browser-trace` and `extension` and `.tar.gz`.

## Side Effects

- Cache populated only under isolated XDG temp under product `browser-trace`.

## Errors

- Ensure error, empty dir, incomplete cache, or wrong product/kind path fails.

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
	lowDir := strings.ToLower(resp.EnsureDir)
	if !strings.Contains(lowDir, "browser-trace") {
		t.Fatalf("EnsureDir missing product browser-trace: %s", resp.EnsureDir)
	}
	if !strings.Contains(lowDir, "extension") {
		t.Fatalf("EnsureDir missing kind extension: %s", resp.EnsureDir)
	}
	if !resp.CacheCompleteAfter {
		t.Fatal("CacheComplete=false after successful EnsureAsset")
	}
	if resp.GETCount < 1 {
		t.Fatalf("GETCount=%d want >= 1", resp.GETCount)
	}
	lowPath := strings.ToLower(resp.LastRequestPath)
	if !strings.Contains(lowPath, "browser-trace") {
		t.Fatalf("LastRequestPath missing browser-trace: %s", resp.LastRequestPath)
	}
	if !strings.Contains(lowPath, "extension") {
		t.Fatalf("LastRequestPath missing extension: %s", resp.LastRequestPath)
	}
	if !strings.Contains(lowPath, ".tar.gz") {
		t.Fatalf("LastRequestPath missing .tar.gz: %s", resp.LastRequestPath)
	}
	assertExitZero(t, resp)
}
```
