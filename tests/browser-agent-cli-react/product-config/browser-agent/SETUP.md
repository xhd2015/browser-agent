# Scenario

**Feature**: ProductBrowserAgent defaults (C1)

```
ProductBrowserAgent
  id/cliName browser-agent
  controlPort 43761
  features include browser-agent
```

## Preconditions

- ProductID = browser-agent.

## Steps

1. Set ProductID to browser-agent.

## Context

- ExtensionDirName should reference Browser-Agent when present.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ProductID = "browser-agent"
	return nil
}
```
