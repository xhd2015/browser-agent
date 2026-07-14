# Scenario

**Feature**: Chrome launch arg builder — load extension, no isolated profile

```
# Best-effort launcher builds argv (tests never exec Chrome)
Test Client -> BuildChromeLaunchArgs(sessionURL, extensionPath)
Arg builder -> [..., --load-extension=<path>, ..., sessionURL]
Arg builder -> must NOT include --user-data-dir
```

## Preconditions

- Mode is chrome-args pure function test.
- Extension path comes from extract when not overridden.
- SessionURL is a realistic control-server session page URL.

## Steps

1. Set `Mode = ModeChromeArgs` (`"chrome-args"`).
2. Set a deterministic SessionURL for assertions.
3. Leave ExtensionPath empty so Run extracts first (stable real path).

## Context

- Product opens default profile in a new window — no `--user-data-dir`.
- Binary name is not part of returned args (args only).

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeChromeArgs
	if req.SessionURL == "" {
		req.SessionURL = "http://127.0.0.1:43759/go?session=launch-args-test"
	}
	return nil
}
```
