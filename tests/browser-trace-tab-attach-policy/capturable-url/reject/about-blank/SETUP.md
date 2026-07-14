# Scenario

**Feature**: `about:blank` is not capturable until real navigation (product prefer)

```
IsCapturableTabURL("about:blank") -> false
ShouldAttemptAttach(open gates, "about:blank") -> false
```

## Preconditions

- Intermediate blank documents should not trigger attach; wait for http(s).
- Gates open.

## Steps

1. Set `URL = "about:blank"`.

## Context

- Requirement notes product preference: `about:blank` → false for attach until
  a real navigation provides a capturable URL.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.URL = "about:blank"
	return nil
}
```
