# Scenario

**Feature**: SKILL.md operator workflow (serve | session new → session cmds)

```
Read cmd/browser-agent/SKILL.md
  -> session new + serve --status + deprecation for serve --session-id
```

## Preconditions

- SkillMDProbe = workflow-session-new.

## Steps

1. Set SkillMDProbe.

## Context

- Skill must not recommend `serve --session-id` as the primary bootstrap path.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.SkillMDProbe = SkillMDProbeWorkflowSessionNew
	return nil
}
```