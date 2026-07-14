## Expected

Requirement **C4**:

- Fake extension observed a job with type `eval`.
- HandleCLI did not time out.
- Prefer nil CLIErr and trailing `\n` on success paths.

## Side Effects

- Temp BaseDir cleaned by harness.

## Errors

- Flat-only dispatch would not post jobs under nested tree until implement lands.

## Exit Code

- 0 preferred on success.

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
	if resp.DispatchTimedOut {
		t.Fatal("session eval timed out")
	}
	if resp.CLIErr != "" {
		t.Fatalf("session eval CLI error: %q stdout=%s stderr=%s",
			resp.CLIErr, truncate(resp.Stdout, 300), truncate(resp.Stderr, 300))
	}
	if !resp.WSJobReceived && resp.ObservedJobType == "" {
		t.Fatalf("fake extension did not observe a job; JobsSeen=%v", resp.JobsSeen)
	}
	got := resp.ObservedJobType
	if got != "eval" {
		t.Fatalf("ObservedJobType=%q, want eval; JobsSeen=%v", got, resp.JobsSeen)
	}
	if resp.Stdout != "" && !strings.HasSuffix(resp.Stdout, "\n") {
		t.Fatalf("stdout should end with \\n; got tail %q", tail(resp.Stdout, 40))
	}
}
```
