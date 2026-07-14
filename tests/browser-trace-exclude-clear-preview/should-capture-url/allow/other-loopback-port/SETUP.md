# Scenario

**Feature**: allow loopback traffic on a non-control port (req #2 extension)

```
ShouldCaptureURL("http://127.0.0.1:8080/api/local") -> true
```

## Preconditions

- Same IP family as control host but **port ≠ 43759**.
- Local dev servers on 8080 must still be capturable.

## Steps

1. Set `CaptureURL = "http://127.0.0.1:8080/api/local"`.

## Context

- Guards against over-broad “any 127.0.0.1” excludes.
- Symmetric intent for localhost:3000 is covered by the same port rule; one
  leaf is enough.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.CaptureURL = "http://127.0.0.1:8080/api/local"
	return nil
}
```
