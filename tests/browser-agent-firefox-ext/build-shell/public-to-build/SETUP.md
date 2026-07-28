# Scenario

**Feature**: BuildFirefoxExtensionShell stages public files into build/

```
stageFirefoxPublic(ShellRoot) -> BuildFirefoxExtensionShell -> build/manifest.json
```

## Preconditions

- BuildShellOp = public-to-build.
- ShellRoot temp with Firefox-Ext-Browser-Agent/public fixture.

## Steps

1. Set BuildShellOp = BuildShellOpPublicToBuild.
2. Allocate ShellRoot temp and stage minimal Firefox public package.

## Context

- Isolated temp root; does not mutate ModuleRoot package.

```go
import (
	"os"
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.BuildShellOp = BuildShellOpPublicToBuild
	req.ShellRoot = filepath.Join(t.TempDir(), "repo")
	if err := os.MkdirAll(req.ShellRoot, 0o755); err != nil {
		return err
	}
	stageFirefoxPublic(t, req.ShellRoot)
	return nil
}
```
