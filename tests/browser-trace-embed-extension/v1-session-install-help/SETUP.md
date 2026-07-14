# Scenario

**Feature**: GET /v1/session exposes extension_install_path and install guidance

```
# Normal session start extracts embed so install path is known
Test Client -> browsertrace.Run (NoOpenChrome) -> Control Server
Extractor -> {BaseDir}/extension/{version}/
Test Client -> GET /v1/session?session=<id>
Control Server -> JSON: extension_install_path, embedded_version, hint, extension.*
```

## Preconditions

- Mode is v1-session HTTP probe.
- Control server runs with BaseDir; product extracts on session start.
- Short ready timeout; probe then cancel (no real extension required for not-connected).

## Steps

1. Set `Mode = ModeV1Session` (`"v1-session"`).
2. Ensure SessionSuffix set (root Setup).
3. Descendants set DoHello / capability as needed.

## Context

- Does not re-test full capability matrix (see browser-trace-session-page).
- Focus: install path fields + install-oriented hint when not ready.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeV1Session
	req.NoOpenChrome = true
	return nil
}
```
