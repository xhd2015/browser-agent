# Scenario

**Feature**: do not attempt attach when not recording (requirement #6)

```
URL capturable; recording=false; windowMatch=true; alreadyAttached=false
IsCapturableTabURL -> true
ShouldAttemptAttach -> false
```

## Preconditions

- Session is idle / stopped / not yet started.
- Other gates remain open; only `Recording` is false.

## Steps

1. Set `Recording = false`.

## Context

- Attach is only for an active recording session in the pinned window.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Recording = false
	return nil
}
```
