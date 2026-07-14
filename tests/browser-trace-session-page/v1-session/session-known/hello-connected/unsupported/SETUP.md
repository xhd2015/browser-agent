# Scenario

**Feature**: extension connected but does **not** support browser-trace

```
# Agent hellos, yet capability gate fails
Test Client -> POST /v1/hello { version?, features? }  # fails support rule
Control Server -> extension.connected=true, supports_browser_trace=false
Test Client -> GET /v1/session
Control Server -> hint mentions update / missing support
```

## Preconditions

- Hello is posted (`DoHello=true` from parent).
- Combined version+features **fail** the support rule
  (missing `browser-trace` feature, and/or version &lt; 1.2.0, and/or features omitted).
- No recording status on this branch (capability message is the focus).

## Steps

1. Set `DoStatusRecording = false`.
2. Children set the concrete failing hello payload (MECE reasons).

## Context

- Connection and support are independent flags.
- Product prefers requiring explicit `browser-trace` in features; version alone is not enough.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.DoStatusRecording = false
	req.HelloOmitFeatures = false
	return nil
}
```
