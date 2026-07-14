# Scenario

**Feature**: Casement server API removed everywhere

```
# walk repo excluding .git/node_modules/dist
Test Client -> search for server/api/casement directory
Test Client <- none found
```

## Preconditions

- `server/api/casement/` must not exist under repo root or `har-viewer/`.

## Steps

1. Set `Leaf = no-casement-server`.

## Context

- Walk skips `.git`, `node_modules`, `dist`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Leaf = LeafNoCasementServer
	return nil
}
```