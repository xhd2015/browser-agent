# Scenario

**Feature**: install-chrome-extension still extracts and prints Load unpacked help

```
InstallChromeExtensionWithHome(stdout, TestHome)
  -> extensions/browser-agent/ + chrome://extensions + Load unpacked
```

## Preconditions

- ChromeRegressionOp = install-chrome-still-works.
- TestHome isolates HOME for install.

## Steps

1. Set ChromeRegressionOp = ChromeRegressionOpInstallChromeStillWorks.

## Context

- Confirms Firefox install CLI did not remove or break Chrome install.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.ChromeRegressionOp = ChromeRegressionOpInstallChromeStillWorks
	return nil
}
```
