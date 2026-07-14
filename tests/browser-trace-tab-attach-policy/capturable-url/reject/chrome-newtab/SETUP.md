# Scenario

**Feature**: `chrome://newtab/` is not capturable (requirement #2)

```
IsCapturableTabURL("chrome://newtab/") -> false
ShouldAttemptAttach(open gates, "chrome://newtab/") -> false
```

## Preconditions

- Typical new-tab URL before user navigates to an http(s) site.
- Gates open.

## Steps

1. Set `URL = "chrome://newtab/"`.

## Context

- Chrome internal pages reject or fail debugger attach; skip list must cover
  the `chrome://` prefix (not only this exact path).

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.URL = "chrome://newtab/"
	return nil
}
```
