# Scenario

**Feature**: /go install panel present but collapsed when connected+supports (req #2)

```
# Capable extension agent
Test Client -> POST /v1/hello {
  version: "1.2.0",
  features: ["browser-trace"]
}
Test Client -> GET /go?session=<id>
Control Server -> panel STILL in DOM, default collapsed
  (open absent or data-default-open=false; summary/markers remain)
```

## Preconditions

- Hello with version ≥ 1.2.0 and feature `browser-trace`.
- Expand policy: `ShouldExpandInstallPanel(true, true) == false`.
- Old product removed panel or set display:none — both are failures now.

## Steps

1. Set `DoHello = true`.
2. Set `HelloVersion = "1.2.0"`.
3. Set `HelloFeatures = ["browser-trace"]`.

## Context

- Recording vs not-yet-recording does not change serve-time expand when
  connected+supports already holds; one hello-only leaf is enough.
- User can still expand manually in the browser (not asserted here).

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.DoHello = true
	req.HelloVersion = "1.2.0"
	req.HelloFeatures = []string{"browser-trace"}
	return nil
}
```
