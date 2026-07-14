## Expected

- After Create, `Get("get-hit")` returns ok true.

## Side Effects

- Session created during Run pre-step.

## Errors

- GetOK false fails this leaf.

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
	if !resp.GetOK {
		t.Fatalf("Get(%q) ok=false want true", req.GetSessionID)
	}
	assertExitZero(t, resp)
}
```