# Scenario

**Feature**: resolve seed source from --links vs --random-links (mutex)

```
Source Resolver
  random-links only -> public catalog
  neither          -> error
  both             -> error
```

## Preconditions

- Mode is source.
- Leaf sets SourceOp.

## Steps

1. Set Mode to source.

## Context

- Library API: ResolveSeedSource(linksPath, randomLinks, opts).
- Random catalog must include example.com, google.com, baidu.com hosts.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeSource
	req.MaxSeeds = 0
	return nil
}
```
