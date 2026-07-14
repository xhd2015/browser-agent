## Expected

- Exit code ≠ 0.
- Error indicates recording never started / ready timeout after extension contact
  (message may mention recording, ready, timeout, or extension).
- No successful complete save.

## Side Effects

- Hello may have been accepted; still not a saved recording session.

## Errors

- Ready-phase failure after hello without `recording` status.

## Exit Code

- Non-zero.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertExitNonZero(t, resp)
	text := combinedErrText(resp)
	low := strings.ToLower(text)
	hasSignal := strings.Contains(low, "record") ||
		strings.Contains(low, "timeout") ||
		strings.Contains(low, "ready") ||
		strings.Contains(low, "extension")
	if !hasSignal {
		t.Fatalf("expected ready/recording failure message; got:\n%s", text)
	}
	if resp.CompletePosted {
		t.Fatal("complete must not have been posted on ready failure")
	}
	// Prefer not treating as full success artifacts.
	if resp.ExitCode == 0 {
		t.Fatal("exit code 0 on hello-without-recording")
	}
}
```
