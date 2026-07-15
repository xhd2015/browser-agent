# Scenario

**Feature**: SKILL.md create-tab + Target polyfill (S1)

```
Read SKILL.md
  create-tab / create_tab
  Target polyfill + tab_id
```

## Preconditions

- Mode already skill-md from parent.

## Steps

1. Set SkillMDProbe = create-tab-target.

## Context

- Requirement S1.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeSkillMD
	req.SkillMDProbe = "create-tab-target"
	return nil
}
```
