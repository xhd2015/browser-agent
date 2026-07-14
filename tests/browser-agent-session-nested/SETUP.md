# Scenario

**Feature**: nested session CLI + agent-run prefix/env + SYSTEM.md without control id

```
# pure id
Test Client -> AgentRunSessionID(control) -> browser-agent-sess-<control>

# pure argv
Test Client -> BuildAgentRunArgs(control, SYSTEM.md, workspace)
  -> --session-id=<agent-run-id> --env BROWSER_AGENT_SESSION_ID=<control> --no-submit

# playbook
Test Client -> FormatSystemPrompt(control)
  -> nested recipes; no concrete control id; mentions BROWSER_AGENT_SESSION_ID

# CLI
Operator -> HandleCLI([session, …] | --help | flat info)
  -> nested dispatch only; flat side cmds fail
```

## Preconditions

- Package `github.com/xhd2015/browser-agent/browseragent` is importable.
- RED until implementer exports `AgentRunSessionID` and lands nested CLI + args.
- No real Chrome / agent-run binaries required.
- Ambient `BROWSER_AGENT_SESSION_ID` process env is ignored when CLIEnv map is set.

## Steps

1. Leave Mode empty at root (grouping/leaf Setup sets Mode).
2. Default CLIEnv to empty map in CLI grouping Setup.
3. Helpers below are shared by all leaves.

## Context

- `AgentRunSessionIDPrefix` contract literal: `browser-agent-sess-`.
- Control id remains serve/disk/jobs id; agent-run id is prefixed for agent-run only.
- Session resolve error text must mention both `--session-id` and `BROWSER_AGENT_SESSION_ID`.

```go
import (
	"strings"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	// Root: no Mode; grouping Setup sets it.
	return nil
}

func assertNoRunErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("Run transport error: %v", err)
	}
}

func assertExitZero(t *testing.T, resp *Response) {
	t.Helper()
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if resp.ExitCode != 0 {
		t.Fatalf("ExitCode=%d, want 0; CLIErr=%q stdout=%s stderr=%s",
			resp.ExitCode, resp.CLIErr, truncate(resp.Stdout, 300), truncate(resp.Stderr, 300))
	}
}

func combinedCLIText(resp *Response) string {
	if resp == nil {
		return ""
	}
	return resp.Stdout + resp.Stderr + resp.CLIErr + resp.ErrText
}

func assertPrintedTrailingNewline(t *testing.T, resp *Response) {
	t.Helper()
	text := resp.Stdout
	if text == "" {
		text = resp.Stderr
	}
	if text == "" {
		t.Fatal("expected printed help/usage body")
	}
	if !strings.HasSuffix(text, "\n") {
		t.Fatalf("printed body must end with \\n; got tail %q", tail(text, 40))
	}
}

func assertSessionResolveErrorText(t *testing.T, text string) {
	t.Helper()
	if !strings.Contains(text, "--session-id") {
		t.Fatalf("error/text must mention --session-id; got %q", text)
	}
	if !strings.Contains(text, "BROWSER_AGENT_SESSION_ID") {
		t.Fatalf("error/text must mention BROWSER_AGENT_SESSION_ID; got %q", text)
	}
}

func argvHasToken(args []string, tok string) bool {
	for _, a := range args {
		if a == tok {
			return true
		}
	}
	return false
}

func argvSessionIDValue(args []string) string {
	for i, a := range args {
		if strings.HasPrefix(a, "--session-id=") {
			return strings.TrimPrefix(a, "--session-id=")
		}
		if a == "--session-id" && i+1 < len(args) {
			return args[i+1]
		}
	}
	return ""
}

// argvEnvValue returns value for --env K=V or --env K V for key K.
func argvEnvValue(args []string, key string) string {
	prefix := key + "="
	for i, a := range args {
		if a == "--env" && i+1 < len(args) {
			v := args[i+1]
			if strings.HasPrefix(v, prefix) {
				return strings.TrimPrefix(v, prefix)
			}
			// --env K V
			if v == key && i+2 < len(args) {
				return args[i+2]
			}
		}
		if strings.HasPrefix(a, "--env=") {
			v := strings.TrimPrefix(a, "--env=")
			if strings.HasPrefix(v, prefix) {
				return strings.TrimPrefix(v, prefix)
			}
		}
	}
	return ""
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

func tail(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[len(s)-n:]
}

func wantAgentRunID(control string) string {
	if strings.HasPrefix(control, AgentRunSessionIDPrefix) {
		return control
	}
	return AgentRunSessionIDPrefix + control
}
```
