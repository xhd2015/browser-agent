## Expected

- Empty FS → `EmbedCompleteFS(..., "session-page")` is **false**.
- Empty FS → `EmbedCompleteFS(..., "extension")` is **false**.
- Response has `BothKinds == true`.

## Side Effects

- None (empty temp dir only).

## Errors

- Either kind reported complete fails this leaf.

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
	if !resp.BothKinds {
		t.Fatal("expected BothKinds=true for empty-fs leaf")
	}
	if resp.CompleteSP {
		t.Fatalf("EmbedCompleteFS(session-page) on empty FS = true, want false; FSRoot=%s",
			resp.FSRoot)
	}
	if resp.CompleteExt {
		t.Fatalf("EmbedCompleteFS(extension) on empty FS = true, want false; FSRoot=%s",
			resp.FSRoot)
	}
	assertExitZero(t, resp)
}
```
