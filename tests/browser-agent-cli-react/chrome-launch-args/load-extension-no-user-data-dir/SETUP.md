# Scenario

**Feature**: --load-extension without --user-data-dir (F1)

```
BuildChromeArgs(url, extractedPath)
  --load-extension=<path> present
  --user-data-dir absent
  session URL present
```

## Preconditions

- ExtensionPath empty → Run extracts first.

## Steps

1. Clear ExtensionPath override.
2. Keep default SessionURL from parent.

## Context

- Pure argv only.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ExtensionPath = ""
	return nil
}
```
