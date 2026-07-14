## Expected

Requirement **C1**:

- Serve healthy.
- `AgentRunCallCount == 0` even though AgentRunFn was set.

## Side Effects

- No agent-run process; injector must not run.

## Errors

- Calling injector when NoAgentRun=true is a contract violation.

## Exit Code

- Not asserted.

```go
import (
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if resp.AgentRunCallCount != 0 {
		t.Fatalf("AgentRunFn called %d times; want 0 when NoAgentRun=true (sid=%q sys=%q)",
			resp.AgentRunCallCount, resp.AgentRunSessionID, resp.AgentRunSystemPath)
	}
}
```
