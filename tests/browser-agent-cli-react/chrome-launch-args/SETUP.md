# Scenario

**Feature**: pure Chrome launch arg builder (no real browser)

```
BuildChromeArgs(sessionURL, extensionInstallPath)
  -> includes --load-extension=<path>
  -> must NOT include --user-data-dir
```

## Preconditions

- ModeChromeArgs.
- May extract first when ExtensionPath empty.
- No process launch of Chrome.

## Steps

1. Set Mode = ModeChromeArgs.
2. Default SessionURL to product session page on 43761.

## Context

- Isolated profile (--user-data-dir) is an explicit non-goal.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeChromeArgs
	if req.SessionURL == "" {
		req.SessionURL = "http://127.0.0.1:43761/go?session=test-sess"
	}
	return nil
}
```
