# Scenario

**Feature**: neither --links nor --random-links is an error

```
ResolveSeedSource("", false) -> error (non-zero)
```

## Preconditions

- SourceOp is missing.
- No links path; RandomLinks false.

## Steps

1. Set SourceOp to missing.

## Context

- Error text should mention --links or --random-links (soft check).

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.SourceOp = SourceMissing
	req.RandomLinks = false
	req.LinksPath = ""
	return nil
}
```
