# Scenario

**Feature**: --help documents that --out defaults to a temp dir (not required)

```
go run ./script/github/release-assets --help
  -> usage: --out default is temp dir
  -> does NOT say --out is required for pack
  -> exit 0
```

## Preconditions

- Parent help Setup sets ModeHelp and Args `--help`.

## Steps

1. Pin Args to `--help`.

## Context

- Complements `pack/default-temp-out` operator discoverability.
- Upload flag coverage remains in `help/mentions-upload`.

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
