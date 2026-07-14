# Scenario

**Feature**: hello version ≥ 1.2.0 but features omit `browser-trace` (requirement #3a)

```
Test Client -> POST /v1/hello {
  version: "1.2.0",
  features: ["multi-tab-window"]   # no browser-trace
}
Control Server -> supports_browser_trace=false (feature missing)
```

## Preconditions

- Version is capable (≥ 1.2.0).
- Features array is present but does **not** contain `browser-trace`.

## Steps

1. Set `HelloVersion = "1.2.0"`.
2. Set `HelloFeatures = ["multi-tab-window"]`.
3. Set `HelloOmitFeatures = false`.

## Context

- Documents that version alone (even when features are sent without browser-trace)
  does not grant support.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.HelloVersion = "1.2.0"
	req.HelloFeatures = []string{"multi-tab-window"}
	req.HelloOmitFeatures = false
	return nil
}
```
