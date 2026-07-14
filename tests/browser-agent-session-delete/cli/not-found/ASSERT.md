## Expected

After implementer lands session delete (**RED** on current code):

- `HandleCLI` returns non-nil error; `ExitCode` 1.
- Error text mentions not found (or unknown session).

## Side Effects

- No session directories created or removed.

## Errors

- Exit 0 or missing not-found message fails.

## Exit Code

- 1.

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
		t.Fatal("session delete timed out")
	}

	assertExitOne(t, resp)
	out := combinedOutput(resp)
	assertContainsFold(t, out, "not found")
}
```