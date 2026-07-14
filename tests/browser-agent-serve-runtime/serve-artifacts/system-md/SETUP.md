# Scenario

**Feature**: serve writes SYSTEM.md with nested session recipes, no control id (A1)

```
serve NoOpenChrome+NoAgentRun
sessionDir/SYSTEM.md contains nested browser-agent session info/eval/run/logs/screenshot
  + BROWSER_AGENT_SESSION_ID
  (does not embed concrete control session id)
```

## Preconditions

- Mode serve-artifacts; probe system-md.
- No Chrome / agent-run.

## Steps

1. Set ServeArtifactProbe system-md.
2. Skip real launch hooks.

## Context

- Playbook must match FormatSystemPrompt nested recipes used by agents.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ServeArtifactProbe = ServeProbeSystemMD
	req.NoOpenChrome = true
	req.NoAgentRun = true
	return nil
}
```
