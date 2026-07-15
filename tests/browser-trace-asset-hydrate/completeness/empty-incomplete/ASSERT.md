## Expected

- Empty FS → `EmbedCompleteFS(..., "extension")` is **false**.

## Side Effects

- None (empty temp dir only).

## Errors

- Extension reported complete fails this leaf.

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
	if resp.Complete {
		t.Fatalf("EmbedCompleteFS(extension) on empty FS = true, want false; FSRoot=%s",
			resp.FSRoot)
	}
	assertExitZero(t, resp)
}
```
