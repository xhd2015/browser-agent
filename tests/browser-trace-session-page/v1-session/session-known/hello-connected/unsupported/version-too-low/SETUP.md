# Scenario

**Feature**: hello has `browser-trace` feature but version &lt; 1.2.0 (requirement #3b)

```
Test Client -> POST /v1/hello {
  version: "1.1.0",
  features: ["browser-trace", "multi-tab-window"]
}
Control Server -> supports_browser_trace=false (version below floor)
```

## Preconditions

- Features include `browser-trace`.
- Version is strictly below `1.2.0` (use `1.1.0`).

## Steps

1. Set `HelloVersion = "1.1.0"`.
2. Set `HelloFeatures = ["browser-trace", "multi-tab-window"]`.
3. Set `HelloOmitFeatures = false`.

## Context

- Both gates are required: feature present is not enough without min version.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.HelloVersion = "1.1.0"
	req.HelloFeatures = []string{"browser-trace", "multi-tab-window"}
	req.HelloOmitFeatures = false
	return nil
}
```
