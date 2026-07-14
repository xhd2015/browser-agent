# Scenario

**Feature**: extension initiates stop by POSTing complete (popup Stop)

```
# Server-driven session; extension user clicks Stop (mocked as complete POST)
Mock Extension -> status recording
Mock Extension -> POST /v1/complete {stop_reason: extension, har: multi-tab}
Session -> saved
```

## Preconditions

- `StopMode = extension`.
- Mock does not need to wait for server `stop` command before completing.

## Steps

1. Set stop mode to extension-driven complete.
2. Default multi-tab HAR body from `Run` when `MockHAR` empty.

## Context

- Product popup shows Stop-only during server session; this leaf is the primary happy path.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.StopMode = StopExtension
	req.MockStopReason = "extension"
	return nil
}
```
