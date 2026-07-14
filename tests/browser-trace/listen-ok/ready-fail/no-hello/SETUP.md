# Scenario

**Feature**: no extension hello within ready timeout

```
# Mock Extension stays silent — never POST /v1/hello
Control Server waiting_extension
(time passes ReadyTimeout)
browser-trace -> fail: extension not connecting / timeout
```

## Preconditions

- `ExtensionScript = none` (mock does not contact the server).
- Short ready timeout from parent.

## Steps

1. Disable mock extension activity.
2. Run until ready deadline fires.

## Context

- Requirement scenario #2.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ExtensionScript = ExtNone
	req.StopMode = StopNone
	return nil
}
```
