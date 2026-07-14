# Scenario

**Feature**: skill --show prints embedded SKILL.md with product markers (C2)

```
HandleCLI(["skill", "--show"], …)
  -> SKILL.md body
  -> markers: browser-agent, BROWSER_AGENT_SESSION_ID, session, eval, 43761
  -> trailing \n; nil error
```

## Preconditions

- SkillAction = show.
- SKILL.md documents nested session env, side commands, control port.

## Steps

1. Set SkillAction show; CLIArgs skill --show.

## Context

- Substring markers (not full SKILL.md golden file).
- Complete refactor: skill should document `browser-agent session …`.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.SkillAction = SkillActionShow
	req.CLIArgs = []string{"skill", "--show"}
	return nil
}
```
