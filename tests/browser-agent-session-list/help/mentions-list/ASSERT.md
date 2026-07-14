## Expected

After implementer updates help (**RED** on current code):

- `HandleCLI session --help` returns nil; `ExitCode` 0.
- Help text contains `session list` or session subcommand enumeration includes `list`.

## Side Effects

- Read-only.

## Errors

- Missing list in help text fails.

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
	hasSessionList := strings.Contains(text, "session list") ||
		strings.Contains(text, "screenshot|cdp|list") ||
		strings.Contains(text, "cdp|list") ||
		strings.Contains(text, "session new|info|delete|eval|run|logs|screenshot|cdp|list")
	if !hasSessionList {
		t.Fatalf("help must document session list; got:\n%s", truncate(resp.HelpText, 800))
	}
}```
