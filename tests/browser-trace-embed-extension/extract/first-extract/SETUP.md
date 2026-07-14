# Scenario

**Feature**: first (cold) extract of embedded extension (requirement #1)

```
# BaseDir has no extension/ yet
Test Client -> ExtractEmbeddedExtension(BaseDir)  [once]
Extractor -> create {BaseDir}/extension/{version}/
Extractor -> return absolute installPath + version
```

## Preconditions

- BaseDir is fresh (root Setup temp dir).
- Single extract pass.

## Steps

1. Set `ExtractPasses = 1`.

## Context

- Directory must contain `manifest.json`; version must be non-empty and match manifest.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ExtractPasses = 1
	return nil
}
```
