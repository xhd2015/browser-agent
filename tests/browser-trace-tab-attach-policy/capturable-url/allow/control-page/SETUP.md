# Scenario

**Feature**: product control page is capturable for attach (requirement #7)

```
URL = "http://127.0.0.1:43759/go"
IsCapturableTabURL(URL) -> true
ShouldAttemptAttach(true, true, false, URL) -> true
# Note: ShouldCaptureURL still excludes control *request* traffic (other tree)
```

## Preconditions

- Control session page on product loopback port **43759**.
- Gates fully open.

## Steps

1. Set `URL = ControlPageFixture` (`http://127.0.0.1:43759/go`).

## Context

- Requirement: attach to control page is OK; multi-tab attach should not
  special-case skip the session tab solely because it is the control host.
- Do **not** confuse with `ShouldCaptureURL` exclude of control-host network
  entries.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.URL = ControlPageFixture
	return nil
}
```
