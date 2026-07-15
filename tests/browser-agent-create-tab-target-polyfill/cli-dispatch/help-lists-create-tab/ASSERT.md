## Expected

Requirement **A1**:

- HandleCLI error is **nil** (CLIErr empty).
- Printed help ends with `\n`.
- Lists nested **`create-tab`** (or `session create-tab`) under session commands.
- Soft: still lists `session` / other known nested cmds (not exclusive list).
- `DispatchTimedOut` false.

## Side Effects

- None.

## Errors

- Non-nil error or missing create-tab is a failure.

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
		t.Fatal("HandleCLI timed out on --help")
	}
	if resp.CLIErr != "" {
		t.Fatalf("--help should return nil error; got %q", resp.CLIErr)
	}
	assertExitZero(t, resp)
	assertPrintedTrailingNewline(t, resp)
	text := strings.ToLower(combinedCLIText(resp))
	if !strings.Contains(text, "create-tab") && !strings.Contains(text, "create_tab") {
		t.Fatalf("help must list create-tab under session; got:\n%s",
			truncate(combinedCLIText(resp), 900))
	}
	// Prefer nested form visible to operators.
	if !strings.Contains(text, "session") {
		t.Fatalf("help must mention session; got:\n%s", truncate(combinedCLIText(resp), 900))
	}
}
```
