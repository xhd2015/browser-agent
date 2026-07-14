## Expected

Requirement **F1** (updated for nested session CLI):

- Non-empty system prompt.
- Contains each nested recipe substring:
  - `browser-agent session info`
  - `browser-agent session eval`
  - `browser-agent session run`
  - `browser-agent session logs`
  - `browser-agent session screenshot`
- Does **not** embed the concrete session id `sess-system-prompt`.
- Mentions `BROWSER_AGENT_SESSION_ID`.

## Side Effects

- None (pure formatter).

## Errors

- Omitting any nested recipe fails agent playbook completeness.
- Embedding control session id fails open-prompt isolation.

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
	// Concrete control id must not appear in playbook body.
	if strings.Contains(p, "sess-system-prompt") {
		t.Fatalf("prompt must not embed concrete session id; prompt=%s", truncate(p, 400))
	}
	needles := []string{
		"browser-agent session info",
		"browser-agent session eval",
		"browser-agent session run",
		"browser-agent session logs",
		"browser-agent session screenshot",
		"BROWSER_AGENT_SESSION_ID",
	}
	for _, n := range needles {
		if !strings.Contains(p, n) {
			t.Fatalf("prompt missing %q; prompt=%s", n, truncate(p, 600))
		}
	}
}
```
