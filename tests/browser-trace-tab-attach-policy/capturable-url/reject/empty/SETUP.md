# Scenario

**Feature**: empty tab URL at create time (requirement #1)

```
# tabs.onCreated often has tab.url == ""
IsCapturableTabURL("") -> false
ShouldAttemptAttach(true, true, false, "") -> false
# later navigation re-evaluates policy (not permanent give-up)
```

## Preconditions

- URL is the empty string (not whitespace-only; whitespace may also reject but
  is not separately asserted here).
- Gates open.

## Steps

1. Set `URL = ""`.

## Context

- Core bug: attach only on create with empty URL fails silently; policy must
  return false now but allow true later when URL becomes capturable.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.URL = ""
	return nil
}
```
