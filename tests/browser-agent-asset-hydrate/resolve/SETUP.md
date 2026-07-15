# Scenario

**Feature**: ResolveSessionPageIndexFS embed-only resolve seam

```
# complete
Test Client -> complete session-page DirFS
  -> ResolveSessionPageIndexFS(fs)
  -> html non-empty, source="embed", err=nil

# incomplete
Test Client -> empty / incomplete DirFS
  -> ResolveSessionPageIndexFS(fs)
  -> clear incomplete/not-available error (no download)
```

## Preconditions

- Mode is resolve.
- Leaf sets ResolveFixtureName and ExpectResolveOK.
- P1: no cache, no network — incomplete is a hard error.

## Steps

1. Set `Mode = ModeResolve`.

## Context

- source is always `"embed"` on success in P1.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeResolve
	return nil
}
```
