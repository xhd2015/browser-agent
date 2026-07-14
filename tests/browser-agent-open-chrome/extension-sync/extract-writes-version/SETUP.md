# Scenario

**Feature**: first extract writes versioned extension dir

```
EnsureManagedExtension -> extensions/{ver}/manifest.json on disk
```

## Preconditions

- ExtensionSyncOp extract-writes-version.

## Steps

1. Set ExtensionSyncOp = ExtensionSyncOpExtractWritesVersion.

## Context

- Cold extract under ExtensionsDir.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ExtensionSyncOp = ExtensionSyncOpExtractWritesVersion
	return nil
}
```
