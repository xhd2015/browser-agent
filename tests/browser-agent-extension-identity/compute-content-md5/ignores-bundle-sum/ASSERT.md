## Expected

**G2 exclude rule**:

- After writing `bundle-sum.js`, second content MD5 equals the first.
- Both digests are 32-hex.
- `bundle-sum.js` path exists on disk after write.

## Side Effects

- `bundle-sum.js` created under temp extension dir.

## Errors

- Digest change after adding only the sum file fails (sum was not excluded).

## Exit Code

- 0.

```go
import (
	"os"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if resp.ErrText != "" {
		t.Fatalf("Compute error: %s", resp.ErrText)
	}
	assertHexMD5(t, resp.MD5First, "MD5First")
	assertHexMD5(t, resp.MD5Second, "MD5Second")
	if resp.MD5First != resp.MD5Second {
		t.Fatalf("content md5 changed after writing bundle-sum.js: before=%q after=%q (must skip sum file)",
			resp.MD5First, resp.MD5Second)
	}
	if resp.BundleSumPath != "" {
		if _, e := os.Stat(resp.BundleSumPath); e != nil {
			t.Fatalf("bundle-sum.js missing at %s: %v", resp.BundleSumPath, e)
		}
	}
	assertExitZero(t, resp)
}
```
