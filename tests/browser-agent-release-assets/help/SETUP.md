# Scenario

**Feature**: release-assets --help documents opt-in upload and --out defaults

```
go run ./script/github/release-assets --help
  -> usage text
  -> mentions --upload
  -> --out default is temp (not required)
  -> exit 0
```

## Preconditions

- Mode is help.
- No pack side effects; no `--out` required.

## Steps

1. Set `Mode = ModeHelp`.
2. Default Args to `--help` when empty.

## Context

- Help must not require embeds, network, or `gh`.
- Leaves: `mentions-upload` (upload flag), `out-default-temp` (temp default wording).

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeHelp
	if len(req.Args) == 0 {
		req.Args = []string{"--help"}
	}
	return nil
}
```

