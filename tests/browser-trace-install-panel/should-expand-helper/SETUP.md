# Scenario

**Feature**: pure `ShouldExpandInstallPanel(connected, supports)` expand policy

```
# No HTTP; pure package helper for server + client parity
Test Client -> browsertrace.ShouldExpandInstallPanel(connected, supports)
  -> !(connected && supports)
```

## Preconditions

- Mode is pure helper (`ModeShouldExpand`).
- Package must export `ShouldExpandInstallPanel`.
- No control server required for these leaves.

## Steps

1. Set `Mode = ModeShouldExpand` (`"should-expand"`).
2. Descendants set `Connected`, `Supports`, and `WantExpand`.

## Context

- Exhaustive 2×2 truth table via expand/* (true) vs collapse/both (false).
- Leaves fail to compile/run until the helper is exported (TDD red → green).

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeShouldExpand
	return nil
}
```
