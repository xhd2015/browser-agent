# Scenario

**Feature**: browser-agent skill subcommand via HandleCLI (C)

```
HandleCLI(["skill", "--list"|"--show"|…], env, stdout, stderr)
  -> skill name | SKILL.md body | help/error
  writers capture output (no binary shell-out)
```

## Preconditions

- Mode = ModeSkill.
- Skill content embedded (cmd/browser-agent/SKILL.md or package-level).
- skillcmd Shape 1 single skill name `browser-agent`.

## Steps

1. Set Mode = ModeSkill.
2. Leaf sets SkillAction / CLIArgs.

## Context

- MaxDispatchWait short; skill must not hang on serve.
- Empty CLIEnv so ambient BROWSER_AGENT_SESSION_ID is ignored.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeSkill
	if req.CLIEnv == nil {
		req.CLIEnv = map[string]string{}
	}
	if req.MaxDispatchWait == 0 {
		req.MaxDispatchWait = 3 * time.Second
	}
	return nil
}
```
