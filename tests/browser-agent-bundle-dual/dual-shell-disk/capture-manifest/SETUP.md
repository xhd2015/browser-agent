# Scenario

**Feature**: Chrome-Ext-Capture-API manifest port (D2)

```
Chrome-Ext-Capture-API/**/manifest.json
  43759 (+ API Capture / browser-trace naming preferred)
```

## Preconditions

- ShellProduct = browser-trace (Capture-API); ShellProbe = manifest.

## Steps

1. Set ShellProduct/ShellProbe for capture manifest.

## Context

- public/manifest.json or build/manifest.json accepted.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ShellProduct = ShellProductCapture
	req.ShellProbe = ShellProbeManifest
	return nil
}
```
