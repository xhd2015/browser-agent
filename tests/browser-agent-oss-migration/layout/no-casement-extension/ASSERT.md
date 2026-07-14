## Expected

- `Chrome-Ext-Casement-Token/` does **not** exist at repo root.

## Side Effects

- None.

## Errors

- Directory present → leaf fails.

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
		t.Fatalf("stat casement extension: %s", resp.RunErr)
	}
	if resp.CasementExtAbs != "" {
		t.Fatalf("Chrome-Ext-Casement-Token must be absent; found %s", resp.CasementExtAbs)
	}
}
```