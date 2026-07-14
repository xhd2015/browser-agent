# Scenario

**Feature**: first cold extract (E1)

```
ExtractEmbeddedExtension(BaseDir) once
  -> abs installPath under {BaseDir}/extension/{version}
  -> manifest.json present; version non-empty
```

## Preconditions

- ExtractOp = first; ExtractPasses = 1.

## Steps

1. Set ExtractOpFirst and ExtractPasses=1.

## Context

- Fresh BaseDir from root Setup.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ExtractOp = ExtractOpFirst
	req.ExtractPasses = 1
	return nil
}
```
