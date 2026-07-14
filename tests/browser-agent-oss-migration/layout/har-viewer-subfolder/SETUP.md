# Scenario

**Feature**: HAR viewer stack lives under `har-viewer/`

```
# required har-viewer paths
Test Client -> stat har-viewer/main.go, har-viewer/server/, har-viewer/project-api-capture-react/
Test Client <- all exist
```

## Preconditions

- Option B layout: HAR viewer is a subfolder, not repo root.

## Steps

1. Set `Leaf = har-viewer-subfolder`.

## Context

- Directory paths may be files or dirs; `server/` and `project-api-capture-react/` are directories.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Leaf = LeafHarViewerSubfolder
	return nil
}
```