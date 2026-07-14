# Scenario

**Feature**: SessionNew opens system Chrome without managed profile flags

```
SessionNew -> LaunchFn(argv) with --new-window + session URL; NO --user-data-dir / --load-extension
```

## Preconditions

- `NoOpenChrome` false (default).

## Steps

1. Set `SessionNewOp = system-chrome-no-user-data-dir`.

## Context

- Session id `sess-ext-install-1` appears in launch URL.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.SessionNewOp = SessionNewOpSystemChromeNoUserDataDir
	return nil
}
```