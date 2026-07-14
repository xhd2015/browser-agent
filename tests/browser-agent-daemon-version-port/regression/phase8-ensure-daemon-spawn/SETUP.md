# Scenario

**Feature**: phase8 ensure daemon spawn

```
EnsureDaemon spawn on explicit --port N -> healthy + server.json
```

## Preconditions

- `RegressionOp = RegressionOpPhase8Spawn`.

## Steps

1. Set `RegressionOp = RegressionOpPhase8Spawn`.

## Context

- See ASSERT.md for expected outcomes.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.RegressionOp = RegressionOpPhase8Spawn
	return nil
}
```
