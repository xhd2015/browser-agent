# Scenario

**Feature**: EnsureDaemon version-aware upgrade Q1/Q2/Q4-Q12

```
EnsureDaemon + CompareVersion -> reuse|warn|kill+respawn
```

## Preconditions

- Mode `ModeEnsureDaemonUpgrade`.
- Leaf sets op-specific field.

## Steps

1. Set `Mode = ModeEnsureDaemonUpgrade`.

## Context

- See root DOCTEST for `Run` dispatch.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeEnsureDaemonUpgrade
	return nil
}
```
