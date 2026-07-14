# Scenario

**Feature**: Chrome-Ext-Browser-Agent manifest port + name (D1)

```
Chrome-Ext-Browser-Agent/**/manifest.json
  Browser Agent + 43761
```

## Preconditions

- ShellProduct = browser-agent; ShellProbe = manifest.

## Steps

1. Set ShellProduct/ShellProbe for agent manifest.

## Context

- public/manifest.json is the usual location.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ShellProduct = ShellProductAgent
	req.ShellProbe = ShellProbeManifest
	return nil
}
```
