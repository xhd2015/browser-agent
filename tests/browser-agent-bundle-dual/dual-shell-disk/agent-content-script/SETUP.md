# Scenario

**Feature**: Browser-Agent content script page marker (D3)

```
Chrome-Ext-Browser-Agent contentScript/content.js
  __BROWSER_AGENT_EXT__ + browser-agent
```

## Preconditions

- ShellProduct = browser-agent; ShellProbe = content-script.

## Steps

1. Set ShellProduct/ShellProbe for agent content script.

## Context

- public/contentScript.js is the usual path for the agent shell.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ShellProduct = ShellProductAgent
	req.ShellProbe = ShellProbeContentScript
	return nil
}
```
