## Expected

Requirement **B4**:

- Fake WS observes job type `screenshot`.
- CLIErr empty; ExitCode 0; stdout ends with `\n`.

## Side Effects

- None required (`-o` write not asserted).

## Errors

- Wrong type is a failure.

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
		t.Fatal("screenshot job-type timed out")
	}
	if resp.CLIErr != "" {
		t.Fatalf("screenshot should succeed; CLIErr=%q stderr=%q stdout=%q",
			resp.CLIErr, resp.Stderr, resp.Stdout)
	}
	assertExitZero(t, resp)
	assertStdoutTrailingNewline(t, resp.Stdout)
	assertObservedJobType(t, resp, "screenshot")
}
```
