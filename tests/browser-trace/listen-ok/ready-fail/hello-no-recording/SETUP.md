# Scenario

**Feature**: hello arrives but recording never starts within ready timeout

```
# Mock Extension announces presence but never reports status recording
Mock Extension -> POST /v1/hello
Mock Extension -> POST /v1/status {state: waiting_extension}  # not recording
(time passes ReadyTimeout)
browser-trace -> fail: did not start recording
```

## Preconditions

- `ExtensionScript = hello-no-recording`.
- Short ready timeout from parent.

## Steps

1. Configure mock to hello + non-recording status heartbeats only.
2. Run until ready deadline fires.

## Context

- Requirement scenario #3.
- Distinguishes “extension online but not recording” from “extension missing”.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ExtensionScript = ExtHelloNoRecording
	req.StopMode = StopNone
	return nil
}
```
