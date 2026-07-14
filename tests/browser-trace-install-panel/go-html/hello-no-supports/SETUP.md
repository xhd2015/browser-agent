# Scenario

**Feature**: /go install panel present and expanded after hello without supports (req #3)

```
# Extension connected but not capable
Test Client -> POST /v1/hello {
  version: "1.2.0",
  features: ["multi-tab-window"]   # omit browser-trace → supports=false
}
Test Client -> GET /go?session=<id>
Control Server -> panel present + expanded
  (connected && supports is false because supports=false)
```

## Preconditions

- Hello received → connected true.
- Features omit `browser-trace` → `supports_browser_trace=false`.
- Expand policy: `ShouldExpandInstallPanel(true, false) == true`.

## Steps

1. Set `DoHello = true`.
2. Set `HelloVersion = "1.2.0"`.
3. Set `HelloFeatures = ["multi-tab-window"]` (no `browser-trace`).

## Context

- One representative !supports reason is enough for HTML expand (missing feature).
- Version-too-low and version-only are covered for capability JSON in
  `browser-trace-session-page`; they share the same expand outcome here.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.DoHello = true
	req.HelloVersion = "1.2.0"
	req.HelloFeatures = []string{"multi-tab-window"}
	return nil
}
```
