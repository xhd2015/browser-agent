# Scenario

**Feature**: Chrome-Ext-Browser-Agent shell present on module disk

```
ModuleRoot/Chrome-Ext-Browser-Agent/
  public/manifest.json (or manifest.json)
    name / description references Browser Agent
    host_permissions / content_scripts mention 43761
```

## Preconditions

- ModeExtShell.
- ModuleRoot from root Setup.
- No build step in harness.

## Steps

1. Set Mode = ModeExtShell.

## Context

- Distinct from Chrome-Ext-Capture-API (browser-trace product).

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeExtShell
	if req.ModuleRoot == "" {
		t.Fatal("ModuleRoot must be set by root Setup")
	}
	return nil
}
```
