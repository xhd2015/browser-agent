# Scenario

**Feature**: serve help host port

```
serve --help / session new --help -> new flags
ResolveControlBaseURL -> server.json
```

## Preconditions

- `FlagsOp = FlagsOpServeHelpHostPort`.

## Steps

1. Set `FlagsOp = FlagsOpServeHelpHostPort`.

## Context

- See ASSERT.md for expected outcomes.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.FlagsOp = FlagsOpServeHelpHostPort
	return nil
}
```
