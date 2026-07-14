## Expected

After implementer lands session-rich (**RED** on current code):

- `HandleCLI` returns nil; `ExitCode` 0.
- Stdout is **human-formatted** (section labels), not JSON-only object.
- Stdout mentions session URL (`/go?session=` or `Session URL`).
- Stdout mentions delete cleanup (`session delete` or delete hint in Next steps).
- Stdout contains session id.

## Side Effects

- Read-only info; no extension.

## Errors

- JSON-only default output fails this leaf.

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
		t.Fatal("session info timed out")
	}

	assertExitZero(t, resp)
	assertStdoutTrailingNewline(t, resp.Stdout)

	if stdoutLooksLikeJSONObject(resp.Stdout) {
		t.Fatalf("default session info must be human output, not JSON-only; stdout=%s",
			truncate(resp.Stdout, 500))
	}

	assertContainsFold(t, resp.Stdout, resp.SessionID)
	assertContainsFold(t, resp.Stdout, "session url", "/go?session=", "session_url")
	assertContainsFold(t, combinedOutput(resp), "delete", "session delete")
}
```