## Expected

- Serve starts and becomes healthy (Run completes without transport error).
- `LaunchCallCount == 0`.
- `OpenChromeCallCount == 0`.

## Side Effects

- Daemon listens; no chrome spawn.

## Errors

- Any hook invocation fails.

## Exit Code

- Not asserted (cancelled daemon may return context error).

```go
import (
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if resp.LaunchCallCount != 0 {
		t.Fatalf("LaunchFn called %d times; want 0 for plain serve", resp.LaunchCallCount)
	}
	if resp.OpenChromeCallCount != 0 {
		t.Fatalf("OpenChromeFn called %d times; want 0 for plain serve", resp.OpenChromeCallCount)
	}
}
```