# Scenario

**Feature**: manifest content_scripts target go session page host (P5)

```
manifest.json content_scripts
  -> matches loopback host (127.0.0.1 / localhost)
  -> matches /go session page path
```

## Preconditions

- Mode is `manifest`.
- ModuleRoot resolved by root Setup.

## Steps

1. Set `Mode = ModeManifest`.
2. Child sets `ManifestProbe`.

## Context

- Content script must run on `/go?session=` pages served by control plane.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeManifest
	return nil
}
```