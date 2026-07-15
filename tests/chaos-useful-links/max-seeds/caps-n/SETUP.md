# Scenario

**Feature**: MaxSeeds=N keeps only N seeds after dedupe

```
MaxSeeds=2 + 5 unique fixture URLs -> exactly 2 seeds
```

## Preconditions

- MaxSeeds is 2.

## Steps

1. Set MaxSeeds to 2.
2. WantCount=2.

## Context

- Order of kept seeds is implementation-defined (first N after dedupe order).

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.MaxSeeds = 2
	req.WantCount = 2
	req.WantCountSet = true
	return nil
}
```
