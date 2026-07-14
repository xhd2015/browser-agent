# Scenario

**Feature**: disconnected session info shows install-chrome-extension + load path

```
session info (disconnected) -> Next steps: install-chrome-extension + Load unpacked from path
```

## Preconditions

- Session created via POST; extension not connected.

## Steps

1. Set `SessionInfoOp = disconnected-install-hint`.

## Context

- Must not suggest `open-chrome` (removed).

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.SessionInfoOp = SessionInfoOpDisconnectedInstallHint
	return nil
}
```