# Scenario

**Feature**: ManagedChromeLayout path resolution

```
DefaultManagedChromeLayout() -> ~/.browser-agent/managed-chrome + data + extensions
LayoutFromRoot(custom)       -> {root}/data + {root}/extensions
```

## Preconditions

- ModeLayout.
- Pure path resolution; no Chrome launch.

## Steps

1. Set Mode = ModeLayout.

## Context

- Default root uses operator home; custom root uses per-leaf temp dir from root Setup.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeLayout
	return nil
}
```
