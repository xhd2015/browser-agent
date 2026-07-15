## Expected

Requirement **B1**:

- Fake WS observes job type **`create_tab`** (underscore; not hyphen).
- CLIErr empty; ExitCode 0; stdout ends with `\n`.
- DispatchTimedOut false.
- Soft: params may omit url or pass empty/about:blank; do not require active field.

## Side Effects

- One job pushed over WS (JobsSeen ≥ 1 preferred).

## Errors

- Wrong type (e.g. cdp / create-tab hyphen) or CLI failure is a failure.

## Exit Code

- 0.

```go
import (
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if resp.DispatchTimedOut {
		t.Fatal("create-tab blank timed out")
	}
	if resp.CLIErr != "" {
		t.Fatalf("create-tab should succeed; CLIErr=%q stderr=%q stdout=%q",
			resp.CLIErr, resp.Stderr, resp.Stdout)
	}
	assertExitZero(t, resp)
	assertStdoutTrailingNewline(t, resp.Stdout)
	assertObservedJobType(t, resp, "create_tab")
}
```
