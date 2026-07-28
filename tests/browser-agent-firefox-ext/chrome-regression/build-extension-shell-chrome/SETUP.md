# Scenario

**Feature**: BuildExtensionShell still targets Chrome-Ext-Browser-Agent only

```
stageChromePublic(ShellRoot) -> BuildExtensionShell -> Chrome-Ext-Browser-Agent/build/
```

## Preconditions

- ChromeRegressionOp = build-extension-shell-chrome.
- ShellRoot temp with Chrome-Ext public fixture (no Firefox package required).

## Steps

1. Set ChromeRegressionOp = ChromeRegressionOpBuildShellChrome.
2. Stage minimal Chrome-Ext-Browser-Agent/public under ShellRoot.

## Context

- Ensures Firefox API work did not repoint default BuildExtensionShell.

```go
import (
	"os"
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.ChromeRegressionOp = ChromeRegressionOpBuildShellChrome
	req.ShellRoot = filepath.Join(t.TempDir(), "repo")
	if err := os.MkdirAll(req.ShellRoot, 0o755); err != nil {
		return err
	}
	stageChromePublic(t, req.ShellRoot)
	return nil
}
```
