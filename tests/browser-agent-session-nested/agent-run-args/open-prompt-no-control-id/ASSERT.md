## Expected

Requirement **A4**:

- Open prompt (tokens after `--`, else last prompt-like token) is non-empty when
  path ends with SYSTEM.md.
- Open prompt does **not** contain the bare control id `ctrl-open-unique-zz9`.
- Control id is still carried via `--env BROWSER_AGENT_SESSION_ID=…` (not the open text).

## Side Effects

- None (pure).

## Errors

- Embedding control id in the open prompt couples agent-run chat to control id.

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
	ctrl := req.AgentArgsControlID
	prompt := resp.OpenPrompt
	if strings.TrimSpace(prompt) == "" {
		// Try recover from full argv
		prompt = extractOpenPrompt(resp.AgentRunArgs)
	}
	if strings.TrimSpace(prompt) == "" {
		t.Fatalf("open prompt empty; args=%v", resp.AgentRunArgs)
	}
	// Absolute SYSTEM.md path required (not bare "at SYSTEM.md").
	if !strings.Contains(prompt, "SYSTEM.md") {
		t.Fatalf("open prompt must mention SYSTEM.md; prompt=%q", prompt)
	}
	if !strings.Contains(prompt, "/") && !strings.Contains(prompt, "\\") {
		t.Fatalf("open prompt SYSTEM.md path should be absolute; prompt=%q", prompt)
	}
	if strings.Contains(prompt, "at SYSTEM.md") {
		t.Fatalf("open prompt must not use bare SYSTEM.md; want abs path; prompt=%q", prompt)
	}
	// SETUP path (or cleaned abs equivalent).
	if want := strings.TrimSpace(req.AgentArgsPromptPath); want != "" {
		if !strings.Contains(prompt, want) && !strings.Contains(prompt, "browser-agent-playbook") {
			t.Fatalf("open prompt should include abs playbook path %q; prompt=%q args=%v",
				want, prompt, resp.AgentRunArgs)
		}
	}
	if strings.Contains(prompt, ctrl) {
		t.Fatalf("open prompt must not contain bare control id %q; prompt=%q args=%v",
			ctrl, prompt, resp.AgentRunArgs)
	}
	// Still require --env carries control (session resolve for nested CLI).
	envVal := argvEnvValue(resp.AgentRunArgs, "BROWSER_AGENT_SESSION_ID")
	if envVal != ctrl {
		t.Fatalf("--env BROWSER_AGENT_SESSION_ID=%q, want control %q; args=%v",
			envVal, ctrl, resp.AgentRunArgs)
	}
}
```
