## Expected

- Exit code ≠ 0.
- Error / stderr mentions timeout and/or extension not connecting
  (install / enable / host_permissions style guidance acceptable).
- Mock never reached recording; no successful `recording.har` save as a completed session.

## Side Effects

- Session may leave partial state under `BaseDir`, but must not report success.
- No implication that recording completed.

## Errors

- Ready-phase failure (extension never said hello).

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
	// Must signal timeout and/or extension connectivity.
	hasTimeout := strings.Contains(low, "timeout") || strings.Contains(low, "timed out") || strings.Contains(low, "deadline")
	hasExt := strings.Contains(low, "extension") || strings.Contains(low, "hello") || strings.Contains(low, "connect")
	if !hasTimeout && !hasExt {
		t.Fatalf("expected ready-timeout / extension-not-connecting message; got:\n%s", text)
	}
	if resp.StatusRecording {
		t.Fatal("StatusRecording should be false when no hello")
	}
	if resp.CompletePosted {
		t.Fatal("complete must not have been posted")
	}
}
```
