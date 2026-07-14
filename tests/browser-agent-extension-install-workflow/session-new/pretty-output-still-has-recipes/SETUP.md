# Scenario

**Feature**: enriched stdout preserves nested session command recipes

```
SessionNew -> stdout still lists session info / eval / run markers
```

## Preconditions

- Extension block added without removing Next recipes.

## Steps

1. Set `SessionNewOp = pretty-output-still-has-recipes`.

## Context

- Regression guard for phase-8 pretty output.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.SessionNewOp = SessionNewOpPrettyOutputStillHasRecipes
	return nil
}
```