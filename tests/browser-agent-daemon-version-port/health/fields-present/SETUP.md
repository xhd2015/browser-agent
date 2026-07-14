# Scenario

**Feature**: fields present

```
RunDaemon -> GET /v1/health {product, daemon_version, base_dir}
RunDaemon -> server.json daemon_version
```

## Preconditions

- `HealthOp = HealthOpFieldsPresent`.

## Steps

1. Set `HealthOp = HealthOpFieldsPresent`.

## Context

- See ASSERT.md for expected outcomes.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.HealthOp = HealthOpFieldsPresent
	return nil
}
```
