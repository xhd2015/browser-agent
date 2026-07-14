## Expected

Requirement **C3**:

- `AgentPort` is **43761**.
- `TracePort` is **43759**.
- `AgentPort != TracePort`.
- Agent/Trace ids are distinct when populated.

## Side Effects

- None.

## Errors

- Equal ports would make dual products collide on localhost.

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
	if resp.AgentPort != 43761 {
		t.Fatalf("AgentPort = %d, want 43761", resp.AgentPort)
	}
	if resp.TracePort != 43759 {
		t.Fatalf("TracePort = %d, want 43759", resp.TracePort)
	}
	if resp.AgentPort == resp.TracePort {
		t.Fatalf("AgentPort and TracePort both %d; must differ", resp.AgentPort)
	}
	if resp.AgentID != "" && resp.TraceID != "" &&
		strings.EqualFold(resp.AgentID, resp.TraceID) {
		t.Fatalf("AgentID and TraceID both %q; must be distinct", resp.AgentID)
	}
}
```
