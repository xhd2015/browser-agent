# Scenario

**Feature**: running daemon status includes embedded extension metadata

```
RunDaemon + session -> QueryDaemonStatus -> ExtensionPath under extensions/browser-agent/
```

## Preconditions

- `QueryStatusOp` running-populates-extension.

## Steps

1. Set `QueryStatusOp = QueryStatusOpRunningPopulatesExtension`.

## Context

- Extension bundle from session or canonical extract under isolated HOME.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.QueryStatusOp = QueryStatusOpRunningPopulatesExtension
	return nil
}
```