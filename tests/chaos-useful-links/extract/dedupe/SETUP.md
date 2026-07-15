# Scenario

**Feature**: identical URLs collapse to one seed after normalize

```
testdata/dedupe.md
  same URL ×3 forms + other
  -> 2 seeds (same + other)
```

## Preconditions

- Fixture is `dedupe.md`.
- Counts.Deduped should be 2; raw candidates ≥ 3.

## Steps

1. Set Fixture to dedupe.md.
2. WantURLs for /same and /other; WantCount=2.

## Context

- Dedupe key is normalized URL.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Fixture = "dedupe.md"
	req.WantURLs = []string{
		"example.com/same",
		"example.com/other",
	}
	req.WantCount = 2
	req.WantCountSet = true
	return nil
}
```
