## Expected

- `IsProcessAlive(os.Getpid())` returns true.

## Side Effects

- None (read-only process check).

## Errors

- ProcessAlive false fails this leaf.

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
	if !resp.ProcessAlive {
		t.Fatal("IsProcessAlive(self)=false want true")
	}
	assertExitZero(t, resp)
}
```