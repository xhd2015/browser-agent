# Scenario

**Feature**: change one file → md5 changes (B2)

```
hash1 = Compute(dir)
mutate contentScript.js
hash2 = Compute(dir)
  -> hash1 != hash2
```

## Preconditions

- Default fixture includes `contentScript.js`.

## Steps

1. Set ComputeProbe = changes-on-file-edit.
2. EditRelPath = contentScript.js; EditNewContent distinct body.

## Context

- Requirement B2.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ComputeProbe = ComputeProbeChangesOnEdit
	req.EditRelPath = "contentScript.js"
	req.EditNewContent = "// mutated for B2\nwindow.__MUTATED__ = true;\n"
	return nil
}
```
