## Expected

Requirement **B1**:

- Non-empty system prompt.
- Contains nested recipe substrings:
  - `browser-agent session info`
  - `browser-agent session eval`
  - `browser-agent session run`
  - `browser-agent session logs`
  - `browser-agent session screenshot`
  - `browser-agent session cdp`
- Does **not** require flat `browser-agent info` (without `session`).

## Side Effects

- None (pure).

## Errors

- Flat-only recipes fail complete-refactor playbook.

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
	needles := []string{
		"browser-agent session info",
		"browser-agent session eval",
		"browser-agent session run",
		"browser-agent session logs",
		"browser-agent session screenshot",
		"browser-agent session cdp",
	}
	for _, n := range needles {
		if !strings.Contains(p, n) {
			t.Fatalf("prompt missing nested recipe %q; prompt=%s", n, truncate(p, 700))
		}
	}
}
```
