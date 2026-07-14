# Scenario

**Feature**: extracted manifest hosts product port 43761 (E3)

```
ExtractEmbeddedExtension
  manifest.json text mentions 43761 (not only 43759)
```

## Preconditions

- ExtractOp = manifest-hosts; single extract.

## Steps

1. Set ExtractOpManifest; ExtractPasses=1.

## Context

- host_permissions / content_scripts / externally_connectable may carry the port.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ExtractOp = ExtractOpManifest
	req.ExtractPasses = 1
	return nil
}
```
