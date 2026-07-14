# Scenario

**Feature**: successful install help on stdout (requirement #3)

```
Test Client -> InstallChromeExtension(stdout, BaseDir)
Installer -> exit 0
Installer -> stdout contains absolute path + chrome://extensions + Load unpacked + Developer
Installer -> stdout ends with \n
```

## Preconditions

- Fresh BaseDir; embed available.

## Steps

1. Leave defaults from ancestors (Mode already install-cli).

## Context

- After success, extension files should exist under BaseDir (extract side effect).

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	// Explicit success path: package InstallChromeExtension only (no HTTP session).
	req.Mode = ModeInstallCLI
	req.DoHello = false
	return nil
}
```
