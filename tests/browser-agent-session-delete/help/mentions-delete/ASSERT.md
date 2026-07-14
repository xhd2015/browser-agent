## Expected

After implementer updates help (**RED** on current code):

- `HandleCLI session --help` returns nil; `ExitCode` 0.
- Help text contains `session delete` or `delete` documented under session subcommands.

## Side Effects

- Read-only.

## Errors

- Missing delete in help text fails.

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
		t.Fatal("session --help timed out")
	}
	if resp.CLIErr != "" {
		t.Fatalf("session --help should return nil error; got %q", resp.CLIErr)
	}
	assertExitZero(t, resp)

	text := strings.ToLower(resp.HelpText)
	if !strings.Contains(text, "session delete") && !strings.Contains(text, "delete") {
		t.Fatalf("help must document session delete; got:\n%s", truncate(resp.HelpText, 800))
	}
}
```