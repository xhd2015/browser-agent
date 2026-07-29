# Scenario

**Feature**: release --help

```
go run ./script/github/release --help
```

## Steps

1. Set Mode = help, Args = --help.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Mode = ModeHelp
	req.Args = []string{"--help"}
	return nil
}
```
