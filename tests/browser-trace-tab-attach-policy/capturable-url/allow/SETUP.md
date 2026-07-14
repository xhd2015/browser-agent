# Scenario

**Feature**: capturable http(s) tab URLs — attach eligible (req #4, #7)

```
IsCapturableTabURL(http(s) url) -> true
ShouldAttemptAttach(open gates, http(s) url) -> true
```

## Preconditions

- Expected: `WantCapturable = true`, `WantAttempt = true`.
- Gates open (recording, window match, not already attached).
- Children choose concrete capturable URL (app https vs control page).

## Steps

1. Set `WantCapturable = true`.
2. Set `WantAttempt = true`.

## Context

- Happy path after navigation from empty/`chrome://` to a real page.
- Control page is capturable for **attach** even though `ShouldCaptureURL`
  excludes control-host *requests* (different helper / different tree).

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.WantCapturable = true
	req.WantAttempt = true
	return nil
}
```
