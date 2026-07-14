# Scenario

**Feature**: git remote points at OSS GitHub origin

```
# read configured origin URL
Test Client -> git remote get-url origin
Test Client <- https://github.com/xhd2015/browser-agent
```

## Preconditions

- Implementer runs `git init` + `git remote add origin` per decision #6.
- No commit required (decision #7).

## Steps

1. Set `Category = git`.

## Context

- Fails RED on empty repo without git init.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Category = CategoryGit
	return nil
}
```