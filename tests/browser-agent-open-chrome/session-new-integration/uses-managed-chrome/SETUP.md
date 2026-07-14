# Scenario

**Feature**: SessionNew opens system Chrome (no managed profile argv)

```
SessionNew(explicit id) -> LaunchFn called -> --new-window + session URL only
```

## Preconditions

- SessionNewIntegrationOp uses-managed-chrome.
- Explicit SessionID from root Setup.

## Steps

1. Set SessionNewIntegrationOp = SessionNewIntegrationOpUsesManagedChrome.

## Context

- Verifies session new uses system Chrome (no --user-data-dir / --load-extension).

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.SessionNewIntegrationOp = SessionNewIntegrationOpUsesManagedChrome
	return nil
}
```
