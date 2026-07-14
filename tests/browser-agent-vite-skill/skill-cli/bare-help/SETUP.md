# Scenario

**Feature**: bare skill without flags — help or skillcmd-consistent error (C3)

```
HandleCLI(["skill"], …)
  -> brief help mentioning --list/--show/--install  (nil or non-nil err)
  OR non-nil err mentioning --show / --list / --install / --help
  -> no hang / no serve
```

## Preconditions

- SkillAction = bare.
- skillcmd empty args currently error: expected one of --show/--install/--list;
  wrapper may map empty → help. Both acceptable.

## Steps

1. Set SkillAction bare; CLIArgs = ["skill"].

## Context

- Document implementer choice; assert non-fatal + informative.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.SkillAction = SkillActionBare
	req.CLIArgs = []string{"skill"}
	return nil
}
```
