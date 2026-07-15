# Scenario

**Feature**: extract bare, markdown, backtick, and angle-bracket URLs as seeds

```
testdata/mixed.md
  bare + [t](url) + `url` + <url>
  -> LoadSeedsFromFile -> ≥4 seeds with those URLs
```

## Preconditions

- Fixture is `mixed.md`.
- Expect distinct seeds for bare, md-link, backtick, and angle forms.

## Steps

1. Set Fixture to mixed.md.
2. Set WantURLs for each form path segment.
3. Set WantCount=4.

## Context

- URLs: example.com/bare, /md-link, /backtick, /angle.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Fixture = "mixed.md"
	req.WantURLs = []string{
		"example.com/bare",
		"example.com/md-link",
		"example.com/backtick",
		"example.com/angle",
	}
	req.WantCount = 4
	req.WantCountSet = true
	return nil
}
```
