## Expected

Requirement **C2** (nested session skill docs):

- HandleCLI nil error; exit 0.
- Stdout is skill body (non-empty).
- Stdout contains:
  - `browser-agent`
  - `BROWSER_AGENT_SESSION_ID`
  - `session` (nested side-command parent)
  - `eval` (side command)
  - `43761` (control port)
- Stdout ends with trailing `\n`.

## Side Effects

- Read-only.

## Errors

- Missing markers / empty body / no trailing newline fails.

## Exit Code

- 0 (CLI)

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	assertExitZero(t, resp)
	assertCLINilErr(t, resp)
	assertStdoutTrailingNewline(t, resp.Stdout)
	body := resp.Stdout
	for _, needle := range []string{
		"browser-agent",
		"BROWSER_AGENT_SESSION_ID",
		"eval",
		"43761",
	} {
		if !strings.Contains(body, needle) {
			t.Fatalf("skill --show missing %q; body=%s", needle, truncate(body, 800))
		}
	}
	// Nested complete refactor: skill must document nested session side-commands.
	low := strings.ToLower(body)
	if !strings.Contains(low, "session info") &&
		!strings.Contains(low, "session eval") &&
		!strings.Contains(body, "browser-agent session") {
		t.Fatalf("skill --show must document nested session cmds (session info/eval or browser-agent session); body=%s",
			truncate(body, 800))
	}
}
```
