## Expected

Requirement **B1**:

- Fake WS observes job type `eval`.
- Params include expression/expr containing `1+1` (or exact `1+1`).
- CLIErr empty; ExitCode 0; stdout ends with `\n`.
- DispatchTimedOut false.

## Side Effects

- One job pushed over WS (JobsSeen ≥ 1 preferred).

## Errors

- Wrong type or missing expression is a failure.

## Exit Code

- 0.

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
		t.Fatal("eval job-type timed out")
	}
	if resp.CLIErr != "" {
		t.Fatalf("eval should succeed; CLIErr=%q stderr=%q stdout=%q",
			resp.CLIErr, resp.Stderr, resp.Stdout)
	}
	assertExitZero(t, resp)
	assertStdoutTrailingNewline(t, resp.Stdout)
	assertObservedJobType(t, resp, "eval")

	expr := paramString(resp.ObservedJobParams, "expression", "expr", "code")
	if !strings.Contains(expr, "1+1") {
		// Also accept raw params JSON mention
		raw := resp.ObservedJobRaw
		if !strings.Contains(raw, "1+1") {
			t.Fatalf("eval params missing expression 1+1; params=%v raw=%s",
				resp.ObservedJobParams, truncate(raw, 400))
		}
	}
}
```
