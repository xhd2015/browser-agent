# Scenario

**Feature**: CLI signal while mock connected → stop command → complete → files written

```
browser-trace recording
  -- cancel/signal --
Mock receives stop via GET /v1/commands
Mock POSTs /v1/complete
files: meta.json + recording.har; exit 0
```

## Preconditions

- Parent `complete-ok` + `cli-stop` configuration.
- Mock script `record-and-complete` with CLI stop mode.

## Steps

1. Confirm `ExtRecordAndComplete` + `StopCLI`.
2. Run drives cancel after recording; mock completes.

## Context

- Requirement scenario #5.
- Assert focuses on stop delivery + successful save (multi-tab details covered under extension-stop leaf).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ExtensionScript = ExtRecordAndComplete
	req.StopMode = StopCLI
	req.MockStopReason = "cli"
	return nil
}
```
