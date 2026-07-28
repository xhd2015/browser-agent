# Scenario

**Feature**: manifest grants scripting and/or injection-ready host permissions

```
manifest.json
  permissions: […, "tabs", "scripting"? …]
  host_permissions: broader than localhost-only when page inject/capture needed
```

## Preconditions

- ManifestTarget = scripting-and-host-permissions.

## Steps

1. Set `ManifestTarget = ManifestScriptingAndHosts`.

## Context

- Accept either:
  - `"scripting"` in permissions (MV3 scripting.executeScript path), **or**
  - clear tabs.executeScript path with sufficient host_permissions
- Hosts: localhost-only is insufficient for arbitrary page jobs; need
  `http://*/*` / `https://*/*` / `<all_urls>` (or equivalent broad patterns).

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.ManifestTarget = ManifestScriptingAndHosts
	return nil
}
```
