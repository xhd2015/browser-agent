# Scenario

**Feature**: /go session page boot JSON is browser=firefox after firefox SessionNew

```
SessionNew(Browser=firefox, NoOpen)
  -> GET /go?session=<id>
  -> boot JSON "browser":"firefox"
```

## Preconditions

- BootInjectOp = go-page-boot-firefox.
- SessionID fixed for readability.

## Steps

1. Set BootInjectOp = BootInjectOpGoPageBootFirefox.
2. Set Browser = "firefox".

## Context

- End-to-end proof that stamp feeds injectSessionBoot without client UA.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.BootInjectOp = BootInjectOpGoPageBootFirefox
	req.Browser = "firefox"
	req.SessionID = "sess-ff-stamp-boot-1"
	return nil
}
```
