## Expected

Requirement **B2** (serve never opens Chrome per extension-install workflow):

- Serve healthy.
- `OpenChromeCallCount == 0` even when `NoOpenChrome=false` and hook is set.

## Side Effects

- Injector not invoked; no real Chrome binary.

## Errors

- Non-zero OpenChrome calls means serve incorrectly launched Chrome.

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
		t.Fatalf("OpenChromeFn call count=%d, want 0 (serve must not launch Chrome)", resp.OpenChromeCallCount)
	}
}
```