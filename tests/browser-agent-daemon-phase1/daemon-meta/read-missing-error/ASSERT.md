## Expected

- `ReadDaemonMeta` on missing file returns non-nil error.

## Side Effects

- None (read only).

## Errors

- Nil ReadErr fails this leaf.

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
	if resp.ReadErr == nil {
		t.Fatal("ReadDaemonMeta(missing) err=nil want non-nil")
	}
	assertExitZero(t, resp)
}
```