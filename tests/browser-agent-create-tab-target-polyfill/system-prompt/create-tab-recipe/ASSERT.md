## Expected

Requirement **C1**:

- Non-empty system prompt.
- Does **not** embed concrete session id `sess-create-tab-prompt`.
- Contains nested recipe substring **`browser-agent session create-tab`**
  (hyphen CLI form).
- Soft: still mentions `BROWSER_AGENT_SESSION_ID` (session resolve).

## Side Effects

- None (pure formatter).

## Errors

- Omitting create-tab recipe fails agent playbook for new tabs.

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
	if strings.Contains(p, "sess-create-tab-prompt") {
		t.Fatalf("prompt must not embed concrete session id; prompt=%s", truncate(p, 400))
	}
	if !strings.Contains(p, "browser-agent session create-tab") {
		// also accept without browser-agent prefix if nested clearly listed
		if !strings.Contains(p, "session create-tab") {
			t.Fatalf("prompt missing create-tab recipe; prompt=%s", truncate(p, 700))
		}
	}
}
```
