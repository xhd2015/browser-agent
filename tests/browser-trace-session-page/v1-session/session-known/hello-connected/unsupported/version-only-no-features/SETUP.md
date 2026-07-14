# Scenario

**Feature**: hello body has only `version` (no features) — support false (requirement #3c)

```
# Backward-compat hello shape without features array
Test Client -> POST /v1/hello { "version": "1.2.0" }   # features key omitted
Control Server -> supports_browser_trace=false
# Documents product preference: require browser-trace in features; version alone is not enough
```

## Preconditions

- Hello is posted with a capable-looking version (`1.2.0`).
- Features key is **omitted** entirely (`HelloOmitFeatures = true`).

## Steps

1. Set `HelloVersion = "1.2.0"`.
2. Set `HelloOmitFeatures = true`.
3. Clear `HelloFeatures` (unused when omit is true).

## Context

- Existing lifecycle mocks historically send version-only hello; session-page
  capability must not treat that as supports_browser_trace=true.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.HelloVersion = "1.2.0"
	req.HelloFeatures = nil
	req.HelloOmitFeatures = true
	return nil
}
```
