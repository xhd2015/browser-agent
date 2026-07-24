## Expected

After implementer lands wait-extension (**RED** on current code):

- Exit code 0.
- Elapsed time < 5s (fast connection, not full 30s).
- Stderr contains "Extension connected".
- Stdout contains normal session output (session-id, export, Session URL, Extension path, Next steps).

## Side Effects

- Session created normally.
- Extension connected via WS.

## Errors

- Exit non-zero or missing extension connected message fails.
- Elapsed time ≥ 5s (should connect quickly) fails.
- Stdout missing session-id fails.

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

	if resp.Elapsed >= 5*time.Second {
		t.Fatalf("elapsed %v >= 5s; should connect quickly", resp.Elapsed)
	}

	assertContains(t, resp.Stderr, "Extension connected")

	assertContains(t, resp.Stdout, "session-id:")
	assertContains(t, resp.Stdout, "Session URL:")
	assertContains(t, resp.Stdout, "Extension:")
	assertContains(t, resp.Stdout, "Next:")
	assertContains(t, resp.Stdout, req.SessionID)
}
```
