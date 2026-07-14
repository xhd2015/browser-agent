# Scenario

**Feature**: session help server port

```
serve --help / session new --help -> new flags
ResolveControlBaseURL -> server.json
```

## Preconditions

- `FlagsOp = FlagsOpSessionHelpServerPort`.

## Steps

1. Set `FlagsOp = FlagsOpSessionHelpServerPort`.

## Context

- See ASSERT.md for expected outcomes.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.FlagsOp = FlagsOpSessionHelpServerPort
	return nil
}
```
