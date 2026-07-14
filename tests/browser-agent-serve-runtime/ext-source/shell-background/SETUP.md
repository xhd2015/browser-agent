# Scenario

**Feature**: Chrome-Ext-Browser-Agent background has WS agent tokens (E1)

```
Chrome-Ext-Browser-Agent/public/background.js (or src/build)
  contains /v1/ws or ws:// + hello + job + result
```

## Preconditions

- ExtSourceTarget = shell-background.

## Steps

1. Set ExtSourceTarget shell-background.

## Context

- Prefer public/ layout used by product shell.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ExtSourceTarget = ExtSrcShellBackground
	return nil
}
```
