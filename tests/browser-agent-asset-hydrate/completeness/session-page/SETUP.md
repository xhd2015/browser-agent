# Scenario

**Feature**: session-page embed completeness rules

```
kind=session-page
  complete: non-empty index|session-page.html + assets JS (or resolved script src)
  incomplete: missing JS / empty / missing index
```

## Preconditions

- Parent mode is completeness.
- AssetKind is session-page.

## Steps

1. Set `AssetKind = KindSessionPage` (`session-page`).

## Context

- Does not cover extension kind (sibling branch).

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.AssetKind = KindSessionPage
	return nil
}
```
