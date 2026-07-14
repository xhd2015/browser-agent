## Expected

- Both calls succeed.
- `ExtensionPath == ExtensionPath2`.
- `ExtensionVer == ExtensionVer2`.

## Side Effects

- Single version directory under canonical layout.

## Errors

- Path or version drift on second call fails.

## Exit Code

- Not asserted.

```go
import (
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if resp.ExtensionPath != resp.ExtensionPath2 {
		t.Fatalf("path drift: first=%q second=%q", resp.ExtensionPath, resp.ExtensionPath2)
	}
	if resp.ExtensionVer != resp.ExtensionVer2 {
		t.Fatalf("version drift: first=%q second=%q", resp.ExtensionVer, resp.ExtensionVer2)
	}
}
```