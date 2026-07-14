## Expected

Requirement **B1**:

- HandleCLI returns nil (CLIErr empty).
- ExitCode 0.
- Stdout non-empty and ends with `\n`.
- Stdout indicates success / result — any of: `"ok"`, `"value"`, `"result"`,
  numeric `2`, or JSON-ish job result text (case-insensitive where applicable).
- DispatchTimedOut false.

## Side Effects

- Job may be audited under BaseDir (optional; not asserted).

## Errors

- Timeout / CLIErr with fake extension auto-complete is a failure.

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
		t.Fatal("eval sidecmd timed out")
	}
	if resp.CLIErr != "" {
		t.Fatalf("eval should succeed; CLIErr=%q stderr=%q stdout=%q",
			resp.CLIErr, resp.Stderr, resp.Stdout)
	}
	assertExitZero(t, resp)
	assertStdoutTrailingNewline(t, resp.Stdout)

	out := strings.ToLower(resp.Stdout)
	// Accept flexible result presentation from HandleCLI.
	okish := strings.Contains(out, "ok") ||
		strings.Contains(out, "value") ||
		strings.Contains(out, "result") ||
		strings.Contains(out, "2") ||
		strings.Contains(out, "true")
	if !okish {
		t.Fatalf("stdout should reflect eval result/ok; stdout=%q", resp.Stdout)
	}
}
```
