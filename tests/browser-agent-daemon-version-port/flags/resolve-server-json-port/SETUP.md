# Scenario

**Feature**: resolve server json port

```
serve --help / session new --help -> new flags
ResolveControlBaseURL -> server.json
```

## Preconditions

- `FlagsOp = FlagsOpResolveServerJSONPort`.

## Steps

1. Set `FlagsOp = FlagsOpResolveServerJSONPort`.

## Context

- See ASSERT.md for expected outcomes.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.FlagsOp = FlagsOpResolveServerJSONPort
	return nil
}
```
