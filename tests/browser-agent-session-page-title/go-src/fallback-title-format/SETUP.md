# Scenario

**Feature**: `writeFallbackSessionHTML` title format `{id} - Browser Agent` (T2)

```
browseragent/server.go writeFallbackSessionHTML
  <title>…{sessionId} - Browser Agent…</title>
  not sole static "browser-agent session"
```

## Preconditions

- Mode already go-src from parent.

## Steps

1. Set `GoSrcProbe = GoSrcFallbackTitle` (`fallback-title-format`).

## Context

- Requirement scenario 2 (fallback path).
- Implementer may use `fmt.Sprintf` / `%s - Browser Agent` / `sessionID + " - Browser Agent"`.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.GoSrcProbe = GoSrcFallbackTitle
	return nil
}
```
