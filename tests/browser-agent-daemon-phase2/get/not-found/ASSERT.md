## Expected

- `Get("no-such-session")` returns ok false on fresh registry.

## Side Effects

- None.

## Errors

- GetOK true fails this leaf.

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
	if resp.GetOK {
		t.Fatalf("Get(%q) ok=true want false", req.GetSessionID)
	}
	assertExitZero(t, resp)
}
```