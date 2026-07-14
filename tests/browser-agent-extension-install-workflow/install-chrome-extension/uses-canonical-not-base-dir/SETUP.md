# Scenario

**Feature**: --base-dir does not change canonical install path

```
HandleCLI(install-chrome-extension --base-dir <temp>) -> path still under TestHome canonical layout
```

## Preconditions

- `BaseDir` is a distinct temp sessions dir (not canonical root).

## Steps

1. Set `InstallChromeExtOp = uses-canonical-not-base-dir`.

## Context

- Daemon `--base-dir` and extension install layout are independent.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.InstallChromeExtOp = InstallChromeExtOpUsesCanonicalNotBaseDir
	return nil
}
```