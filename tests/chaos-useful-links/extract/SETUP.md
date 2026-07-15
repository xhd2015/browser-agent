# Scenario

**Feature**: extract and normalize URLs from markdown/plain text into Seeds

```
Link File / text
  -> Extractor (bare, markdown, backtick, angle; strip trailers; dedupe)
  -> Resolved.Seeds
```

## Preconditions

- Mode is extract.
- Leaf sets Fixture (under testdata/) and WantURLs / WantCount as needed.
- Default Options: IncludeArchived=false, MaxSeeds=0.

## Steps

1. Set Mode to extract.

## Context

- Uses LoadSeedsFromFile when Fixture is set.
- Covers mixed forms, trailing punctuation, and dedupe leaves.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeExtract
	req.IncludeArchived = false
	req.MaxSeeds = 0
	return nil
}
```
