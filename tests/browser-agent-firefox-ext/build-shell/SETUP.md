# Scenario

**Feature**: BuildFirefoxExtensionShell copies public → build

```
Test Client -> BuildFirefoxExtensionShell(ShellRoot)
  -> ShellRoot/Firefox-Ext-Browser-Agent/build/ (abs path)
  -> error when public/manifest.json missing
```

## Preconditions

- Mode = build-shell.
- ShellRoot is a temp dir staged per leaf (public-to-build stages fixture; missing leaves empty tree).

## Steps

1. Set Mode = ModeBuildShell.
2. Leaf sets BuildShellOp and stages ShellRoot.

## Context

- No npm/vite; pure copy shell like Chrome BuildExtensionShell.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Mode = ModeBuildShell
	return nil
}
```
