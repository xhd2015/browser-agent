# Scenario

**Feature**: Browser=firefox opens via OpenFirefoxFn with session URL

```
SessionNew(Browser=firefox) -> OpenFirefoxFn(sessionURL) once
  # OpenChromeFn never called
  # stdout has browser-agent-firefox + about:debugging
```

## Preconditions

- SessionNewFirefoxOp = open-records-url.
- NoOpenChrome = false (default open).

## Steps

1. Set SessionNewFirefoxOp = SessionNewFirefoxOpOpenRecordsURL.
2. Ensure NoOpenChrome remains false.

## Context

- Primary happy path for firefox session new open.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.SessionNewFirefoxOp = SessionNewFirefoxOpOpenRecordsURL
	req.NoOpenChrome = false
	return nil
}
```
