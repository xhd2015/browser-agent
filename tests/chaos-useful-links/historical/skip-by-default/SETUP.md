# Scenario

**Feature**: default load skips historical/archived/deprecated section links

```
IncludeArchived=false
  -> active.example.com/* kept
  -> old.example.com/* under skipped headings omitted
```

## Preconditions

- IncludeArchived is false (default).

## Steps

1. Set IncludeArchived false.
2. Want active URLs; WantNot old URLs.

## Context

- ArchivedSkipped count should be ≥ 1.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.IncludeArchived = false
	req.WantURLs = []string{
		"active.example.com/top",
		"active.example.com/after",
		"active.example.com/last",
	}
	req.WantNotURLs = []string{
		"old.example.com/hist-a",
		"old.example.com/hist-nested",
		"old.example.com/archived",
		"old.example.com/deprecated",
	}
	req.WantCount = 3
	req.WantCountSet = true
	return nil
}
```
