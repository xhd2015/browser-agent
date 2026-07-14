# Scenario

**Feature**: URL capturability class for debugger attach (`IsCapturableTabURL`)

```
# Gates left open (root defaults) so Attempt tracks Capturable only
Test Client -> IsCapturableTabURL(url) -> reject schemes / empty | allow http(s)
Test Client -> ShouldAttemptAttach(true, true, false, url)
          -> same boolean as IsCapturableTabURL under open gates
```

## Preconditions

- Session gates remain open (recording, window match, not already attached).
- Children split MECE: **reject** (not capturable) vs **allow** (capturable).
- Both Capturable and Attempt are asserted (Attempt mirrors Capturable when gates open).

## Steps

1. Keep root gate defaults.
2. Descendants set `URL`, `WantCapturable`, and `WantAttempt`.

## Context

- Primary product bug surface: create-time empty/`chrome://` URLs must reject
  attach without preventing a later attempt after navigation (policy re-evaluated
  on each event — not tested as state machine here).

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	// Grouping marker: leave URL and wants to reject/allow children.
	// Reaffirm gates open so this branch isolates URL class.
	req.Recording = true
	req.WindowMatch = true
	req.AlreadyAttached = false
	return nil
}
```
