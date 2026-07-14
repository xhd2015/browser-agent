## Expected

Requirement **A1**:

- `AgentRunSessionID("demo")` equals `browser-agent-sess-demo`.
- Result starts with `browser-agent-sess-`.

## Side Effects

- None (pure).

## Errors

- Missing export / wrong prefix fails.

## Exit Code

- Not asserted (pure).

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
	want := "browser-agent-sess-demo"
	if resp.AgentRunID != want {
		t.Fatalf("AgentRunSessionID(%q)=%q, want %q", req.ControlSessionID, resp.AgentRunID, want)
	}
	if !strings.HasPrefix(resp.AgentRunID, AgentRunSessionIDPrefix) {
		t.Fatalf("agent-run id must start with %q; got %q", AgentRunSessionIDPrefix, resp.AgentRunID)
	}
}
```
