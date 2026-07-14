## Expected

Requirement **C3**:

- Non-nil HandleCLI error (not successful side-command).
- Must **not** take the old flat-handler session-resolve path (dual mention of
  `--session-id` + `BROWSER_AGENT_SESSION_ID` alone as the failure mode).
- Prefer `unknown` command wording, or brief usage that steers operators to
  `session` / `serve` without resolving session for flat `info`.

## Side Effects

- None (no server).

## Errors

- If flat `info` still runs the old session-info resolve handler, fail.

## Exit Code

- Non-zero preferred.

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
	if resp.DispatchTimedOut {
		t.Fatal("HandleCLI timed out on flat info")
	}
	if resp.CLIErr == "" {
		t.Fatal("flat info must not succeed as a side-command handler (want non-nil error)")
	}
	text := strings.ToLower(combinedCLIText(resp))
	// Reject successful-looking session snapshot.
	if strings.Contains(text, `"session_id"`) && !strings.Contains(text, "unknown") {
		t.Fatalf("flat info looks like successful session handler output; text=%s",
			truncate(combinedCLIText(resp), 500))
	}
	// Complete refactor: flat info must not be the dual-source resolve error path.
	// That path is reserved for nested `session info` without a session id.
	dualResolve := strings.Contains(text, "--session-id") &&
		strings.Contains(text, "browser_agent_session_id")
	if dualResolve && !strings.Contains(text, "unknown") {
		t.Fatalf("flat info still uses session-resolve error path; want unknown/usage after nested refactor; text=%s",
			truncate(combinedCLIText(resp), 500))
	}
	okShape := strings.Contains(text, "unknown") ||
		(strings.Contains(text, "usage") && (strings.Contains(text, "session") || strings.Contains(text, "serve"))) ||
		(strings.Contains(text, "session") && strings.Contains(text, "serve") && !dualResolve)
	if !okShape {
		t.Fatalf("flat info error should be unknown/usage-style for nested refactor; got %q",
			truncate(combinedCLIText(resp), 500))
	}
}
```
