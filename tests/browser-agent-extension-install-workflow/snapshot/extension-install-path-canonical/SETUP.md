# Scenario

**Feature**: session snapshot exposes canonical extension_install_path

```
SessionNew -> GET /v1/session?session=<id> -> extension_install_path canonical segment
```

## Preconditions

- Fresh session via SessionNew package API.

## Steps

1. Set `SnapshotOp = extension-install-path-canonical`.

## Context

- Field snake_case `extension_install_path`.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.SnapshotOp = SnapshotOpExtensionInstallPathCanonical
	return nil
}
```