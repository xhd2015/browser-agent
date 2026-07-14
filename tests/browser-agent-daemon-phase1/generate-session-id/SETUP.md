# Scenario

**Feature**: GenerateSessionID produces auto session ids for later CLI

```
GenerateSessionID() -> "sess-" + 6 [a-z0-9]
```

## Preconditions

- Mode is generate-session-id.

## Steps

1. Set Mode to generate-session-id.

## Context

- Auto-generate format: `sess-<6 lowercase alnum>`.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeGenerateSessionID
	return nil
}
```