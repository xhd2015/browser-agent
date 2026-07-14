# Scenario

**Feature**: Dual extension shells on ModuleRoot disk

```
ModuleRoot/Chrome-Ext-Browser-Agent  (agent, 43761, __BROWSER_AGENT_EXT__)
ModuleRoot/Chrome-Ext-Capture-API    (trace, 43759, __BROWSER_TRACE_EXT__)
```

## Preconditions

- ModeShellDisk.
- ModuleRoot from root Setup.
- ShellProduct + ShellProbe set by leaves.
- No build/npm — read-only probes.

## Steps

1. Set Mode = ModeShellDisk.
2. Leave ShellProduct/ShellProbe to leaves.

## Context

- Distinct from staged Bundle fixture under temp Root.
- Covers production shell sources, not only embed mini.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeShellDisk
	if req.ModuleRoot == "" {
		t.Fatal("ModuleRoot must be set by root Setup")
	}
	return nil
}
```
