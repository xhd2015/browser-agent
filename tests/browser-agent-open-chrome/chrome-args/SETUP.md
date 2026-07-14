# Scenario

**Feature**: BuildManagedChromeArgs pure argv builder

```
BuildManagedChromeArgs(dataDir, extPath, url)
  --user-data-dir + --load-extension + --new-window [+ url]
```

## Preconditions

- ModeChromeArgs.
- Ensures extension exists via EnsureManagedExtension first in Run.

## Steps

1. Set Mode = ModeChromeArgs.

## Context

- No Chrome process launch.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeChromeArgs
	return nil
}
```
