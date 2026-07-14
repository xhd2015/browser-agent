# Scenario

**Feature**: InstallChromeExtension stdout help (E4)

```
InstallChromeExtension(stdout, BaseDir)
  -> absolute path
  -> chrome://extensions
  -> Load unpacked / Developer
  -> trailing \n
```

## Preconditions

- ExtractOp = install-cli.

## Steps

1. Set ExtractOpInstallCLI.

## Context

- CLI flag wrapper is thin; package API is the GREEN path.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ExtractOp = ExtractOpInstallCLI
	return nil
}
```
