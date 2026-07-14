# Scenario

**Feature**: not-running status still resolves canonical extension

```
no server.json -> QueryDaemonStatus -> ExtensionPath from EnsureCanonicalExtension
```

## Preconditions

- `QueryStatusOp` not-running-populates-extension.

## Steps

1. Set `QueryStatusOp = QueryStatusOpNotRunningPopulatesExt`.

## Context

- No `server.json`; `Running=false`; extension block still populated.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.QueryStatusOp = QueryStatusOpNotRunningPopulatesExt
	return nil
}
```