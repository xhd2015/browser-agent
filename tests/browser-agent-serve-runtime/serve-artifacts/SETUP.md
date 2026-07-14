# Scenario

**Feature**: serve writes SYSTEM.md + meta.json under session dir

```
# Operator serve with launch disabled
Test Client -> browseragent.Run(NoOpenChrome, NoAgentRun) -> Control Server
Serve Runtime -> {BaseDir}/sessions/{id}/SYSTEM.md
Serve Runtime -> {BaseDir}/sessions/{id}/meta.json
```

## Preconditions

- Mode is `serve-artifacts`.
- Launch hooks off (no Chrome / agent-run).

## Steps

1. Set `Mode = ModeServeArtifacts`.
2. Force `NoOpenChrome` and `NoAgentRun` (also enforced in Run).
3. Children set `ServeArtifactProbe`.

## Context

- Requirement A1–A2.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeServeArtifacts
	req.NoOpenChrome = true
	req.NoAgentRun = true
	return nil
}
```
