# Scenario

**Feature**: extension embed completeness rules

```
kind=extension
  complete: non-empty manifest.json + non-empty background.js
```

## Preconditions

- Parent mode is completeness.
- AssetKind is extension.

## Steps

1. Set `AssetKind = KindExtension` (`extension`).

## Context

- P1 default check uses `background.js` (not only service_worker path).

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.AssetKind = KindExtension
	return nil
}
```
