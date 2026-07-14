## Expected

- `SessionNew` succeeds.
- `LaunchCallCount == 0`.
- `ExtensionPath` non-empty with canonical segment.

## Side Effects

- Extract happens even when chrome open skipped.

## Errors

- Launch invoked or missing canonical path fails.

## Exit Code

- Not asserted.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if resp.SessionNewErr != "" {
		t.Fatalf("SessionNew error: %s", resp.SessionNewErr)
	}
	if resp.LaunchCallCount != 0 {
		t.Fatalf("LaunchFn call count=%d, want 0 when NoOpenChrome", resp.LaunchCallCount)
	}
	if strings.TrimSpace(resp.ExtensionPath) == "" {
		t.Fatal("ExtensionPath empty; extract should still run")
	}
	assertCanonicalPathSegment(t, resp.ExtensionPath)
}
```