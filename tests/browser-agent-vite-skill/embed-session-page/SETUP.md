# Scenario

**Feature**: read embedded session-page production fixture (A)

```
# no server — pure embed FS
Test Client -> browseragent.SessionPageFS()
  -> index.html | session-page.html under embedded/session-page/
```

## Preconditions

- Mode = ModeEmbedFS.
- Committed fixture exists after implement (TDD red until embed lands).

## Steps

1. Set Mode = ModeEmbedFS.
2. Leaf asserts HTML content of embed entry file.

## Context

- Does not start control server.
- Does not require npm or network.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeEmbedFS
	return nil
}
```
