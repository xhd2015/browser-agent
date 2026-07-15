# Scenario

**Feature**: strip trailing punctuation from extracted URLs

```
testdata/trailing-punct.txt
  https://example.com/trail).  etc.
  -> clean URLs without ).,;]
```

## Preconditions

- Fixture is `trailing-punct.txt`.
- Normalized URLs must not end with `)`, `.`, `,`, `;`, or `]`.

## Steps

1. Set Fixture to trailing-punct.txt.
2. Set WantURLs for clean path suffixes.

## Context

- Requirement: strip trailing `).,;]` and wrapping punctuation.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Fixture = "trailing-punct.txt"
	req.WantURLs = []string{
		"example.com/trail",
		"example.com/comma",
		"example.com/semi",
		"example.com/paren",
		"example.com/brack",
	}
	req.WantCount = 5
	req.WantCountSet = true
	return nil
}
```
