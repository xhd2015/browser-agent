## Expected

Requirement **B3**:

- Non-empty system prompt.
- Mentions `BROWSER_AGENT_SESSION_ID` as env source for session resolve.

## Side Effects

- None (pure).

## Errors

- Omitting env docs breaks agent side-command recipes after nested refactor.

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
	p := resp.SystemPrompt
	if strings.TrimSpace(p) == "" {
		t.Fatal("SystemPrompt empty")
	}
	if !strings.Contains(p, "BROWSER_AGENT_SESSION_ID") {
		t.Fatalf("prompt must mention BROWSER_AGENT_SESSION_ID; prompt=%s", truncate(p, 600))
	}
}
```
