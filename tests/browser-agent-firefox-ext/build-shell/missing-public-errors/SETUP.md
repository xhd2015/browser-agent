# Scenario

**Feature**: BuildFirefoxExtensionShell errors when public/manifest.json is missing

```
empty ShellRoot (no Firefox-Ext public) -> BuildFirefoxExtensionShell -> error
```

## Preconditions

- BuildShellOp = missing-public-errors.
- ShellRoot is empty temp (no Firefox-Ext-Browser-Agent/public).

## Steps

1. Set BuildShellOp = BuildShellOpMissingPublic.
2. Allocate empty ShellRoot temp dir.

## Context

- Run captures error in BuildShellErr and returns nil transport err so Assert checks the product error.

```go
import (
	"os"
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.BuildShellOp = BuildShellOpMissingPublic
	req.ShellRoot = filepath.Join(t.TempDir(), "empty-repo")
	return os.MkdirAll(req.ShellRoot, 0o755)
}
```
