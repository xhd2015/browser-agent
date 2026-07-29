# Scenario

**Feature**: install.sh at module root

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Mode = ModeInstallSH
	return nil
}
```
