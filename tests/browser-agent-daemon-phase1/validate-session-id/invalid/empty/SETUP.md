# Scenario

**Feature**: empty session id is rejected

```
"id=" -> ValidateSessionID -> error
```

## Preconditions

- SessionID is empty string.

## Steps

1. Set SessionID to `""`.

## Context

- Empty id must not be accepted.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.SessionID = ""
	return nil
}
```