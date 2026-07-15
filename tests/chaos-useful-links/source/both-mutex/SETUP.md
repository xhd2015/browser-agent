# Scenario

**Feature**: both --links and --random-links is a mutex error

```
ResolveSeedSource(path, true) -> error (non-zero)
```

## Preconditions

- SourceOp is both-mutex.
- LinksPath points at a real fixture; RandomLinks true.

## Steps

1. Set SourceOp to both-mutex.
2. LinksPath defaults to mixed.md via Run when empty; set explicitly for clarity.

## Context

- Error when both sources set.

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.SourceOp = SourceBoth
	req.RandomLinks = true
	req.LinksPath = filepath.Join(DOCTEST_ROOT, "testdata", "mixed.md")
	return nil
}
```
