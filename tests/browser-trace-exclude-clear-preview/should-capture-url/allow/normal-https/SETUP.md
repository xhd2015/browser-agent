# Scenario

**Feature**: allow normal HTTPS API URL (req #2)

```
ShouldCaptureURL("https://api.example.com/v1/users?limit=10") -> true
```

## Preconditions

- Typical application HTTPS traffic (primary happy-path allow case).

## Steps

1. Set `CaptureURL = "https://api.example.com/v1/users?limit=10"`.

## Context

- Query strings must not affect allow decision.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.CaptureURL = "https://api.example.com/v1/users?limit=10"
	return nil
}
```
