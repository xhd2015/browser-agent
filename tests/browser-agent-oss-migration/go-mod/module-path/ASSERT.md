## Expected

- `go.mod` exists at repo root.
- First non-empty line is exactly `module github.com/xhd2015/browser-agent`.

## Side Effects

- None (read-only).

## Errors

- Missing `go.mod` or wrong module path → leaf fails.

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
		t.Fatalf("read go.mod: %s", resp.RunErr)
	}
	want := "module github.com/xhd2015/browser-agent"
	if resp.GoModFirstLine != want {
		t.Fatalf("go.mod first line = %q, want %q", resp.GoModFirstLine, want)
	}
}
```