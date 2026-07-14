## Expected

Requirement **B1**:

- Serve healthy; no transport error.
- `OpenChromeCallCount == 0` even though OpenChromeFn was set.

## Side Effects

- No Chrome process; injector must not run.

## Errors

- Calling injector when NoOpenChrome=true is a contract violation.

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
	if resp.OpenChromeCallCount != 0 {
		t.Fatalf("OpenChromeFn called %d times; want 0 when NoOpenChrome=true (url=%q path=%q)",
			resp.OpenChromeCallCount, resp.OpenChromeSessionURL, resp.OpenChromeExtPath)
	}
}
```
