# Scenario

**Feature**: Operator skill doc `cmd/browser-agent/SKILL.md`

```
Read SKILL.md from module tree
  -> workflow: serve | session new → session side commands
  -> serve --status probe
```

## Preconditions

- Mode `ModeSkillMD`.
- Leaf sets `SkillMDProbe`.

## Steps

1. Set `Mode = ModeSkillMD`.

## Context

- Package embed may mirror `cmd/browser-agent/SKILL.md`; harness reads file path.

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