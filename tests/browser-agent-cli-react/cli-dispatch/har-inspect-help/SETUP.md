# Scenario

**Feature**: `har inspect --help` lists summary/paths/entries/show

```
HandleCLI(["har", "inspect", "--help"])
  -> nil error; lists inspect commands
```

## Preconditions

- ModeCLIDispatch from parent.

## Steps

1. Set CLIArgs to har inspect --help.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.CLIArgs = []string{"har", "inspect", "--help"}
	req.CLIEnv = map[string]string{}
	return nil
}
```
