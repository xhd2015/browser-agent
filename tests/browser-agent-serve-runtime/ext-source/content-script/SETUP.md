# Scenario

**Feature**: contentScript sets __BROWSER_AGENT_EXT__ with browser-agent feature (E2)

```
contentScript.js -> window.__BROWSER_AGENT_EXT__ + features include browser-agent
```

## Preconditions

- ExtSourceTarget = content-script.

## Steps

1. Set ExtSourceTarget content-script.

## Context

- Marker used by session page / product detection.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ExtSourceTarget = ExtSrcContentScript
	return nil
}
```
