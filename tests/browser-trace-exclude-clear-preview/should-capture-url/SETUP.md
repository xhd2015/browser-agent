# Scenario

**Feature**: pure `ShouldCaptureURL(url)` — exclude product control hosts

```
# No HTTP; pure package helper used by extension (and optionally server docs)
Test Client -> browsertrace.ShouldCaptureURL(url)
  -> false when url is under http://127.0.0.1:43759 or http://localhost:43759
  -> true otherwise (normal app traffic)
```

## Preconditions

- Mode is pure helper (`ModeShouldCapture`).
- Package must export `ShouldCaptureURL`.
- No control server required for these leaves.
- Product control port is **43759** (hard-coded exclude targets).

## Steps

1. Set `Mode = ModeShouldCapture` (`"should-capture"`).
2. Descendants set `CaptureURL` and `WantCapture`.

## Context

- MECE split below: reject (control) vs allow (non-control).
- Leaves fail to compile/run until the helper is exported (TDD red → green).

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeShouldCapture
	return nil
}
```
