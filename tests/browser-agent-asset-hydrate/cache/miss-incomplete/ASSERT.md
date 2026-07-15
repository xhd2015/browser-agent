## Expected

- `CacheComplete` is **false** for the empty key.
- `OpenAssetCache` returns `ok=false`.
- Open may return nil err (clean miss) or a descriptive error — either OK as
  long as `ok=false`.

## Side Effects

- No successful cache population.

## Errors

- true complete or open ok=true fails this leaf.

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
	if resp.CacheComplete {
		t.Fatalf("CacheComplete=true on empty cache; dir=%q", resp.CacheDir)
	}
	if resp.OpenOK {
		t.Fatalf("OpenAssetCache ok=true on empty cache; dir=%q", resp.OpenDir)
	}
	assertExitZero(t, resp)
}
```
