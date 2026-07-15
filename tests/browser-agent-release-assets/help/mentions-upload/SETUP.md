# Scenario

**Feature**: --help mentions --upload flag

```
go run ./script/github/release-assets --help
  -> combined stdout/stderr contains "--upload"
  -> exit 0
  -> trailing newline
```

## Preconditions

- Parent help Setup sets ModeHelp and Args `--help`.

## Steps

1. Pin Args to `--help` (or accept `-h` product-wide; this leaf pins `--help`).

## Context

- Documents opt-in upload for operators; upload path itself is not executed.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Args = []string{"--help"}
	return nil
}
```
