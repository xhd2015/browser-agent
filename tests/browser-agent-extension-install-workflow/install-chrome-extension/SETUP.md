# Scenario

**Feature**: install-chrome-extension CLI uses canonical path

```
Operator -> HandleCLI(install-chrome-extension [--base-dir]) -> stdout path + steps
```

## Preconditions

- `TestHome` isolates canonical layout from real user home.
- Path must use `extensions/browser-agent/` regardless of `--base-dir`.

## Steps

1. Set `Mode = install-chrome-extension`.
2. Leaf sets `InstallChromeExtOp`.

## Context

- CLI must exit 0 on success.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeInstallChromeExt
	return nil
}
```