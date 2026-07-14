# Scenario

**Feature**: final HAR arrives via POST /v1/complete within complete-timeout

```
Mock Extension -> recording -> (stop) -> POST /v1/complete {har, stop_reason}
Control Server -> write meta.json + recording.har under session dir
browser-trace -> exit 0, print session path with trailing newline
```

## Preconditions

- Mock uses `record-and-complete` (or CLI stop then complete).
- Complete timeout is sufficient for local mock.

## Steps

1. Ensure complete timeout is at least a few seconds.
2. Descendants split on stop initiator: extension vs CLI signal.

## Context

- Requirement scenarios #4, #5, #7, #8.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ExtensionScript = ExtRecordAndComplete
	req.CompleteTimeout = 5 * time.Second
	return nil
}
```
