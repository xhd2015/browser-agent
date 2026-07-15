# Scenario

**Feature**: Operator skill doc SKILL.md documents create-tab + Target polyfill

```
Read browseragent/SKILL.md (or cmd/browser-agent/SKILL.md)
  -> create-tab recipe
  -> Target polyfill / tab_id language
```

## Preconditions

- Mode `ModeSkillMD`.
- Leaf sets `SkillMDProbe`.

## Steps

1. Set `Mode = ModeSkillMD`.

## Context

- Requirement S1. Prefer package `browseragent/SKILL.md`; accept cmd path.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeSkillMD
	return nil
}
```
