# Scenario

**Feature**: Firefox-Ext manifest permissions for Phase 2 script injection / capture

```
Read Firefox-Ext-Browser-Agent/public/manifest.json
  -> permissions include scripting when using scripting.executeScript
  -> host_permissions allow page jobs beyond localhost-only control plane
```

## Preconditions

- Mode = manifest.
- Manifest under `Firefox-Ext-Browser-Agent` (public preferred).

## Steps

1. Set `Mode = ModeManifest`.
2. Leave `ManifestTarget` for the leaf.

## Context

- MV3 temporary add-on: host access gates executeScript and capture on user pages.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Mode = ModeManifest
	return nil
}
```
