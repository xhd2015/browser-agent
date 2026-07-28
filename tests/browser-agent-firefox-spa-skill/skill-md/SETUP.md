# Scenario

**Feature**: Operator skill docs cover Firefox install + session new

```
Read browseragent/SKILL.md (or cmd/browser-agent/SKILL.md)
  -> install-firefox-extension
  -> about:debugging + temporary add-on
  -> session new --browser firefox
```

## Preconditions

- Mode `ModeSkillMD`.
- Leaf sets `SkillMDProbe`.

## Steps

1. Set `Mode = ModeSkillMD`.

## Context

- Prefer package `browseragent/SKILL.md`; cmd path accepted / combined.
- Classic TDD: Chrome-primary skill today → RED until Firefox section lands.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Mode = ModeSkillMD
	if req.ModuleRoot == "" {
		t.Fatal("ModuleRoot must be set by root Setup")
	}
	return nil
}
```
