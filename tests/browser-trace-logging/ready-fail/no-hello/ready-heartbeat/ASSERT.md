## Expected

- Exit code ≠ 0 (ready timeout after heartbeats).
- Stderr shows ready-wait **heartbeat** progress while still not ready:
  - tokens such as **waiting**, **left**, heartbeat/stage `no_hello`, or
    repeated “still waiting” style lines
  - at least **two** occurrences of a waiting/heartbeat family token, **or**
    clear multi-line wait progress (count of `waiting` ≥ 2, or `left` ≥ 2)
- Stage language for no hello is acceptable on heartbeats or final message.

## Side Effects

- No successful recording complete.

## Errors

- Ready timeout after wait with heartbeats.

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

	stderr := resp.Stderr
	low := strings.ToLower(stderr)

	// Heartbeat / wait progress tokens (product may use any of these).
	waitingN := countFold(stderr, "waiting")
	leftN := countFold(stderr, "left")
	heartbeatN := countFold(stderr, "heartbeat")
	noHelloN := countFold(stderr, "no_hello")
	stillN := countFold(stderr, "still")

	progressScore := waitingN + leftN + heartbeatN
	if progressScore < 2 && !(waitingN >= 2 || leftN >= 2 || heartbeatN >= 2) {
		// Allow still+waiting combinations counted separately above; if still weak:
		if progressScore+stillN+noHelloN < 2 {
			t.Fatalf("expected ≥2 ready-heartbeat/waiting progress signals on stderr; "+
				"waiting=%d left=%d heartbeat=%d no_hello=%d still=%d; stderr:\n%s",
				waitingN, leftN, heartbeatN, noHelloN, stillN, stderr)
		}
	}

	// Stage should reflect no hello at some point.
	hasStage := strings.Contains(low, "no_hello") ||
		strings.Contains(low, "hello") ||
		strings.Contains(low, "connect")
	if !hasStage {
		t.Fatalf("heartbeat/ready stderr should mention stage no_hello or connect/hello; got:\n%s", stderr)
	}
}
```
