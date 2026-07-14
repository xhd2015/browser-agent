# Scenario

**Feature**: ready fail with no extension hello (stage no_hello)

```
# Mock Extension stays silent — never POST /v1/hello
Control Server waiting_extension (stage no_hello)
(time passes ReadyTimeout)
browser-trace -> ready timeout
```

## Preconditions

- `ExtensionScript = none`.
- Stage for messages is **no_hello** / connect language.

## Steps

1. Disable mock extension activity.
2. Descendants choose short timeout vs heartbeat window.

## Context

- Requirement #3 (rich timeout) and heartbeat subcase.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ExtensionScript = ExtNone
	req.StopMode = StopNone
	return nil
}
```
