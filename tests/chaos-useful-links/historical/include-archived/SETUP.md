# Scenario

**Feature**: IncludeArchived keeps links under historical/archived/deprecated

```
IncludeArchived=true
  -> active + old section links all present
```

## Preconditions

- IncludeArchived is true.

## Steps

1. Set IncludeArchived true.
2. Want both active and old URLs.

## Context

- ArchivedSkipped should be 0 when include-archived is on.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.IncludeArchived = true
	req.WantURLs = []string{
		"active.example.com/top",
		"active.example.com/after",
		"active.example.com/last",
		"old.example.com/hist-a",
		"old.example.com/hist-nested",
		"old.example.com/archived",
		"old.example.com/deprecated",
	}
	req.WantCount = 7
	req.WantCountSet = true
	return nil
}
```
