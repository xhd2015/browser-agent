## Expected

- `SessionNew` succeeds.
- Stdout mentions `open-managed-chrome` (not bare `open-chrome` command).
- Stdout contains session id in context of managed hint line.

## Side Effects

- None beyond stdout.

## Errors

- Missing managed hint fails.

## Exit Code

- Not asserted.

```go
import (
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if resp.SessionNewErr != "" {
		t.Fatalf("SessionNew error: %s", resp.SessionNewErr)
	}
	assertContainsFold(t, resp.Stdout, "open-managed-chrome", req.SessionID)
	assertNotContainsFold(t, resp.Stdout, "browser-agent open-chrome")
}
```