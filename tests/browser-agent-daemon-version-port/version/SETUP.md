# Scenario

**Feature**: VERSION.txt embed + loose semver compare

```
ClientVersion() <- go:embed VERSION.txt
CompareVersion(a,b) -> -1|0|+1
```

## Preconditions

- Mode `ModeVersion`.
- Leaf sets op-specific field.

## Steps

1. Set `Mode = ModeVersion`.

## Context

- See root DOCTEST for `Run` dispatch.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeVersion
	return nil
}
```
