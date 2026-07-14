# Scenario

**Feature**: install CLI / package API prints extension path and load steps

```
# User (or test) invokes install-only path — no full capture session
User -> browser-trace --install-chrome-extension
# Product calls package InstallChromeExtension(stdout, BaseDir)
Installer -> Extractor -> path
Installer -> stdout: path, Developer mode, Load unpacked, chrome://extensions, trailing \n
```

## Preconditions

- Mode is install-cli.
- Package `InstallChromeExtension` is the contract behind the CLI flag.
- BaseDir is writable temp dir.

## Steps

1. Set `Mode = ModeInstallCLI` (`"install-cli"`).
2. Do not start a capture session / control server for the install path itself.

## Context

- Stdout is user-facing: must end with `\n`.
- Exact wording may vary; required tokens: absolute path, chrome://extensions,
  load unpacked (or Load unpacked), developer (Developer mode).

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeInstallCLI
	return nil
}
```
