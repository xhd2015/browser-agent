# Scenario

**Feature**: missing `tabs` permission → validate error mentions tabs (A3)

```
manifest.json permissions without tabs
  -> ValidateExtensionManifestJSON -> error contains "tabs"
```

## Preconditions

- Sibling fixture omits only `tabs` (other required fields present).

## Steps

1. Load `manifest.json` into ManifestJSON.

## Context

- Requirement A3.

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
