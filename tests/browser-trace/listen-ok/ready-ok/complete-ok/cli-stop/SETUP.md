# Scenario

**Feature**: CLI initiates stop (SIGINT / context cancel) while mock is connected

```
# User hits Ctrl-C on the CLI after recording started
Mock Extension -> recording
User/CLI -> cancel context / SIGINT
Control Server -> queue command stop
Mock Extension <- stop (long-poll)
Mock Extension -> POST /v1/complete {stop_reason: cli}
Session -> saved, exit 0
```

## Preconditions

- `StopMode = cli-signal`.
- Mock waits for `stop` command before POSTing complete.

## Steps

1. Set CLI stop mode and `MockStopReason = cli`.
2. `Run` cancels context after observing `StatusRecording`.

## Context

- Requirement scenario #5.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.StopMode = StopCLI
	req.MockStopReason = "cli"
	return nil
}
```
