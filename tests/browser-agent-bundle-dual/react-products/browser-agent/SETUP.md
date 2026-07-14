# Scenario

**Feature**: React browser-agent product port 43761 (E1)

```
react/src/products/browser-agent.ts
  contains 43761 (+ browser-agent id)
```

## Preconditions

- ReactProductID = browser-agent.

## Steps

1. Set ReactProductID to browser-agent.

## Context

- Mirrors Go ProductBrowserAgent.ControlPort.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ReactProductID = ReactProductAgent
	return nil
}
```
