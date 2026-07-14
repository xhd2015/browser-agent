## Expected

- `RemoveDaemonMeta` on non-existent path returns nil.

## Side Effects

- None (file already absent).

## Errors

- Non-nil RemoveErr fails this leaf.

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
	if resp.RemoveErr != nil {
		t.Fatalf("RemoveDaemonMeta(missing) err=%v want nil", resp.RemoveErr)
	}
	assertExitZero(t, resp)
}
```