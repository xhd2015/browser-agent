# Scenario

**Feature**: re-extract same embedded version is idempotent (requirement #2)

```
# First extract creates path
Test Client -> ExtractEmbeddedExtension(BaseDir) -> path1, ver1
# Second extract must reuse stable path for same version
Test Client -> ExtractEmbeddedExtension(BaseDir) -> path2, ver2
# path2 == path1; ver2 == ver1; manifest still valid
```

## Preconditions

- Same BaseDir for both passes.
- Embedded version does not change between calls.

## Steps

1. Set `ExtractPasses = 2` so Run extracts twice.

## Context

- Implementer may overwrite files; path and version string must stay stable.
- Must not create a second sibling version dir for the same version.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ExtractPasses = 2
	return nil
}
```
