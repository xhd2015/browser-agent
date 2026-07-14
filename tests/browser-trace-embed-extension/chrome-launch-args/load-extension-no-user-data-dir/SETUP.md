# Scenario

**Feature**: launch args include --load-extension and omit --user-data-dir (requirement #4)

```
Test Client -> extract (if needed) -> extensionPath
Test Client -> BuildChromeLaunchArgs(sessionURL, extensionPath)
Args include --load-extension=<extensionPath>
Args omit --user-data-dir entirely
Args include sessionURL (new window target)
```

## Preconditions

- SessionURL set by grouping Setup.
- ExtensionPath empty → Run extracts under BaseDir.

## Steps

1. Keep Mode chrome-args; no further flags.

## Context

- Assert is pure string/slice inspection — no OS process start.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeChromeArgs
	// Empty ExtensionPath → Run extracts first so --load-extension uses real path.
	req.ExtensionPath = ""
	if req.SessionURL == "" {
		req.SessionURL = "http://127.0.0.1:43759/go?session=launch-args-test"
	}
	return nil
}
```
