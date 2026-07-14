# Scenario

**Feature**: known session, no hello → waiting; connected false (E1)

```
No WS hello
GET /v1/session -> connected=false, supports_browser_agent=false, phase waiting*
```

## Preconditions

- DoWSHello false.

## Steps

1. Set DoWSHello false.

## Context

- Hint should guide install/enable when product provides one (soft non-empty).

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.DoWSHello = false
	return nil
}
```
