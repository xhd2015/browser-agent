# Scenario

**Feature**: ProductConfig dual export (browser-agent vs browser-trace)

```
Test Client -> browseragent.ProductBrowserAgent | ProductBrowserTrace
  -> id, cliName, controlPort, features, pageMarker, extensionDirName
```

## Preconditions

- Mode is product-config (pure package constants/struct; no server).
- ProductID set by leaf: browser-agent | browser-trace.

## Steps

1. Set ModeProductConfig.
2. Leave ProductID to leaf.

## Context

- Dual export preferred so React/Go share one design vocabulary.
- Control ports must not be swapped (agent 43761, trace 43759).

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeProductConfig
	return nil
}
```
