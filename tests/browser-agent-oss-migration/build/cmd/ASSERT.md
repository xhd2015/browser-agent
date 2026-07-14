## Expected

- `go build -o /dev/null ./cmd/browser-agent/` exits 0 from repo root.

## Side Effects

- Build cache only; no installed binary.

## Errors

- Link/compile failure → leaf fails with stderr.

## Exit Code

- Must be 0.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	assertRunShellErr(t, resp)
	if resp.ExitCode != 0 {
		t.Fatalf("go build cmd/browser-agent exit=%d stderr:\n%s", resp.ExitCode, resp.Stderr)
	}
}
```