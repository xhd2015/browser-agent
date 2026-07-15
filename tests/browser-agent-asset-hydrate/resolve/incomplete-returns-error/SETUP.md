# Scenario

**Feature**: incomplete FS resolve returns clear error (no pretend success)

```
empty temp DirFS (incomplete)
  -> ResolveSessionPageIndexFS(fs)
  -> err != nil (incomplete / not available / embed incomplete)
  -> must not return success with empty or invented HTML
```

## Preconditions

- Empty FS is forced incomplete.
- ExpectResolveOK false.
- P1 must not download or fall through silently.

## Steps

1. Set `ResolveFixtureName = FixtureEmpty`.
2. Set `ExpectResolveOK = false`.

## Context

- Callers branch on completeness or on this error before later hydrate phases.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ResolveFixtureName = FixtureEmpty
	req.ExpectResolveOK = false
	return nil
}
```
