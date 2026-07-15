# Scenario

**Feature**: --random-links loads built-in public https seeds

```
ResolveSeedSource("", true) / RandomSeeds()
  -> ≥3 seeds; hosts example.com, google.com, baidu.com
```

## Preconditions

- SourceOp is random-links.

## Steps

1. Set SourceOp to random-links.

## Context

- Source.Type should be random-links; Path empty.
- Seed.Source should indicate random-links.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.SourceOp = SourceRandom
	req.RandomLinks = true
	req.LinksPath = ""
	return nil
}
```
