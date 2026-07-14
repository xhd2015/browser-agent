# Scenario

**Feature**: known six true; unknown false (F1)

```
IsKnownJobType("info"|"eval"|"run"|"logs"|"screenshot"|"cdp") -> true
IsKnownJobType(""|"foo"|"Eval"|"navigate") -> false
```

## Preconditions

- Mode already go-job-types from parent.

## Steps

1. Ensure Mode stays ModeGoJobTypes (single leaf; Run uses fixed probe lists).

## Context

- Requirement F1.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeGoJobTypes
	return nil
}
```
