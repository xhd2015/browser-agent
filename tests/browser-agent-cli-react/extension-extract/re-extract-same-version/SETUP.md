# Scenario

**Feature**: re-extract same version is idempotent (E2)

```
ExtractEmbeddedExtension twice
  -> same installPath and version
```

## Preconditions

- ExtractOp = re; ExtractPasses = 2.

## Steps

1. Set ExtractOpRe and ExtractPasses=2.

## Context

- Path drift breaks user-facing install instructions.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ExtractOp = ExtractOpRe
	req.ExtractPasses = 2
	return nil
}
```
