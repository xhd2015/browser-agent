# Scenario

**Feature**: MaxSeeds=0 keeps all seeds after dedupe

```
MaxSeeds=0 + 5 unique fixture URLs -> 5 seeds
```

## Preconditions

- MaxSeeds is 0.

## Steps

1. Set MaxSeeds to 0.
2. WantCount=5.

## Context

- 0 means unlimited (all).

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.MaxSeeds = 0
	req.WantCount = 5
	req.WantCountSet = true
	req.WantURLs = []string{
		"seed.example.com/a",
		"seed.example.com/b",
		"seed.example.com/c",
		"seed.example.com/d",
		"seed.example.com/e",
	}
	return nil
}
```
