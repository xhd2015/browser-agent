# Scenario

**Feature**: install-chrome-extension stdout shows canonical path and Load unpacked steps

```
HandleCLI(install-chrome-extension) -> stdout with extensions/browser-agent/ + chrome://extensions
```

## Preconditions

- No `--base-dir` flag.

## Steps

1. Set `InstallChromeExtOp = stdout-path-and-steps`.

## Context

- Trailing newline on stdout.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.InstallChromeExtOp = InstallChromeExtOpStdoutPathAndSteps
	return nil
}
```