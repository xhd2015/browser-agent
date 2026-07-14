## Expected

- No directory whose relative path is `server/api/casement` anywhere under repo root.

## Side Effects

- None.

## Errors

- Any `server/api/casement/` directory → leaf fails with its absolute path.

## Exit Code

- Not asserted.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if resp.RunErr != "" {
		t.Fatalf("walk for casement server: %s", resp.RunErr)
	}
	if resp.CasementSrvAbs != "" {
		t.Fatalf("server/api/casement must be absent; found %s", resp.CasementSrvAbs)
	}
}
```