## Expected

After implementer lands wait-extension (**RED** on current code):

- Exit code 0.
- Elapsed time < 1s (no waiting at all).
- Stdout contains normal session output (session-id, Session URL, Extension path, Next steps).
- Stderr has NO extension connected message and NO warning.

## Side Effects

- Session created without waiting.

## Errors

- Exit non-zero fails.
- Elapsed ≥ 1s fails (should return immediately).
- Any extension-related messages in stderr fails.

## Exit Code

- 0.

```go
import (
	"testing"
	"time"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}

	assertExitZero(t, resp)

	if resp.Elapsed >= 1*time.Second {
		t.Fatalf("elapsed %v >= 1s; should return immediately with --no-open-chrome", resp.Elapsed)
	}

	assertContains(t, resp.Stdout, "session-id:")
	assertContains(t, resp.Stdout, req.SessionID)

	// No extension messages in output when skipping wait.
	assertNotContainsFold(t, combinedOutput(resp), "extension connected", "extension did not connect")
}
```
