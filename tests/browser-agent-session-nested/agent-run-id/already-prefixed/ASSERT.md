## Expected

Requirement **A2**:

- Input already prefixed → returned unchanged.
- No double prefix (`browser-agent-sess-browser-agent-sess-…`).

## Side Effects

- None (pure).

## Errors

- Double prefix breaks agent-run session continuity.

## Exit Code

- Not asserted.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	in := req.ControlSessionID
	if resp.AgentRunID != in {
		t.Fatalf("already-prefixed must be idempotent: got %q want %q", resp.AgentRunID, in)
	}
	double := AgentRunSessionIDPrefix + AgentRunSessionIDPrefix
	if strings.HasPrefix(resp.AgentRunID, double) {
		t.Fatalf("double prefix detected: %q", resp.AgentRunID)
	}
	// Exactly one prefix occurrence at start.
	rest := strings.TrimPrefix(resp.AgentRunID, AgentRunSessionIDPrefix)
	if strings.HasPrefix(rest, AgentRunSessionIDPrefix) {
		t.Fatalf("double prefix after trim: %q", resp.AgentRunID)
	}
}
```
