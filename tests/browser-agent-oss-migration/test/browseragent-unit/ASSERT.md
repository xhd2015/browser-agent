## Expected

- `go test ./browseragent/...` exits 0 from repo root.

## Side Effects

- Test cache only.

## Errors

- Test or compile failures reported in combined output.

## Exit Code

- Must be 0.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	assertRunShellErr(t, resp)
	if resp.ExitCode != 0 {
		t.Fatalf("go test ./browseragent/... exit=%d\nstdout:\n%s\nstderr:\n%s",
			resp.ExitCode, resp.Stdout, resp.Stderr)
	}
}
```