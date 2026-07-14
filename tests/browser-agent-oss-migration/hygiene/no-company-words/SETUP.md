# Scenario

**Feature**: no company words remain in tracked tree

```
# search banned company tokens
Test Client -> rg company-word pattern from repo root
Test Client <- zero hits outside .git/node_modules/dist
```

## Preconditions

- Words to ban: originals from REQUIREMENT-DESIGN masking table (case-insensitive).

## Steps

1. Set `Leaf = no-company-words`.

## Context

- Masked replacements (`some-x`, etc.) are allowed.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Leaf = LeafNoCompanyWords
	return nil
}
```