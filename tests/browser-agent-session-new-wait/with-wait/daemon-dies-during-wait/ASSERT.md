## Expected

After implementer lands wait-extension (**RED** on current code):

- Exit code non-zero.
- Error message indicates daemon became unreachable (connection refused, EOF, or similar).
- Elapsed time is short (daemon dies after ~200ms).

## Side Effects

- Session created, daemon killed, wait aborted with error.

## Errors

- Exit code 0 (no error for unreachable daemon) fails.
- Missing error message fails.

## Exit Code

- Non-zero (1).

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

	assertExitNonZero(t, resp)

	if resp.ErrStr == "" {
		t.Fatal("expected non-empty error string for daemon-dies-during-wait")
	}

	// Daemon should die quickly; elapsed should be well under the full timeout.
	if resp.Elapsed > 4*time.Second {
		t.Fatalf("elapsed %v > 4s; daemon should die quickly", resp.Elapsed)
	}
}
```
