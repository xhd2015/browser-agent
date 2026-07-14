# Scenario

**Feature**: `origin` remote URL is the OSS repo

```
# exact remote URL
Test Client -> git remote get-url origin (cwd=RepoRoot)
Test Client <- https://github.com/xhd2015/browser-agent
```

## Preconditions

- Locked decision #6.

## Steps

1. Set `Leaf = remote-origin`.

## Context

- Exact string match (no trailing slash).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Leaf = LeafRemoteOrigin
	return nil
}
```