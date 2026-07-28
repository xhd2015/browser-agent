# Scenario

**Feature**: SKILL.md documents about:debugging temporary add-on flow (S2)

```
browseragent/SKILL.md
  about:debugging
  Load Temporary Add-on / temporary add-on
```

## Preconditions

- Mode already skill-md from parent.

## Steps

1. Set `SkillMDProbe = SkillMDAboutDebuggingTemporaryAddon`.

## Context

- Matches install-firefox-extension CLI help / session-new firefox output.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.SkillMDProbe = SkillMDAboutDebuggingTemporaryAddon
	return nil
}
```
