# Scenario

**Feature**: valid manifest JSON with all required perms + hosts → Validate OK (A1)

```
manifest.json: debugger, tabs, alarms, storage + 43761 hosts + <all_urls>
  -> ValidateExtensionManifestJSON -> nil
```

## Preconditions

- Sibling `manifest.json` is complete and valid MV3 JSON.

## Steps

1. Read `manifest.json` into ManifestJSON.
2. ManifestSource remains bytes (parent).

## Context

- Requirement A1.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	data := readLeafManifest(t, "manifest.json")
	mustValidJSONObject(t, data)
	req.ManifestJSON = data
	req.ManifestSource = ManifestSourceBytes
	return nil
}
```
