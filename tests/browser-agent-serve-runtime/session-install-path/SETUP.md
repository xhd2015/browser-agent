# Scenario

**Feature**: GET /v1/session exposes extension_install_path after serve extract (F1)

```
Test Client -> Run(NoOpenChrome, NoAgentRun) -> extract + listen
Test Client -> GET /v1/session?session=... -> extension_install_path non-empty
```

## Preconditions

- Mode is `session-install-path`.
- Launch hooks off.

## Steps

1. Set `Mode = ModeSessionInstallPath`.
2. Child leaf probes live session JSON.

## Context

- Requirement F1 (optional thin smoke). Prefer top-level field; nested meta also OK.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeSessionInstallPath
	req.NoOpenChrome = true
	req.NoAgentRun = true
	return nil
}
```
