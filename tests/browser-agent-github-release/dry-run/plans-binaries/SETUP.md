# Scenario

**Feature**: dry-run prints would build: browser-agent-v*-{os}-{arch}

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	return nil
}
```
