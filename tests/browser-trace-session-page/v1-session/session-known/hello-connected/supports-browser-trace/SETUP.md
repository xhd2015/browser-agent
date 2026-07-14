# Scenario

**Feature**: hello satisfies capability gate (browser-trace feature + version ≥ 1.2.0)

```
# Capable extension agent
Test Client -> POST /v1/hello {
  version: "1.2.0" (or higher),
  features: ["browser-trace", ...]
}
Control Server -> extension.supports_browser_trace=true, version echoed
```

## Preconditions

- Hello includes feature string `browser-trace`.
- Hello version is semver-ish ≥ `1.2.0`.
- Connected is true; support is true.

## Steps

1. Set `HelloVersion = "1.2.0"`.
2. Set `HelloFeatures = ["browser-trace", "multi-tab-window"]`.
3. Set `HelloOmitFeatures = false`.
4. Children choose not-yet-recording vs recording status staging.

## Context

- Product default capability floor is **1.2.0**.
- Multi-tab feature may be present but is not required for supports_browser_trace.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.HelloVersion = "1.2.0"
	req.HelloFeatures = []string{"browser-trace", "multi-tab-window"}
	req.HelloOmitFeatures = false
	return nil
}
```
