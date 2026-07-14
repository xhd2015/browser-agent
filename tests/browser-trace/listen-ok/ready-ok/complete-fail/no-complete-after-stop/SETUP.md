# Scenario

**Feature**: complete never arrives after stop → exit ≠ 0, no corrupt HAR

```
Mock Extension: hello -> start -> status recording -> (observe stop) -> hang
browser-trace: cancel/stop -> wait CompleteTimeout -> fail
meta may record timeout/failure; recording.har must not be a corrupt partial final
```

## Preconditions

- Parent sets `ExtRecordNoComplete`, short complete timeout, CLI stop mode.

## Steps

1. Rely on parent script/timeouts.
2. Optional: set `MockStopReason` unused (no complete body).

## Context

- Requirement scenario #6.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	// Explicit leaf confirmation of parent intent.
	req.ExtensionScript = ExtRecordNoComplete
	req.StopMode = StopCLI
	return nil
}
```
