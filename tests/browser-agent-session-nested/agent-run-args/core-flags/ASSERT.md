## Expected

Requirement **A3**:

- Non-empty argv including `run`.
- `--session-id` value is `browser-agent-sess-ctrl-core-1` (agent-run id).
- `--env` carries `BROWSER_AGENT_SESSION_ID=ctrl-core-1` (control id).
- `--no-submit` present.
- Core agent-run tokens: `grok-tty`, `--auto-send-or-resume`, `--new-terminal`, `--open`.

## Side Effects

- None (pure).

## Errors

- Bare control as agent-run session id fails isolation.
- Missing `--env` forces process env overlay (forbidden).

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
	args := resp.AgentRunArgs
	if len(args) == 0 {
		t.Fatal("BuildAgentRunArgs returned empty argv")
	}
	ctrl := req.AgentArgsControlID
	wantAgent := wantAgentRunID(ctrl)

	// run
	if !argvHasToken(args, "run") {
		t.Fatalf("argv missing run; args=%v", args)
	}

	// --session-id = agent-run id (not bare control unless control was already prefixed)
	sid := argvSessionIDValue(args)
	if sid == "" {
		t.Fatalf("argv missing --session-id; args=%v", args)
	}
	if sid != wantAgent {
		t.Fatalf("--session-id=%q, want agent-run id %q; args=%v", sid, wantAgent, args)
	}
	if sid == ctrl && !strings.HasPrefix(ctrl, AgentRunSessionIDPrefix) {
		t.Fatalf("--session-id must be prefixed agent-run id, not bare control %q", ctrl)
	}

	// --env BROWSER_AGENT_SESSION_ID=<control>
	envVal := argvEnvValue(args, "BROWSER_AGENT_SESSION_ID")
	if envVal == "" {
		t.Fatalf("argv missing --env BROWSER_AGENT_SESSION_ID=<control>; args=%v", args)
	}
	if envVal != ctrl {
		t.Fatalf("--env BROWSER_AGENT_SESSION_ID=%q, want control %q; args=%v", envVal, ctrl, args)
	}

	// --no-submit
	if !argvHasToken(args, "--no-submit") {
		found := false
		for _, a := range args {
			if a == "--no-submit" || strings.HasPrefix(a, "--no-submit=") {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("argv missing --no-submit; args=%v", args)
		}
	}

	joined := strings.Join(args, " ")
	for _, needle := range []string{"grok-tty", "auto-send-or-resume", "new-terminal"} {
		if !strings.Contains(joined, needle) {
			t.Fatalf("argv missing %q; args=%v", needle, args)
		}
	}
	if !argvHasToken(args, "--open") {
		t.Fatalf("argv missing --open; args=%v", args)
	}
}
```
