# Scenario

**Feature**: SKILL.md documents session new --browser firefox (S3)

```
browseragent/SKILL.md
  browser-agent session new --browser firefox
  (or session new with browser firefox recipe)
```

## Preconditions

- Mode already skill-md from parent.

## Steps

1. Set `SkillMDProbe = SkillMDSessionNewBrowserFirefox`.

## Context

- Prefer exact flag form `--browser firefox` near `session new`.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.SkillMDProbe = SkillMDSessionNewBrowserFirefox
	return nil
}
```
