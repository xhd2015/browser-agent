## Expected

Requirement **C1** (nested session CLI):

- Non-empty system prompt.
- Does **not** embed concrete session id `sess-cdp-prompt`.
- Contains each nested recipe substring:
  - `browser-agent session info`
  - `browser-agent session eval`
  - `browser-agent session run`
  - `browser-agent session logs`
  - `browser-agent session screenshot`
  - `browser-agent session cdp`
- Mentions `BROWSER_AGENT_SESSION_ID`.

## Side Effects

- None (pure formatter).

## Errors

- Omitting any nested recipe (especially cdp) fails playbook completeness.

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
	if strings.Contains(p, "sess-cdp-prompt") {
		t.Fatalf("prompt must not embed concrete session id; prompt=%s", truncate(p, 400))
	}
	needles := []string{
		"browser-agent session info",
		"browser-agent session eval",
		"browser-agent session run",
		"browser-agent session logs",
		"browser-agent session screenshot",
		"browser-agent session cdp",
		"BROWSER_AGENT_SESSION_ID",
	}
	for _, n := range needles {
		if !strings.Contains(p, n) {
			t.Fatalf("prompt missing %q; prompt=%s", n, truncate(p, 700))
		}
	}
}
```
