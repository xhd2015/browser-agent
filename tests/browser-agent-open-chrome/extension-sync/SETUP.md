# Scenario

**Feature**: EnsureManagedExtension syncs embedded MV3 under managed layout

```
EnsureManagedExtension(layout)
  -> {ExtensionsDir}/{version}/manifest.json
```

## Preconditions

- ModeExtensionSync.
- Isolated ManagedRoot temp dir per leaf.

## Steps

1. Set Mode = ModeExtensionSync.

## Context

- Uses embedded extension fixture from browseragent package.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeExtensionSync
	return nil
}
```
