# Scenario

**Feature**: do not re-attempt attach when already attached (requirement #5)

```
URL capturable; recording=true; windowMatch=true; alreadyAttached=true
IsCapturableTabURL -> true
ShouldAttemptAttach -> false
```

## Preconditions

- Tab id is already in `attachedTabs`.
- Other gates remain open; only `AlreadyAttached` is true.

## Steps

1. Set `AlreadyAttached = true`.

## Context

- Prevents double `debugger.attach` / thrashing on repeated `onUpdated` events
  for a tab that is already watched.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.AlreadyAttached = true
	return nil
}
```
