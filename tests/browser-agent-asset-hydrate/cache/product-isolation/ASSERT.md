## Expected

- Write for `browser-agent` succeeds.
- `Open`/`CacheComplete` for **browser-agent** → hit (ok / true).
- `Open`/`CacheComplete` for **browser-trace** (same version + kind) → miss
  (`OtherOpenOK == false`, `OtherCacheComplete == false`).

## Side Effects

- Cache only under writer product path under XDG temp.

## Errors

- Writer miss or other-product hit fails this leaf.

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
	if resp.WriteErr != nil {
		t.Fatalf("WriteAssetCache(browser-agent) err=%v", resp.WriteErr)
	}
	if strings.TrimSpace(resp.WriteDir) == "" {
		t.Fatal("WriteDir empty")
	}
	if !resp.OpenOK || !resp.CacheComplete {
		t.Fatalf("browser-agent open/complete miss after write: openOK=%v complete=%v openErr=%v",
			resp.OpenOK, resp.CacheComplete, resp.OpenErr)
	}
	if resp.OtherOpenOK {
		t.Fatal("OpenAssetCache(browser-trace) ok=true; want product isolation miss")
	}
	if resp.OtherCacheComplete {
		t.Fatal("CacheComplete(browser-trace)=true; want product isolation miss")
	}
	assertExitZero(t, resp)
}
```
