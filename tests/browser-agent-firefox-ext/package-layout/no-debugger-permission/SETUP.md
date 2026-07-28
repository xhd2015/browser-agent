# Scenario

**Feature**: Firefox package must not request Chrome-only debugger permission

```
public/manifest.json permissions -> must NOT include "debugger"
```

## Preconditions

- PackageLayoutOp = no-debugger-permission.

## Steps

1. Set PackageLayoutOp = PackageLayoutOpNoDebugger.

## Context

- tabs/storage/alarms are allowed; debugger is Chrome CDP-only.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.PackageLayoutOp = PackageLayoutOpNoDebugger
	return nil
}
```
