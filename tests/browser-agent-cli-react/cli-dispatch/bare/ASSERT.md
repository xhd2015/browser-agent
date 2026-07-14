## Expected

Requirement **A1**:

- `HandleCLI` returns non-nil error (CLIErr non-empty).
- Combined stdout+stderr non-empty; printed body ends with `\n`.
- Mentions command `serve` (case-insensitive OK for product name variants;
  literal `serve` subcommand preferred).
- `DispatchTimedOut` is false.

## Side Effects

- No HTTP listen / Chrome / agent-run.

## Errors

- Non-nil CLI error expected for bare invocation.

## Exit Code

- Non-zero preferred (ExitCode 1 via harness when CLIErr set).

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
		t.Fatal("HandleCLI timed out — must not start long-running serve on bare args")
	}
	if resp.CLIErr == "" {
		t.Fatal("expected non-nil HandleCLI error for bare invocation")
	}
	assertPrintedTrailingNewline(t, resp)
	text := combinedCLIText(resp)
	if !strings.Contains(strings.ToLower(text), "serve") {
		t.Fatalf("brief usage must mention serve; got:\n%s", truncate(text, 600))
	}
}
```
