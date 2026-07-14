# Scenario

**Feature**: session gates that block attach even when URL is capturable

```
# Fixed capturable URL; flip exactly one gate off
URL = CapturableFixture (https casement)
IsCapturableTabURL(URL) -> true  (always under this branch)
ShouldAttemptAttach(...) -> false when any of:
  !recording | !windowMatch | alreadyAttached
```

## Preconditions

- URL is fixed to `CapturableFixture` so Attempt=false isolates gate logic.
- `WantCapturable = true` for all children.
- `WantAttempt = false` for all children (each leaf blocks one gate).
- Children set exactly one of: `Recording=false`, `WindowMatch=false`,
  `AlreadyAttached=true`.

## Steps

1. Set `URL = CapturableFixture`.
2. Set `WantCapturable = true`.
3. Set `WantAttempt = false`.
4. Leave which gate is closed to the leaf.

## Context

- MECE on which single gate fails: not-recording | wrong-window | already-attached.
- Multi-gate failures are redundant (any false short-circuits) and omitted.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.URL = CapturableFixture
	req.WantCapturable = true
	req.WantAttempt = false
	// Start from open gates; leaf closes exactly one.
	req.Recording = true
	req.WindowMatch = true
	req.AlreadyAttached = false
	return nil
}
```
