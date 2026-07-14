# Scenario

**Feature**: HandleCLI `session delete` against live daemon

```
RunDaemon -> POST /v1/sessions -> waiting_extension
HandleCLI session delete --session-id ID --base-dir BaseDir
  -> success or rejection (extension connected / not found)
```

## Preconditions

- Mode `cli`; daemon started per leaf unless not-found uses unknown id on live daemon.

## Steps

1. Set `Mode = cli`.
2. Leaf Setup sets `CLIOp` and extension-connect flag.

## Context

- Omit `--addr`; addr resolves from `server.json` (session-addr-resolve pattern).

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeCLI
	return nil
}
```