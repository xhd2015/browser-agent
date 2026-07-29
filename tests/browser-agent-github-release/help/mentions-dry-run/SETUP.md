# Scenario

**Feature**: --help mentions --dry-run, credentials, install.sh

```
go run ./script/github/release --help
```

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	// parent already set ModeHelp + Args
	return nil
}
```
