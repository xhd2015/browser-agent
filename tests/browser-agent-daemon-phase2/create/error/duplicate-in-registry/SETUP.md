# Scenario

**Feature**: second Create with same id returns ErrSessionExists

```
Create(id) ok -> Create(id) -> ErrSessionExists
```

## Preconditions

- SessionID `dup-reg`.

## Steps

1. Set CreateCase duplicate-in-registry.

## Context

- First Create must succeed inside Run before second attempt.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.CreateCase = CreateCaseDuplicateInReg
	req.SessionID = "dup-reg"
	return nil
}
```