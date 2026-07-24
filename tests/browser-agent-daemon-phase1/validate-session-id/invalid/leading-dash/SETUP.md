# Scenario

**Feature**: session id starting with dash is rejected

```
"id=-bad" -> ValidateSessionID -> error
```

## Preconditions

- SessionID is `-bad`.

## Steps

1. Set SessionID to `-bad`.

## Context

- First character must be alphanumeric.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.SessionID = "-bad"
	return nil
}
```