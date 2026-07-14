# Scenario

**Feature**: `session info` human-default output with optional `--json`

```
RunDaemon -> create session [optional fake hello]
HandleCLI session info --session-id ID [--json]
```

## Preconditions

- Mode is `info`.
- Human default unless leaf sets `InfoOpJSONEnriched`.

## Steps

1. Set `Mode = ModeInfo`.
2. Leaves set `InfoOp`.

## Context

- Human sections: Session, Created, Status, Pages, Session URL, Next steps.
- `--json` emits enriched machine snapshot.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeInfo
	return nil
}
```