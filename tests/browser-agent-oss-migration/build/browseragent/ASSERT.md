## Expected

- `go build ./browseragent/...` exits 0 from repo root.

## Side Effects

- Writes build cache artifacts only.

## Errors

- Compile errors surface in stderr; leaf fails.

## Exit Code

- Must be 0.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	assertRunShellErr(t, resp)
	if resp.ExitCode != 0 {
		t.Fatalf("go build ./browseragent/... exit=%d stderr:\n%s", resp.ExitCode, resp.Stderr)
	}
}
```