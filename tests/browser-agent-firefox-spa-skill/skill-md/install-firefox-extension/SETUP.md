# Scenario

**Feature**: SKILL.md documents install-firefox-extension (S1)

```
browseragent/SKILL.md
  browser-agent install-firefox-extension
```

## Preconditions

- Mode already skill-md from parent.

## Steps

1. Set `SkillMDProbe = SkillMDInstallFirefoxExtension`.

## Context

- Command name may appear bare or with `browser-agent ` prefix.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.SkillMDProbe = SkillMDInstallFirefoxExtension
	return nil
}
```
