# Scenario

**Feature**: session-page + popup Vite app entries (G2)

```
react/src/apps/session-page/main.tsx (or main.ts / index.tsx)
react/src/apps/popup/main.tsx (or main.ts / index.tsx)
```

## Preconditions

- ReactProbe = apps-entries.

## Steps

1. Set ReactProbeApps.

## Context

- Multi-page Vite config targets these entries.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ReactProbe = ReactProbeApps
	return nil
}
```
