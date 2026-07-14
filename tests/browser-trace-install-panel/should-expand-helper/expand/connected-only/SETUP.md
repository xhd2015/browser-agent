# Scenario

**Feature**: expand when connected but supports=false (req #4)

```
ShouldExpandInstallPanel(true, false) -> true
```

## Preconditions

- Connected=true, Supports=false.
- Matches hello without browser-trace capability.

## Steps

1. Set `Connected = true`.
2. Set `Supports = false`.

## Context

- Parity with `go-html/hello-no-supports` expand expectation.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Connected = true
	req.Supports = false
	return nil
}
```
