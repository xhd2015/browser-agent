# Scenario

**Feature**: non-capturable tab URLs — skip attach (req #1–#3 + skip list)

```
IsCapturableTabURL(non-capturable) -> false
ShouldAttemptAttach(open gates, non-capturable) -> false
```

## Preconditions

- Expected: `WantCapturable = false`, `WantAttempt = false`.
- Gates open so a false Attempt is solely due to URL class.
- Children choose concrete non-capturable form (empty, chrome, extension, …).

## Steps

1. Set `WantCapturable = false`.
2. Set `WantAttempt = false`.

## Context

- Aligns with `attachAllTabsInWindow` skip prefixes: `chrome://`,
  `chrome-extension://`, `devtools://`, plus empty and `about:blank`.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.WantCapturable = false
	req.WantAttempt = false
	return nil
}
```
