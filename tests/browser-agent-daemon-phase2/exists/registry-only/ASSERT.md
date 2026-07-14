## Expected

- After Create, `Exists("reg-only")` returns true.

## Side Effects

- Session dir and registry entry from Create.

## Errors

- Exists false fails this leaf.

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
	if !resp.Exists {
		t.Fatalf("Exists(%q)=false want true", req.ExistsSessionID)
	}
	assertExitZero(t, resp)
}
```