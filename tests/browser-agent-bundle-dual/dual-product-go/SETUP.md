# Scenario

**Feature**: Dual ProductConfig Go exports (browser-agent vs browser-trace)

```
Test Client -> browseragent.ProductBrowserAgent | ProductBrowserTrace
  -> ControlPort, ID/CLI, Features, PageMarker, ExtensionDirName
```

## Preconditions

- ModeProductGo (pure package constants; no server, no Bundle).
- ProductProbe set by leaf: browser-agent | browser-trace | ports-differ.

## Steps

1. Set Mode = ModeProductGo.
2. Leave ProductProbe to leaf.

## Context

- Control ports must not be swapped (agent 43761, trace 43759).
- Complements sealed cli-react product-config leaves; this tree adds ports-differ.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeProductGo
	return nil
}
```
