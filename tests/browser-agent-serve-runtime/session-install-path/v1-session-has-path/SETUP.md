# Scenario

**Feature**: session JSON includes extension_install_path after extract-on-serve (F1)

```
GET /v1/session → extension_install_path under BaseDir/extension/...
```

## Preconditions

- Live serve with extract.
- NoOpenChrome + NoAgentRun.

## Steps

1. Reinforce launch-off flags for this leaf.
2. Keep Mode from parent (`session-install-path`).

## Context

- Path should be non-empty; absolute preferred.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	// Leaf reinforces isolation: no real Chrome / agent-run.
	req.NoOpenChrome = true
	req.NoAgentRun = true
	req.InjectOpenChromeFn = false
	req.InjectAgentRunFn = false
	return nil
}
```
