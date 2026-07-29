# Scenario

**Feature**: release --dry-run plans multi-platform binaries

```
go run ./script/github/release --dry-run
```

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Mode = ModeDryRun
	req.Args = []string{"--dry-run"}
	return nil
}
```
