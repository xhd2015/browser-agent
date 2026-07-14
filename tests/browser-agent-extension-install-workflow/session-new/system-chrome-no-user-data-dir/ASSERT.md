## Expected

- `SessionNew` succeeds.
- `LaunchCallCount == 1`.
- `LaunchArgs` has `--new-window` and session URL with `/go` + session id.
- `LaunchArgs` has **no** `--user-data-dir` or `--load-extension`.

## Side Effects

- Session registered on daemon.

## Errors

- Managed chrome argv or zero launch calls fail.

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
	if resp.SessionNewErr != "" {
		t.Fatalf("SessionNew error: %s", resp.SessionNewErr)
	}
	if resp.LaunchCallCount != 1 {
		t.Fatalf("LaunchFn call count=%d, want 1", resp.LaunchCallCount)
	}
	assertArgsNoManagedChrome(t, resp.LaunchArgs)
	assertArgsHasNewWindowAndURL(t, resp.LaunchArgs, req.SessionID)
}
```