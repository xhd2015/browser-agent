# Scenario

**Feature**: default managed chrome root under home

```
home -> ~/.browser-agent/managed-chrome
  data/
  extensions/
```

## Preconditions

- LayoutOp default-root.

## Steps

1. Set LayoutOp = LayoutOpDefaultRoot.

## Context

- Asserts suffix under os.UserHomeDir(), not a temp override.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.LayoutOp = LayoutOpDefaultRoot
	return nil
}
```
