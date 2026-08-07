## Expected

After implementer lands wait-extension (**RED** on current code):

- Exit code 0.
- Stderr contains timeout warning ("extension did not connect within").
- Stderr contains user-handling banner:
  `Please run or ask user to run manually: this needs user handling`
- Stderr contains install command (`install-chrome-extension`) and Load unpacked guidance.
- Stdout contains normal session output (session-id, Session URL, Extension path, Next steps).
- Elapsed time is approximately ~3s (the configured timeout; may be slightly more due to polling).

## Side Effects

- Session created normally despite no extension.

## Errors

- Exit non-zero on timeout fails (timeout should be graceful).
- Missing session output on stdout fails.
- Missing stderr warning or user-handling / install help fails.

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

	// Elapsed should be approximately the timeout (3s) plus a small polling margin.
	if resp.Elapsed < 2*time.Second {
		t.Fatalf("elapsed %v < 2s; should have waited near ~3s timeout", resp.Elapsed)
	}

	assertContains(t, resp.Stderr, "extension did not connect within")
	assertContains(t, resp.Stderr, "Please run or ask user to run manually: this needs user handling")
	assertContains(t, resp.Stderr, "install-chrome-extension")
	assertContains(t, resp.Stderr, "Load unpacked")

	// Stdout should still have normal session output.
	assertContains(t, resp.Stdout, "session-id:")
	assertContains(t, resp.Stdout, "Session URL:")
	assertContains(t, resp.Stdout, req.SessionID)
}
```
