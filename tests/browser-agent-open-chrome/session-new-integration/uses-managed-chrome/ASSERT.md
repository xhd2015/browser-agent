## Expected

- `SessionNew` returns nil error.
- `LaunchCallCount == 1`.
- `LaunchArgs` have `--new-window` and session URL with `/go` and session id.
- `LaunchArgs` have **no** `--user-data-dir` or `--load-extension` (system Chrome).

## Side Effects

- Session created on daemon; system Chrome launch spied (not real browser).

## Errors

- SessionNew error, zero LaunchFn calls, or managed argv fails.

## Exit Code

- N/A (package API).

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("SessionNew integration error: %v", err)
	}
	if resp.SessionNewErr != "" {
		t.Fatalf("SessionNewErr = %q", resp.SessionNewErr)
	}
	if resp.LaunchCallCount != 1 {
		t.Fatalf("LaunchCallCount = %d, want 1 (SessionNew opens system Chrome)", resp.LaunchCallCount)
	}
	assertArgsNoManagedChrome(t, resp.LaunchArgs)
	joined := strings.Join(resp.LaunchArgs, " ")
	if !strings.Contains(joined, "/go") {
		t.Fatalf("LaunchArgs should include session /go URL; args=%v", resp.LaunchArgs)
	}
	if !strings.Contains(joined, req.SessionID) {
		t.Fatalf("LaunchArgs should include session id %q; args=%v", req.SessionID, resp.LaunchArgs)
	}
}
```