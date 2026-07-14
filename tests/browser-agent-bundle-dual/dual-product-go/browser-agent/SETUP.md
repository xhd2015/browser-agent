# Scenario

**Feature**: ProductBrowserAgent defaults (C1)

```
ProductBrowserAgent
  id/cli browser-agent
  controlPort 43761
  features include browser-agent
```

## Preconditions

- ProductProbe = browser-agent.

## Steps

1. Set ProductProbe to browser-agent.

## Context

- ExtensionDirName should reference Browser-Agent when present.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ProductProbe = ProductProbeAgent
	return nil
}
```
