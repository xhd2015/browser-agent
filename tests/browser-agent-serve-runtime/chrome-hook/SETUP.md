# Scenario

**Feature**: OpenChromeFn honors NoOpenChrome (injectable; never real Chrome)

```
# skipped path
NoOpenChrome=true + OpenChromeFn set -> fn never called

# called path
NoOpenChrome=false + OpenChromeFn records -> once(sessionURL, extensionInstallPath)
```

## Preconditions

- Mode is `chrome-hook`.
- Agent-run isolated off (`NoAgentRun=true`).
- Harness always injects recording OpenChromeFn.

## Steps

1. Set `Mode = ModeChromeHook`.
2. Children set `HookExpect` (skipped | called).

## Context

- Requirement B1–B2.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeChromeHook
	req.NoAgentRun = true
	req.InjectOpenChromeFn = true
	return nil
}
```
