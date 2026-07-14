# Scenario

**Feature**: skill --list prints skill name browser-agent (C1)

```
HandleCLI(["skill", "--list"], …)
  -> stdout "browser-agent\n" (name present + trailing newline)
  -> nil error
```

## Preconditions

- SkillAction = list.

## Steps

1. Set SkillAction list; CLIArgs skill --list.

## Context

- Prefer exact line `browser-agent` with trailing `\n` (extra topic lines OK if multi-topic later).

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.SkillAction = SkillActionList
	req.CLIArgs = []string{"skill", "--list"}
	return nil
}
```
