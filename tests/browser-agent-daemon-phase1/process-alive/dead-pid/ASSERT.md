## Expected

- `IsProcessAlive(999999999)` returns false.

## Side Effects

- None (read-only process check).

## Errors

- ProcessAlive true fails this leaf.

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
	if resp.ProcessAlive {
		t.Fatalf("IsProcessAlive(%d)=true want false", req.PID)
	}
	assertExitZero(t, resp)
}
```