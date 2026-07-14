## Expected

After implementer lands session-rich (**RED** on current code):

- `HandleCLI` returns nil; `ExitCode` 0.
- Stdout table headers include `Created`, `Pages`, `Browser`, and `Status`.
- Stdout contains session id `sess-rich-list-zero`.
- Stdout or stderr footer hints 0-page cleanup mentioning `session delete` or `delete`.

## Side Effects

- Read-only list after hello telemetry.

## Errors

- Old Phase/Connected-only columns without new columns fails.
- Missing delete hint for 0-page session fails.

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
		t.Fatal("session list timed out")
	}

	assertExitZero(t, resp)
	assertStdoutTrailingNewline(t, resp.Stdout)

	assertContainsFold(t, resp.Stdout, "created", "pages", "browser", "status")
	assertContainsFold(t, resp.Stdout, req.SessionID)

	out := combinedOutput(resp)
	// Footer hint for 0-page sessions should suggest session delete cleanup.
	assertContainsFold(t, out, "delete")
}
```