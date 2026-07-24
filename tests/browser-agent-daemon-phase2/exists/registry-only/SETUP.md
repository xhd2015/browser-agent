# Scenario

**Feature**: Exists true after Create (registry entry)

```
Create(id) -> Exists(id) -> true
```

## Preconditions

- ExistsPreCreate true.

## Steps

1. Set ExistsSessionID `reg-only`; ExistsPreCreate true.

## Context

- Normal live session path.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.ExistsSessionID = "reg-only"
	req.ExistsPreCreate = true
	return nil
}
```